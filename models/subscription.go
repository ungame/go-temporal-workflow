package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Subscription struct {
	ID          primitive.ObjectID `bson:"_id"`
	UserID      primitive.ObjectID `bson:"user_id"`
	Type        string             `bson:"type"`
	Price       float64            `bson:"price"`
	Canceled    bool               `bson:"canceled"`
	Activations int                `bson:"activations"`
	ActivatedAt time.Time          `bson:"activated_at"`
	ExpiresAt   time.Time          `bson:"expires_at"`
	CanceledAt  time.Time          `bson:"canceled_at"`
}
