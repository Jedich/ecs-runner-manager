package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"runner-manager-backend/internal/runners/dto"
	"runner-manager-backend/pkg/constant"
	"time"
)

type Runner struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty"`
	Color       string                `bson:"color"`
	Name        string                `bson:"name"`
	PrivateIPv4 string                `bson:"private_ipv4"`
	Status      constant.RunnerStatus `bson:"status"`
	CreatedAt   time.Time             `bson:"created_at"`
	UpdatedAt   time.Time             `bson:"updated_at"`
}

type Metrics struct {
	Timestamp time.Time              `bson:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata"`
}

func NewRunner(data *dto.UpdateRunnerRequest) *Runner {
	return &Runner{
		Name:        data.Name,
		PrivateIPv4: data.PrivateIPv4,
		Status:      data.Status,
		CreatedAt:   time.Now(),
	}
}
