package main

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/delivery/http"
	"runner-controller-ecs/internal/delivery/reconciler"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/tools"
	"runner-controller-ecs/internal/usecase/aws"
	"runner-controller-ecs/internal/usecase/broker"
	"runner-controller-ecs/internal/usecase/credentials"
)

func main() {
	logs.NewLogger()
	tools.CheckEnvVars()

	webhookRequest := broker.NewBroker[model.WorkflowJobWebhook]()
	go webhookRequest.Start()

	credentialsUC := credentials.NewCredentialUC()
	awsUC := aws.NewAWSUC(credentialsUC)

	r := reconciler.NewReconciler(awsUC, webhookRequest)

	http.StartWebhookServer(webhookRequest)
	delivery.StartReconcileLoop(r)
}
