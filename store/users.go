package store

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"go-temporal-workflow/models"
	"go-temporal-workflow/utils"
)

type UsersStore interface {
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Get(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type usersStore struct {
	conn *mongo.Collection
}

func NewUsersStore(conn *mongo.Database) UsersStore {
	return &usersStore{conn: conn.Collection("users")}
}

func (s *usersStore) Create(ctx context.Context, user *models.User) error {
	result, err := s.conn.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	log.Println("user created: ", result)
	return nil
}

func (s *usersStore) Update(ctx context.Context, user *models.User) error {

	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"password":   user.Password,
			"balance":    user.Balance,
			"role":       user.Role,
			"updated_at": user.UpdatedAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": user.ID}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Println("user updated: ", result)
	return nil
}

func (s *usersStore) Get(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *usersStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.conn.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *usersStore) GetAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	cursor, err := s.conn.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer utils.HandleCloseContext(ctx, cursor)
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}


func (s *usersStore) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.conn.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	log.Println("user deleted: ", result)
	return nil
}