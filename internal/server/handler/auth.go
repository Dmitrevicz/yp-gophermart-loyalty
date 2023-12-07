package handler

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// authService implements jwt auth tokens signing and validation.
type authService struct {
	// secret key for tokens to be signed with
	secretKey []byte

	// token lifetime duration until expiration
	expiry time.Duration
}

func NewAuthService(secretKey string, expiry time.Duration) *authService {
	return &authService{
		secretKey: []byte(secretKey),
		expiry:    expiry,
	}
}

// Claims - jwt fields used as auth claims.
// FIXME: might move auth code somewhere else...
type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

// CreateToken creates new jwt token for user.
func (a *authService) CreateToken(userID int64) (string, error) {
	ts := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(ts),
			ExpiresAt: jwt.NewNumericDate(ts.Add(a.expiry)), // XXX: maybe should be put into config
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(a.secretKey) // TODO: get secret from config
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken parses and validates the token.
func (a *authService) ParseToken(tokenString string) (userID int64, err error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			// validate the alg is what we expect
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return a.secretKey, nil
		},
	)
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, errors.New("invalid token")
	}

	return claims.UserID, nil
}
