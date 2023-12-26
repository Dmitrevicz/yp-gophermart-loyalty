package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/gin-gonic/gin"
)

type requestUserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register - регистрация пользователя.
//
// Route: POST /api/user/register
func (h *handlers) Register(c *gin.Context) {
	var (
		creds requestUserLogin
		user  model.User
		token string
		err   error
	)

	// parse json body
	if err = json.NewDecoder(c.Request.Body).Decode(&creds); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// sanitize inputs
	creds.Login = strings.TrimSpace(creds.Login)
	creds.Password = strings.TrimSpace(creds.Password)

	// validate inputs
	if creds.Login == "" || creds.Password == "" {
		// login and password must not be empty
		c.AbortWithStatus(http.StatusBadRequest) // TODO: add error messages to response
		return
	}

	if len(creds.Password) > h.auth.MaxPasswordLength() || len(creds.Password) < 1 {
		// "password is too long - must be less then 72 bytes"
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// check if user with provided login is already exists
	if _, err = h.storage.Users().FindByLogin(creds.Login); err == nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	} else {
		if errors.Is(err, storage.ErrNotFound) {
			user.Login = creds.Login
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	// calculate password hash
	if user.PasswordHash, err = h.auth.PasswordHash(creds.Password); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// register new user
	if user.ID, err = h.storage.Users().Create(user); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// create new auth token
	token, err = h.auth.CreateToken(user.ID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// TODO: return created id and jwt token as json?
	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

// Login - аутентификация пользователя.
//
// Route: POST /api/user/login
func (h *handlers) Login(c *gin.Context) {
	var (
		creds requestUserLogin
		user  model.User
		token string
		err   error
	)

	// parse json body
	if err = json.NewDecoder(c.Request.Body).Decode(&creds); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// sanitize inputs
	creds.Login = strings.TrimSpace(creds.Login)
	creds.Password = strings.TrimSpace(creds.Password)

	// validate inputs
	if creds.Login == "" || creds.Password == "" {
		// login and password must not be empty
		c.AbortWithStatus(http.StatusBadRequest) // TODO: add error messages to response
		return
	}

	if len(creds.Password) > h.auth.MaxPasswordLength() || len(creds.Password) < 1 {
		// "password is too long - must be less then 72 bytes"
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// check if user with provided login is already exists
	if user, err = h.storage.Users().FindByLogin(creds.Login); err == nil {
		if pErr := h.auth.CheckPasswordHash(user.PasswordHash, creds.Password); pErr != nil {
			// "wrong login or password"
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	} else {
		if errors.Is(err, storage.ErrNotFound) {
			// again, "wrong login or password"
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	// create new auth token
	token, err = h.auth.CreateToken(user.ID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}
