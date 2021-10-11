package users

import (
	"context"
	"fmt"
	"go-temporal-workflow/forms"
	"go-temporal-workflow/models"
	"go-temporal-workflow/security/passwords"
	"go-temporal-workflow/security/tokens"
	"go-temporal-workflow/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Service interface {
	SignUp(ctx context.Context, in *forms.SignUpInput) error
	SignIn(ctx context.Context, in *forms.SignInInput) (*forms.SignInOutput, error)
	GetUser(ctx context.Context, id string) (*forms.UserOutput, error)
	Deposit(ctx context.Context, in *forms.DepositInput) (*forms.UserOutput, error)
}

type service struct {
	usersStore store.UsersStore
}

func NewService(usersStore store.UsersStore) Service {
	return &service{usersStore: usersStore}
}

func (s *service) SignIn(ctx context.Context, form *forms.SignInInput) (*forms.SignInOutput, error) {
	user, err := s.usersStore.GetByEmail(ctx, form.Email)
	if err != nil {
		return nil, err
	}
	err = passwords.OK(user.Password, form.Password)
	if err != nil {
		return nil, err
	}
	token, err := tokens.New(user.ID.Hex())
	if err != nil {
		return nil, err
	}
	return &forms.SignInOutput{Token: token}, nil
}

func (s *service) SignUp(ctx context.Context, form *forms.SignUpInput) error {
	exists, err := s.usersStore.GetByEmail(ctx, form.Email)
	if err == mongo.ErrNoDocuments {
		var user models.User
		user.ID = primitive.NewObjectID()
		user.Email = form.Email
		user.Password, err = passwords.New(form.Password)
		if err != nil {
			return err
		}
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		return s.usersStore.Create(ctx, &user)
	}
	if err == nil || exists != nil {
		return fmt.Errorf("%s already registered", form.Email)
	}
	return fmt.Errorf("signup failed: %s", form.Email)
}

func (s *service) GetUser(ctx context.Context, id string) (*forms.UserOutput, error) {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	user, err := s.usersStore.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &forms.UserOutput{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Password:  user.Password,
		Balance:   user.Balance,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *service) Deposit(ctx context.Context, form *forms.DepositInput) (*forms.UserOutput, error) {
	userID, err := primitive.ObjectIDFromHex(form.UserID)
	if err != nil {
		return nil, err
	}
	user, err := s.usersStore.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Balance += form.Amount
	user.UpdatedAt = time.Now()
	err = s.usersStore.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return &forms.UserOutput{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Password:  user.Password,
		Balance:   user.Balance,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}