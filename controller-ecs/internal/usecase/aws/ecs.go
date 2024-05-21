package aws

import (
	"context"
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"
	"net/http"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
)

type AWSUC struct {
}

func (c *AWSUC) GetTaskEnvironment() (map[string]string, error) {
	meta, err := metadata.Get(context.Background(), &http.Client{})
	if err != nil {
		return nil, err
	}

	// Based on the Fargate platform version, we'll have access
	// to v3 or v4 of the ECS Metadata format
	switch m := meta.(type) {
	case *metadata.TaskMetadataV3:
		logs.InfoF("%s %s:%s", m.Cluster, m.Family, m.Revision)
		for _, t := range m.Containers {
			logs.Info(t.Name)
			for _, n := range t.Networks {
				logs.InfoF("%v", n.IPv4Addresses)
			}
		}

	case *metadata.TaskMetadataV4:
		logs.InfoF("%s(%s) %s:%s", m.Cluster, m.AvailabilityZone, m.Family, m.Revision)

	}
	return nil, nil
}

func (c *AWSUC) CreateRunner() error {
	//TODO implement me
	panic("implement me")
}

func NewAWSUC() usecase.IAWSUC {
	return &AWSUC{}
}
