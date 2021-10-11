package subscriptions

import (
	"context"
	"errors"
	"go-temporal-workflow/forms"
	"go-temporal-workflow/models"
	"go-temporal-workflow/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/sdk/client"
	"log"
	"time"
)

type Service interface {
	Subscribe(ctx context.Context, in *forms.SubscribeInput) error
	Charge(ctx context.Context, state SubscriptionState) (SubscriptionState, error)
	GetWorkflow(ctx context.Context, id string) (*forms.SubscriptionOutput, error)
	Cancel(ctx context.Context, in *forms.CancelSubscriptionInput) error
	Delete(ctx context.Context, state SubscriptionState) (SubscriptionState, error)
}

type service struct {
	usersStore         store.UsersStore
	subscriptionsStore store.SubscriptionsStore
	temporalClient     client.Client
}

func NewService(usersStore store.UsersStore, subscriptionsStore store.SubscriptionsStore, temporalClient client.Client) Service {
	return &service{
		usersStore:         usersStore,
		subscriptionsStore: subscriptionsStore,
		temporalClient:     temporalClient,
	}
}

func (s *service) Subscribe(ctx context.Context, in *forms.SubscribeInput) error {
	userID, err := primitive.ObjectIDFromHex(in.UserID)
	if err != nil {
		return err
	}

	user, err := s.usersStore.Get(ctx, userID)
	if err != nil {
		return err
	}

	subs, err := s.subscriptionsStore.GetByUserID(ctx, user.ID)
	if err != nil {
		return err
	}

	log.Printf("current subscriptions: %d\n", len(subs))

	for index := range subs {
		if !subs[index].Canceled {
			log.Printf("subscription already activated: %s", subs[index].ID.Hex())
			return errors.New("subscription already activated")
		}
	}

	id := primitive.NewObjectID()

	state := NewDefaultSubscriptionState(id, user)

	options := client.StartWorkflowOptions{
		ID:                 id.Hex(),
		TaskQueue:          TaskQueueName,
		WorkflowRunTimeout: time.Minute * 10,
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, options, SubscriptionWorkflow, state, &Activities{svc: s})
	if err != nil {
		return err
	}

	log.Printf("workflow started: ID=%s, RunID=%s\n", we.GetID(), we.GetRunID())

	m := &models.Subscription{
		ID:          id,
		UserID:      userID,
		Type:        state.Type,
		Price:       state.Price,
		Activations: state.Activations,
		ActivatedAt: time.Unix(state.ActivatedAt, 0),
		ExpiresAt:   time.Unix(state.ExpiresAt, 0),
	}

	return s.subscriptionsStore.Create(ctx, m)
}

func (s *service) Charge(ctx context.Context, state SubscriptionState) (SubscriptionState, error) {
	userID, err := primitive.ObjectIDFromHex(state.UserID)
	if err != nil {
		return state, err
	}

	user, err := s.usersStore.Get(ctx, userID)
	if err != nil {
		return state, err
	}

	subID, _ := primitive.ObjectIDFromHex(state.ID)

	subscription, err := s.subscriptionsStore.Get(ctx, subID)
	if err != nil {
		return state, err
	}

	if user.Balance < subscription.Price {
		log.Printf("insufficient funds: UserID=%s, Balance=%.2f, SubscriptionID=%s\n", state.UserID, user.Balance, subscription.ID.Hex())
		return state, errors.New("insufficient funds")
	}

	user.Balance -= subscription.Price
	user.UpdatedAt = time.Now()

	err = s.usersStore.Update(ctx, user)
	if err != nil {
		return state, err
	}

	subscription.Activations++
	activatedAt := time.Now()
	subscription.ActivatedAt = activatedAt
	subscription.ExpiresAt = activatedAt.Add(defaultExpiration)

	err = s.subscriptionsStore.Update(ctx, subscription)
	if err != nil {
		return state, err
	}

	state.Activations = subscription.Activations
	state.ActivatedAt = subscription.ActivatedAt.Unix()
	state.ExpiresAt = subscription.ExpiresAt.Unix()

	return state, nil
}

func (s *service) GetWorkflow(ctx context.Context, id string) (*forms.SubscriptionOutput, error) {
	subID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	res, err := s.temporalClient.QueryWorkflow(ctx, subID.Hex(), "", QuerySubscriptionState)
	if err != nil {
		return nil, err
	}

	var state SubscriptionState
	err = res.Get(&state)
	if err != nil {
		return nil, err
	}

	return &forms.SubscriptionOutput{
		ID:          state.ID,
		UserID:      state.UserID,
		Type:        state.Type,
		Price:       state.Price,
		Canceled:    state.Canceled,
		Deleted:     state.Deleted,
		Activations: state.Activations,
		ActivatedAt: time.Unix(state.ActivatedAt, 0),
		ExpiresAt:   time.Unix(state.ExpiresAt, 0),
		CanceledAt:  time.Unix(state.CanceledAt, 0),
		DeletedAt:   time.Unix(state.DeletedAt, 0),
	}, nil
}

func (s *service) Cancel(ctx context.Context, in *forms.CancelSubscriptionInput) error {
	subID, err := primitive.ObjectIDFromHex(in.SubID)
	if err != nil {
		return err
	}
	sub, err := s.subscriptionsStore.Get(ctx, subID)
	if err != nil {
		return err
	}
	if sub.UserID.Hex() != in.UserID {
		return errors.New("cannot cancel subscription")
	}
	sub.Canceled = true
	sub.CanceledAt = time.Now()
	err = s.subscriptionsStore.Update(ctx, sub)
	if err != nil {
		return err
	}
	return s.temporalClient.SignalWorkflow(ctx, subID.Hex(), "", SignalCancelSubscription, sub.Canceled)
}

func (s *service) Delete(ctx context.Context, state SubscriptionState) (SubscriptionState, error) {
	subID, err := primitive.ObjectIDFromHex(state.ID)
	if err != nil {
		return state, err
	}
	err = s.subscriptionsStore.Delete(ctx, subID)
	if err != nil {
		return state, err
	}
	state.Deleted = true
	state.DeletedAt = time.Now().Unix()
	return state, nil
}
