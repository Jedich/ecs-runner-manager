package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runner-controller-ecs/internal/tools"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"

	"runner-controller-ecs/internal/domain"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
	runnerFile "runner-controller-ecs/runner"
)

type AWSUC struct {
	credentialsUC usecase.ICredentialUC
	cfg           *aws.Config

	defaultTaskDefinition *ecs.RegisterTaskDefinitionInput
	executionRoleArn      string
	taskDefinitionArn     string

	controllerMetadata *metadata.TaskMetadataV4
	controllerPublicIP string
	accountID          string
	region             string
	subnets            []string
	//TODO: Support task metadata V3
}

const (
	TaskDefinitionFamily  = "github-runner-task"
	ExecutionRoleName     = "runnerTaskExecutionRole"
	ExporterContainerName = "ecs-container-exporter"
)

func NewAWSUC(credentialsUC usecase.ICredentialUC) usecase.IAWSUC {
	return &AWSUC{
		credentialsUC: credentialsUC,
	}
}

func (c *AWSUC) LoadConfig() (*aws.Config, error) {
	if c.cfg != nil {
		return c.cfg, nil
	}
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.region))
	if err != nil {
		return nil, err
	}

	c.cfg = &cfg
	return &cfg, nil
}

func (c *AWSUC) GetPublicIP() string {
	return c.controllerPublicIP
}

func (c *AWSUC) GetTaskMetadata() (*metadata.TaskMetadataV4, error) {
	ctx := context.TODO()

	if c.controllerMetadata != nil {
		return c.controllerMetadata, nil
	}

	meta, err := metadata.Get(context.Background(), &http.Client{})
	if err != nil {
		return nil, err
	}

	metav4, ok := meta.(*metadata.TaskMetadataV4)
	if !ok {
		return nil, errors.New("unsupported metadata type")
	}

	c.controllerMetadata = metav4

	cfg, err := c.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Create an IAM client
	iamClient := iam.NewFromConfig(*cfg)
	ecsClient := ecs.NewFromConfig(*cfg)

	arnSplit := strings.Split(metav4.TaskARN, ":")
	c.accountID = arnSplit[4]
	c.region = arnSplit[3]

	logs.InfoF("%s %s %s %s", c.accountID, c.region, metav4.Cluster, metav4.TaskARN)

	if c.subnets == nil || c.controllerPublicIP == "" {

		var tasks *ecs.DescribeTasksOutput
		done := false
		retries := 3
		var eniID string
		for !done && retries > 0 {

			logs.Info("Waiting for ENI to be attached to the task...")
			time.Sleep(10 * time.Second)

			tasks, err = ecsClient.DescribeTasks(ctx, &ecs.DescribeTasksInput{
				Cluster: aws.String(c.controllerMetadata.Cluster),
				Tasks:   []string{c.controllerMetadata.TaskARN},
			})
			if err != nil {
				return nil, err
			}

			// Extract the ENI ID from the task
			for _, task := range tasks.Tasks {
				for _, attachment := range task.Attachments {
					if *attachment.Type == "ElasticNetworkInterface" {
						for _, detail := range attachment.Details {
							if detail.Name != nil && *detail.Name == "networkInterfaceId" {
								eniID = *detail.Value
								done = true
								break
							}
						}
					}
				}
			}
			retries--
		}
		var subnets []string

		subnets = strings.Split(*tasks.Tasks[0].Attachments[0].Details[0].Value, ",")

		ec2Client := ec2.NewFromConfig(*cfg)

		if eniID == "" {
			return nil, errors.New("no ENI found for controller")
		}

		// Describe the ENI to get the public IP address
		enis, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{eniID},
		})
		if err != nil {
			return nil, err
		}

		// Extract the public IP address
		var publicIP string
		for _, eni := range enis.NetworkInterfaces {
			if eni.Association != nil && eni.Association.PublicIp != nil {
				publicIP = *eni.Association.PublicIp
			}
		}

		if publicIP == "" {
			return nil, fmt.Errorf("no public IP found for ENI %s", eniID)
		}

		c.controllerPublicIP = publicIP

		c.subnets = subnets
	}

	// Check if the IAM role exists
	roleArn, err := c.checkIAMRole(ctx, iamClient)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			roleArn, err = c.createIAMRole(ctx, iamClient)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Check if the ECS task definition exists
	taskDefArn, err := c.checkTaskDefinition(ctx, ecsClient)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			taskDefArn, err = c.createTaskDefinition(ctx, ecsClient, roleArn)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if forceNewTaskDef := os.Getenv("FORCE_NEW_TASKDEF"); forceNewTaskDef == "true" {
		logs.InfoF("Found task definition: %s, but creating new revision", taskDefArn)
		taskDefArn, err = c.createTaskDefinition(ctx, ecsClient, roleArn)
	}

	if err != nil {
		return nil, err
	}

	logs.InfoF("Using IAM Role ARN: %s", roleArn)
	logs.InfoF("Using Task Definition ARN: %s", taskDefArn)

	return metav4, nil
}

func (c *AWSUC) CreateRunner() (*model.Runner, error) {
	ctx := context.TODO()

	cfg, err := c.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Create an ECS client
	ecsClient := ecs.NewFromConfig(*cfg)

	// Run ECS task
	task, name, err := c.runTask(ctx, ecsClient)
	if err != nil {
		return nil, err
	}

	runner := &model.Runner{
		Name:   name,
		ARN:    *task.TaskArn,
		Status: model.RunnerStatusReady,
	}
	for _, container := range task.Containers {
		if *container.Name == ExporterContainerName {
			for _, network := range container.NetworkInterfaces {
				logs.InfoF("Runner %s, exporter PrivateIPv4: %v", runner.Name, *network.PrivateIpv4Address)
				runner.PrivateIPv4 = *network.PrivateIpv4Address
			}
			break
		}
	}

	return runner, nil
}

