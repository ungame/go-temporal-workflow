package subscriptions

import (
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

const (
	TaskQueueName            = "SubscriptionsTaskQueue"
	QuerySubscriptionState   = "QuerySubscriptionState"
	SignalCancelSubscription = "SignalCancelSubscription"
)

type SubscriptionState struct {
	ID          string
	UserID      string
	Type        string
	Price       float64
	Canceled    bool
	Deleted     bool
	Activations int
	ActivatedAt int64
	ExpiresAt   int64
	CanceledAt  int64
	DeletedAt   int64
}

func (s *SubscriptionState) HasExpired(t time.Time) bool {
	return t.After(time.Unix(s.ExpiresAt, 0))
}

func SubscriptionWorkflow(ctx workflow.Context, state SubscriptionState, activities *Activities) (SubscriptionState, error) {

	logger := workflow.GetLogger(ctx)

	err := workflow.SetQueryHandler(ctx, QuerySubscriptionState, func() (SubscriptionState, error) {
		return state, nil
	})
	if err != nil {
		return state, err
	}

	cancelSelector := workflow.NewSelector(ctx)
	cancelCh := workflow.GetSignalChannel(ctx, SignalCancelSubscription)
	cancelSelector.AddReceive(cancelCh, func(ch workflow.ReceiveChannel, _ bool) {
		var cancelSignal bool
		ch.Receive(ctx, &cancelSignal)
		state.Canceled = cancelSignal
		state.CanceledAt = time.Now().Unix()
	})

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{ErrInsufficientFunds.Error()},
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	for {

		timeout := time.Unix(state.ExpiresAt, 0).Sub(time.Unix(state.ActivatedAt, 0))

		_, err = workflow.AwaitWithTimeout(ctx, timeout, func() bool {
			return state.Canceled
		})
		if err != nil {
			return state, err
		}

		logger.Info("subscriptions expired", "user_id", state.UserID)

		err = workflow.ExecuteActivity(ctx, activities.Charge, state).Get(ctx, &state)
		if err != nil {
			if strings.Contains(err.Error(), ErrInsufficientFunds.Error()) {
				break
			}
			return state, err
		}

		logger.Info("subscription charged", "user_id", state.UserID)

		for cancelSelector.HasPending() {
			cancelSelector.Select(ctx)
		}

		if state.Canceled {
			logger.Info("subscription canceled", "id", state.ID, "user_id", state.UserID)
			break
		}
	}

	if !state.Canceled {
		err = workflow.ExecuteActivity(ctx, activities.Delete, state).Get(ctx, &state)
		if err == nil {
			logger.Info("subscription deleted", "id", state.ID, "user_id", state.UserID)
		}
	}

	return state, err
}
