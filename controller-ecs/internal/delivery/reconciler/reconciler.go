package reconciler

import (
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase/aws"
	"runner-controller-ecs/internal/usecase/credentials"
)

type Reconciler struct {
}

func NewReconciler() delivery.Reconciler {
	return &Reconciler{}
}

func (r *Reconciler) Init() error {
	creds := credentials.NewCredentialUC()

	awsUC := aws.NewAWSUC(creds)
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
