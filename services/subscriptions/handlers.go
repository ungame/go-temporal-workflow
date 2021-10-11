package subscriptions

import (
	"context"
	"encoding/json"
	"go-temporal-workflow/forms"
	"go-temporal-workflow/rmq"
)

type subscriptionsHandlers struct {
	svc Service
}

func NewHandler(svc Service, consumer rmq.Consumer) {
	handler := &subscriptionsHandlers{svc: svc}

	consumer.HandleFunc(forms.SubscribeInput{}, handler.HandleSubscribe)
	consumer.HandleFunc(forms.CancelSubscriptionInput{}, handler.HandleCancel)
}

func (h *subscriptionsHandlers) HandleSubscribe(data []byte) error {
	var in forms.SubscribeInput
	err := json.Unmarshal(data, &in)
	if err != nil {
		return err
	}
	return h.svc.Subscribe(context.Background(), &in)
}

func (h *subscriptionsHandlers) HandleCancel(data []byte) error {
	var in forms.CancelSubscriptionInput
	err := json.Unmarshal(data, &in)
	if err != nil {
		return err
	}
	return h.svc.Cancel(context.Background(), &in)
}

