package reconciler

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase/aws"
)

type Reconciler struct {
}

func NewReconciler() delivery.Reconciler {
	return &Reconciler{}
}

func (r *Reconciler) Init() error {
	awsUC := aws.NewAWSUC()
	_, err := awsUC.GetTaskEnvironment()
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) Reconcile() error {
	logs.Info("Doing something...")
	return nil
}
