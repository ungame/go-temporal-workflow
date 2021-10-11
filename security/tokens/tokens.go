package tokens

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

const defaultExpiration = time.Minute * 60

func New(userID string) (string, error) {
	issuedAt := time.Now()

	claims := jwt.StandardClaims{
		ExpiresAt: issuedAt.Add(defaultExpiration).Unix(),
		Id:        userID,
		IssuedAt:  issuedAt.Unix(),
		Issuer:    "workflows",
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(jwtSecretKey))
}

type TokenPayload struct {
	UserID    string
	ExpiresAt time.Time
	IssuedAt  time.Time
	Issuer    string
	Subject   string
}

func getKey(token *jwt.Token) (interface{}, error) {
	_, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(jwtSecretKey), nil
}

func Parse(tokenString string) (*TokenPayload, error) {
	claims := new(jwt.StandardClaims)
	token, err := jwt.ParseWithClaims(tokenString, claims, getKey)
	if err != nil {
		return nil, err
	}
	var ok bool
	claims, ok = token.Claims.(*jwt.StandardClaims)
	if !token.Valid || !ok {
		return nil, jwt.NewValidationError("invalid token", jwt.ValidationErrorMalformed)
	}
	return &TokenPayload{
		UserID:    claims.Id,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
	}, nil
}
