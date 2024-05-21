package runner

import (
	"embed"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"log"
)

//go:embed task-definition.json
var taskDefinition embed.FS

func GetDefaultTaskDefinition() *ecs.RegisterTaskDefinitionInput {
	var defn ecs.RegisterTaskDefinitionInput
	fileBytes, err := taskDefinition.ReadFile("task-definition.json")
	if err != nil {
		log.Println("cannot read task file", err.Error())
		return nil
	}
	err = json.Unmarshal(fileBytes, &defn)
	if err != nil {
		log.Println("cannot parse task file", err.Error())
		return nil
	}
	return &defn
}
