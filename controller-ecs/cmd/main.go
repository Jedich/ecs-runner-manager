package main

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/delivery/reconciler"
)

func main() {
	r := reconciler.NewReconciler()

	delivery.StartReconcileLoop(r)
}