func (c *AWSUC) checkIAMRole(ctx context.Context, client *iam.Client) (string, error) {
	if c.executionRoleArn != "" {
		return c.executionRoleArn, nil
	}

	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(ExecutionRoleName),
	})

	if err != nil {
		var notFoundErr *iamTypes.NoSuchEntityException
		if !errors.As(err, &notFoundErr) {
			return "", err
		}
		return "", domain.ErrNotFound
	}

	c.executionRoleArn = fmt.Sprintf("arn:aws:iam::%s:role/%s", c.accountID, ExecutionRoleName)
	return c.executionRoleArn, nil
}

func (c *AWSUC) createIAMRole(ctx context.Context, client *iam.Client) (string, error) {
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"Service": "ecs-tasks.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	}
	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trust policy, %v", err)
	}

	createRoleOutput, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(ExecutionRoleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create role, %v", err)
	}

	err = c.attachPolicies(ctx, client, ExecutionRoleName)
	if err != nil {
		return "", err
	}

	c.executionRoleArn = *createRoleOutput.Role.Arn
	return c.executionRoleArn, nil
}

func (c *AWSUC) attachPolicies(ctx context.Context, client *iam.Client, roleName string) error {
	policies := []string{
		"arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
	}

	for _, policyArn := range policies {
		_, err := client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return fmt.Errorf("failed to attach policy %s: %v", policyArn, err)
		}
	}

	return nil
}

func (c *AWSUC) checkTaskDefinition(ctx context.Context, client *ecs.Client) (string, error) {
	if c.taskDefinitionArn != "" {
		return c.taskDefinitionArn, nil
	}

	listTaskDefOutput, err := client.ListTaskDefinitions(ctx, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(TaskDefinitionFamily),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list task definitions: %v", err)
	}

	if len(listTaskDefOutput.TaskDefinitionArns) > 0 {
		c.taskDefinitionArn = listTaskDefOutput.TaskDefinitionArns[len(listTaskDefOutput.TaskDefinitionArns)-1]
		return c.taskDefinitionArn, nil
	}

	return "", domain.ErrNotFound
}

func (c *AWSUC) createTaskDefinition(ctx context.Context, client *ecs.Client, roleArn string) (string, error) {
	taskDef := runnerFile.GetDefaultTaskDefinition()
	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return "", err
	}

	var container ecsTypes.ContainerDefinition
	ok := false
	for i, cont := range taskDef.ContainerDefinitions {
		if *cont.Name == "github-runner" {
			container = cont
			container.Environment = append(container.Environment, []ecsTypes.KeyValuePair{
				{
					Name:  aws.String("RUNNER_NAME"),
					Value: aws.String("linux-runner"),
				},
				{
					Name:  aws.String("GITHUB_ACTIONS_RUNNER_CONTEXT"),
					Value: aws.String(fmt.Sprintf("https://github.com/%s/%s", creds.Owner, creds.Repo)),
				},
				{
					Name:  aws.String("GITHUB_ACCESS_TOKEN"),
					Value: aws.String(creds.GithubPAT),
				},
				{
					Name:  aws.String("LABELS"),
					Value: aws.String("mark1"),
				},
			}...)
			taskDef.ContainerDefinitions[i] = container
			ok = true
			break
		}
	}

	if !ok {
		return "", errors.New("container github-runner not found in default task definition")
	}

	taskDef.TaskRoleArn = aws.String(roleArn)

	taskDefOutput, err := client.RegisterTaskDefinition(ctx, taskDef)
	if err != nil {
		return "", fmt.Errorf("failed to register task definition, %v", err)
	}

	c.taskDefinitionArn = *taskDefOutput.TaskDefinition.TaskDefinitionArn
	return c.taskDefinitionArn, nil
}

func (c *AWSUC) runTask(ctx context.Context, client *ecs.Client) (*ecsTypes.Task, string, error) {
	if c.controllerMetadata == nil || c.taskDefinitionArn == "" {
		return nil, "", errors.New("task metadata (cluster name) or task definition not set")
	}

	name := "linux-" + tools.RandString(6)

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(c.controllerMetadata.Cluster),
		TaskDefinition: aws.String(c.taskDefinitionArn),
		Count:          aws.Int32(1),
		LaunchType:     ecsTypes.LaunchTypeFargate,
		Overrides: &ecsTypes.TaskOverride{
			ContainerOverrides: []ecsTypes.ContainerOverride{
				{
					Name: aws.String("github-runner"),
					Environment: []ecsTypes.KeyValuePair{
						{
							Name:  aws.String("RUNNER_NAME"),
							Value: aws.String(name),
						},
					},
				},
			},
		},
		NetworkConfiguration: &ecsTypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecsTypes.AwsVpcConfiguration{
				Subnets:        c.subnets,
				AssignPublicIp: ecsTypes.AssignPublicIpEnabled,
			},
		},
	}

	runTaskOutput, err := client.RunTask(ctx, runTaskInput)
	if err != nil {
		return nil, "", fmt.Errorf("failed to run task, %v", err)
	}

	logs.InfoF("Task %s started, waiting for task to be provisioned...", name)
	time.Sleep(5 * time.Second)

	tasks, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(c.controllerMetadata.Cluster),
		Tasks:   []string{*runTaskOutput.Tasks[0].TaskArn},
	})
	if err != nil {
		return nil, "", err
	}

	if len(tasks.Tasks) == 0 {
		return nil, "", errors.New("no tasks found or created or may have exited early")
	}

	return &tasks.Tasks[0], name, nil
}
