package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"runner-manager-backend/internal/ctrls/dto"
	"runner-manager-backend/internal/runners/entities"
	"time"
)

type RunnerController struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Runners   []*entities.Runner `bson:"runners"`
}

func NewRunnerController(data *dto.CreateRunnerControllerRequest) *RunnerController {
	return &RunnerController{
		Name:      data.Name,
		Runners:   make([]*entities.Runner, 0),
		CreatedAt: time.Now(),
	}
}
