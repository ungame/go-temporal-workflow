package store

import (
	"context"
	"go-temporal-workflow/models"
	"go-temporal-workflow/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type SubscriptionsStore interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	Update(ctx context.Context, subscription *models.Subscription) error
	Get(ctx context.Context, id primitive.ObjectID) (*models.Subscription, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Subscription, error)
	GetAll(ctx context.Context) ([]*models.Subscription, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type subscriptionsStore struct {
	conn *mongo.Collection
}

func NewSubscriptionsStore(conn *mongo.Database) SubscriptionsStore {
	return &subscriptionsStore{conn: conn.Collection("subscriptions")}
}

func (s *subscriptionsStore) Create(ctx context.Context, subscription *models.Subscription) error {
	result, err := s.conn.InsertOne(ctx, subscription)
	if err != nil {
		return err
	}
	log.Println("subscription created: ", result)
	return nil
}

func (s *subscriptionsStore) Update(ctx context.Context, subscription *models.Subscription) error {

	update := bson.M{
		"$set": bson.M{
			"type":         subscription.Type,
			"price":        subscription.Price,
			"activations":  subscription.Activations,
			"activated_at": subscription.ActivatedAt,
			"expires_at":   subscription.ExpiresAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": subscription.ID}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Println("subscription updated: ", result)
	return nil
}

func (s *subscriptionsStore) Get(ctx context.Context, id primitive.ObjectID) (*models.Subscription, error) {
	var subscription models.Subscription
	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(&subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (s *subscriptionsStore) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Subscription, error) {
	cursor, err := s.conn.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer utils.HandleCloseContext(ctx, cursor)
	var subscriptions []*models.Subscription
	err = cursor.All(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *subscriptionsStore) GetAll(ctx context.Context) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	cursor, err := s.conn.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer utils.HandleCloseContext(ctx, cursor)
	err = cursor.All(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *subscriptionsStore) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.conn.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	log.Println("subscription deleted: ", result)
	return nil
}
