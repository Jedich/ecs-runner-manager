package delivery

import (
	"runner-controller-ecs/internal/domain"
	"runner-controller-ecs/internal/infrastructure/logs"
	"time"
)

const INTERVAL = 5 * time.Second

type Reconciler interface {
	Init() error
	Reconcile() error
}

func StartReconcileLoop(r Reconciler) {
	logs.Info("Initializing reconcile loop")
	err := r.Init()
	if err != nil {
		logs.Fatal(err)
	}

	reconcileErrors := make(chan error)
	defer close(reconcileErrors)

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := r.Reconcile(); err != nil {
					reconcileErrors <- err
				} else {
					logs.Info("Job successful, waiting for next interval...")
				}
			}
		}
	}()

	go func() {
		for err := range reconcileErrors {
			switch err {
			case domain.ErrNotImplemented:
				logs.Fatal(err)
			default:
				logs.Error(err)
			}
		}
	}()

	select {}
}
