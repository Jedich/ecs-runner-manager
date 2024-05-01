package reconciler

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/infrastructure/logs"
)

type Reconciler struct {
}

func NewReconciler() delivery.Reconciler {
	return &Reconciler{}
}

func (r *Reconciler) Init() error {
	logs.Info("Reconciler ready")
	return nil
}

func (r *Reconciler) Reconcile() error {
	logs.Info("Doing something...")
	return nil
}
