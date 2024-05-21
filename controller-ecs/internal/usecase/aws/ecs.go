package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"
	"log"
	"net/http"
	"runner-controller-ecs/internal/usecase"
	"strings"

	"runner-controller-ecs/runner"
)

type AWSUC struct {
	credentialsUC usecase.ICredentialUC

	defaultTaskDefinition *ecs.RegisterTaskDefinitionInput
	executionRoleArn      string
	taskDefinitionArn     string

	taskMetadata *metadata.TaskMetadataV4
	accountID    string
	region       string
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

	_, err := uc.GetTaskMetadata()
	if err != nil {
		log.Fatalf("failed to get task metadata: %v", err)
	}
	return uc
}

func (c *AWSUC) GetTaskEnvironment() (map[string]string, error) {
	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(creds.AWSRegion))
	if err != nil {
		return nil, err
	}

	// Create an IAM client
	iamClient := iam.NewFromConfig(cfg)
	ecsClient := ecs.NewFromConfig(cfg)

	// Check if the IAM role exists
	roleExists, roleArn := c.checkIAMRole(ctx, iamClient)
	if !roleExists {
		roleArn = c.createIAMRole(ctx, iamClient)
	}

	// Check if the ECS task definition exists
	taskDefExists, taskDefArn := c.checkTaskDefinition(ctx, ecsClient)
	if !taskDefExists {
		taskDefArn = c.createTaskDefinition(ctx, ecsClient, roleArn)
	}

	fmt.Printf("Using IAM Role ARN: %s\n", roleArn)
	fmt.Printf("Using Task Definition ARN: %s\n", taskDefArn)

	// Run ECS task
	err = c.runTask(ctx, ecsClient)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *AWSUC) checkIAMRole(ctx context.Context, client *iam.Client) (bool, string) {
	if c.executionRoleArn != "" {
		return true, c.executionRoleArn
	}

	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(ExecutionRoleName),
	})

	if err != nil {
		var notFoundErr *iamTypes.NoSuchEntityException
		if !errors.As(err, &notFoundErr) {
			log.Fatalf("failed to check IAM role: %v", err)
		}
		return false, ""
	}

	c.executionRoleArn = fmt.Sprintf("arn:aws:iam::%s:role/%s", c.accountID, ExecutionRoleName)
	return true, c.executionRoleArn
}

func (c *AWSUC) createIAMRole(ctx context.Context, client *iam.Client) string {
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
		log.Fatalf("failed to marshal trust policy, %v", err)
	}

	createRoleOutput, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(ExecutionRoleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
	})
	if err != nil {
		log.Fatalf("failed to create role, %v", err)
	}

	attachPolicies(ctx, client, ExecutionRoleName)

	c.executionRoleArn = *createRoleOutput.Role.Arn
	return c.executionRoleArn
}

func attachPolicies(ctx context.Context, client *iam.Client, roleName string) {
	policies := []string{
		"arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
	}

	for _, policyArn := range policies {
		_, err := client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			log.Fatalf("failed to attach policy %s: %v", policyArn, err)
		}
	}
}

func (c *AWSUC) checkTaskDefinition(ctx context.Context, client *ecs.Client) (bool, string) {
	if c.taskDefinitionArn != "" {
		return true, c.taskDefinitionArn
	}

	listTaskDefOutput, err := client.ListTaskDefinitions(ctx, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(TaskDefinitionFamily),
	})
	if err != nil {
		log.Fatalf("failed to list task definitions: %v", err)
	}

	if len(listTaskDefOutput.TaskDefinitionArns) > 0 {
		c.taskDefinitionArn = listTaskDefOutput.TaskDefinitionArns[len(listTaskDefOutput.TaskDefinitionArns)-1]
		return true, c.taskDefinitionArn
	}

	return false, ""
}

func (c *AWSUC) createTaskDefinition(ctx context.Context, client *ecs.Client, roleArn string) string {
	taskDef := runner.GetDefaultTaskDefinition()
	fmt.Println(taskDef)

	taskDef.TaskRoleArn = aws.String(roleArn)

	taskDefOutput, err := client.RegisterTaskDefinition(ctx, taskDef)
	if err != nil {
		log.Fatalf("failed to register task definition, %v", err)
	}

	c.taskDefinitionArn = *taskDefOutput.TaskDefinition.TaskDefinitionArn
	return c.taskDefinitionArn
}

func (c *AWSUC) runTask(ctx context.Context, client *ecs.Client) error {
	if c.taskMetadata != nil || c.taskDefinitionArn == "" {
		return errors.New("task metadata (cluster name) or task definition not set")
	}

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(c.taskMetadata.Cluster),
		TaskDefinition: aws.String(c.taskDefinitionArn),
		Count:          aws.Int32(1),
		LaunchType:     ecsTypes.LaunchTypeFargate,
		NetworkConfiguration: &ecsTypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecsTypes.AwsVpcConfiguration{
				Subnets:        []string{"subnet-12345678"},
				AssignPublicIp: ecsTypes.AssignPublicIpEnabled,
			},
		},
	}

	runTaskOutput, err := client.RunTask(ctx, runTaskInput)
	if err != nil {
		log.Fatalf("failed to run task, %v", err)
	}

	for _, task := range runTaskOutput.Tasks {
		fmt.Printf("Started task: %v\n", *task.TaskArn)
	}

	return nil
}

func (c *AWSUC) CreateRunner() error {
	//TODO implement me
	panic("implement me")
}

func (c *AWSUC) GetTaskMetadata() (*metadata.TaskMetadataV4, error) {
	if c.taskMetadata != nil {
		return c.taskMetadata, nil
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

	fmt.Println(c.accountID, c.region, metav4.Cluster, metav4.TaskARN)

	c.taskMetadata = metav4
	return metav4, nil
}
