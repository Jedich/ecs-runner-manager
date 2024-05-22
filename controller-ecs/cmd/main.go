package main

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/delivery/http"
	"runner-controller-ecs/internal/delivery/reconciler"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/tools"
)

func main() {
	logs.NewLogger()
	tools.CheckEnvVars()

	r := reconciler.NewReconciler()

	http.StartWebhookServer()
	delivery.StartReconcileLoop(r)
}
