package subscriptions

import (
	"context"
	"go.temporal.io/sdk/temporal"
)

type Activities struct {
	svc Service
}

func (a *Activities) Charge(ctx context.Context, state SubscriptionState) (SubscriptionState, error) {
	state, err := a.svc.Charge(ctx, state)
	if err != nil && err.Error() == ErrInsufficientFunds.Error() {
		return state, temporal.NewNonRetryableApplicationError(err.Error(), "user_poor", err, nil)
	}
	return state, err
}

func (a *Activities) Delete(ctx context.Context, state SubscriptionState) (SubscriptionState, error) {
	return a.svc.Delete(ctx, state)
}