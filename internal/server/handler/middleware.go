package handler

import (
	"net/http"
	"strings"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/gin-gonic/gin"
)

// GetToken - retrieve auth token from header.
func GetToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	return tokenString
}

type middlewares struct {
	cfg  *config.Config
	auth *authService
}

func NewMiddlewares(cfg *config.Config, auth *authService) *middlewares {
	return &middlewares{
		cfg:  cfg,
		auth: auth,
	}
}

func (m *middlewares) CheckAuth(exclude ...string) gin.HandlerFunc {
	// Build a set of excluded paths to later be checked on.
	// Race conditions must not be the case since I initialize the map only once
	// and then leave it in closure for reads only.
	excludedPaths := make(map[string]struct{}, len(exclude))
	for _, s := range exclude {
		excludedPaths[s] = struct{}{}
	}

	return func(c *gin.Context) {
		// Check if url is in exclusion list.
		// Trailing slash must not break the code as long as
		// gin has RedirectTrailingSlash set to true.
		if _, ok := excludedPaths[c.Request.URL.Path]; ok {
			// skip authentication if url is in the list
			return
		}

		authToken := GetToken(c.Request)
		if authToken == "" {
			// "empty auth token" or "missing auth token"
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := m.auth.ParseToken(authToken)
		if err != nil {
			// "bad/invalid/expired/wrong token"
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		setContextUserID(c, userID)
	}
}
