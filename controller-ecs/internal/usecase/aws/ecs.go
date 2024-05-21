package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"

	"runner-controller-ecs/internal/domain"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
	"runner-controller-ecs/runner"
)

type AWSUC struct {
	credentialsUC usecase.ICredentialUC

	defaultTaskDefinition *ecs.RegisterTaskDefinitionInput
	executionRoleArn      string
	taskDefinitionArn     string

	controllerMetadata *metadata.TaskMetadataV4
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
	uc := &AWSUC{
		credentialsUC: credentialsUC,
	}

	return uc
}

func (c *AWSUC) GetTaskMetadata() (*metadata.TaskMetadataV4, error) {
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

	arnSplit := strings.Split(metav4.TaskARN, ":")
	c.accountID = arnSplit[4]
	c.region = arnSplit[3]

	logs.InfoF("%s %s %s %s", c.accountID, c.region, metav4.Cluster, metav4.TaskARN)

	c.controllerMetadata = metav4
	return metav4, nil
}

func (c *AWSUC) CreateRunner() ([]*model.Runner, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.region))
	if err != nil {
		return nil, err
	}

	// Create an IAM client
	iamClient := iam.NewFromConfig(cfg)
	ecsClient := ecs.NewFromConfig(cfg)

	if c.subnets == nil {
		var subnets []string

		tasks, err := ecsClient.DescribeTasks(ctx, &ecs.DescribeTasksInput{
			Cluster: aws.String(c.controllerMetadata.Cluster),
			Tasks:   []string{c.controllerMetadata.TaskARN},
		})
		if err != nil {
			return nil, err
		}

		subnets = strings.Split(*tasks.Tasks[0].Attachments[0].Details[2].Value, ",")

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

	logs.InfoF("Using IAM Role ARN: %s\n", roleArn)
	logs.InfoF("Using Task Definition ARN: %s\n", taskDefArn)

	// Run ECS task
	tasks, err := c.runTask(ctx, ecsClient)
	if err != nil {
		return nil, err
	}

	runners := make([]*model.Runner, 1)
	for _, task := range tasks {
		r := &model.Runner{
			ARN: *task.TaskArn,
		}
		for _, container := range task.Containers {
			if *container.Name == ExporterContainerName {
				for _, network := range container.NetworkInterfaces {
					r.MetricsPrivateIP = *network.PrivateIpv4Address
				}
				break
			}
		}
		runners = append(runners, r)
	}

	return runners, nil
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
	taskDef := runner.GetDefaultTaskDefinition()
	fmt.Println(taskDef)

	taskDef.TaskRoleArn = aws.String(roleArn)

	taskDefOutput, err := client.RegisterTaskDefinition(ctx, taskDef)
	if err != nil {
		return "", fmt.Errorf("failed to register task definition, %v", err)
	}

	c.taskDefinitionArn = *taskDefOutput.TaskDefinition.TaskDefinitionArn
	return c.taskDefinitionArn, nil
}

func (c *AWSUC) runTask(ctx context.Context, client *ecs.Client) ([]ecsTypes.Task, error) {
	if c.controllerMetadata != nil || c.taskDefinitionArn == "" {
		return nil, errors.New("task metadata (cluster name) or task definition not set")
	}

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(c.controllerMetadata.Cluster),
		TaskDefinition: aws.String(c.taskDefinitionArn),
		Count:          aws.Int32(1),
		LaunchType:     ecsTypes.LaunchTypeFargate,
		NetworkConfiguration: &ecsTypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecsTypes.AwsVpcConfiguration{
				Subnets:        c.subnets,
				AssignPublicIp: ecsTypes.AssignPublicIpDisabled,
			},
		},
	}

	runTaskOutput, err := client.RunTask(ctx, runTaskInput)
	if err != nil {
		return nil, fmt.Errorf("failed to run task, %v", err)
	}

	logs.Info("Task started, waiting for task to be provisioned...")
	time.Sleep(5 * time.Second)

	tasks, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(c.controllerMetadata.Cluster),
		Tasks:   []string{*runTaskOutput.Tasks[0].TaskArn},
	})
	if err != nil {
		return nil, err
	}

	for _, task := range tasks.Tasks {
		fmt.Printf("Started task: %v\n", *task.TaskArn)
		for _, container := range task.Containers {
			if *container.Name == ExporterContainerName {
				for _, network := range container.NetworkInterfaces {
					logs.InfoF("%s container IP: %v\n", *container.Name, *network.PrivateIpv4Address)
				}
				break
			}
		}
	}

	return tasks.Tasks, nil
}
