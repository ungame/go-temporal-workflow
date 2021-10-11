package subscriptions

import (
	"go-temporal-workflow/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const defaultExpiration = time.Second * 30

func NewDefaultSubscriptionState(id primitive.ObjectID, user *models.User) SubscriptionState {
	activatedAt := time.Now()

	return SubscriptionState{
		ID:          id.Hex(),
		UserID:      user.ID.Hex(),
		Type:        "DEFAULT",
		Price:       50.0,
		Activations: 0,
		ActivatedAt: activatedAt.Unix(),
		ExpiresAt:   activatedAt.Add(defaultExpiration).Unix(),
	}
}
