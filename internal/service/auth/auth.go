// Package auth contains AuthService implementation.
package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const defaultTokenLifetime = time.Hour

// authService implements jwt auth tokens signing and validation,
// and passwords hashing. Implements AuthService interface.
type authService struct {
	// secret key for tokens to be signed with
	secretKey []byte

	// token lifetime duration until expiration
	expiry time.Duration
}

func New(secretKey string, expiry time.Duration) *authService {
	if expiry <= 0 {
		expiry = defaultTokenLifetime
	}

	return &authService{
		secretKey: []byte(secretKey),
		expiry:    expiry,
	}
}

const maxPasswordLength = 72 // limited by bcrypt

// MaxPasswordLength returns max password length, which sometimes can be
// limited (like in bcrypt).
func (a *authService) MaxPasswordLength() int {
	return maxPasswordLength
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
			ExpiresAt: jwt.NewNumericDate(ts.Add(a.expiry)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(a.secretKey)
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

// PasswordHash calculates hash for password.
func (a *authService) PasswordHash(password string) (hash string, err error) {
	bytesHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Not shure if I need to hex it. I guess not, but at least it's easier to read while debugging.
	hash = hex.EncodeToString(bytesHash)

	return hash, nil
}

// CheckPasswordHash compares password and its expected hash.
func (a *authService) CheckPasswordHash(hash, password string) (err error) {
	bytesHash, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(bytesHash, []byte(password))
}
