package forms

import (
	"time"
)

type SignUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInOutput struct {
	Token string `json:"token"`
}

type UserOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Balance   float64   `json:"balance"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DepositInput struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type SubscribeInput struct {
	UserID string `json:"user_id"`
}

type SubscriptionOutput struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	Price       float64   `json:"price"`
	Canceled    bool      `json:"canceled"`
	Deleted     bool      `json:"deleted"`
	Activations int       `json:"activations"`
	ActivatedAt time.Time `json:"activated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	CanceledAt  time.Time `json:"canceled_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type CancelSubscriptionInput struct {
	UserID string `json:"user_id"`
	SubID  string `json:"sub_id"`
}
