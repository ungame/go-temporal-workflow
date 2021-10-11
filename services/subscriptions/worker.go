package subscriptions

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func NewWorker(temporalClient client.Client, svc Service) {
	w := worker.New(temporalClient, TaskQueueName, worker.Options{})

	w.RegisterWorkflow(SubscriptionWorkflow)
	w.RegisterActivity(&Activities{svc: svc})

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Panicln(err)
	}
}
