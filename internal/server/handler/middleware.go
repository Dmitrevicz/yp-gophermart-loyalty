package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	auth service.AuthService
}

func NewMiddlewares(cfg *config.Config, auth service.AuthService) *middlewares {
	return &middlewares{
		cfg:  cfg,
		auth: auth,
	}
}

// CheckAuth checks if user is authorized properly. Stores user id in context on
// success. Parses and validates auth token. Paths can be skipped by using arg.
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

// LogErrors writes errors to stderr.
func (m *middlewares) LogErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// some middlewares may modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		errs := c.Errors.ByType(gin.ErrorTypePrivate)
		if len(errs) == 0 {
			return
		}

		end := time.Now()
		latency := end.Sub(start)

		fields := []zapcore.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Duration("latency", latency),
		}

		userID := readContextUserID(c)
		if userID > 0 {
			fields = append(fields, zap.Int64("user_id", userID))
		}

		errMsg := errs[0].Error()
		// if many - print all of them
		if len(errs) > 1 {
			fields = append(fields, zap.Strings("errors", c.Errors.Errors()))
		}

		// Workaround: WithOptions allows to skip, in this case, unnecessary stacktrace output
		logger.Log.WithOptions(zap.AddStacktrace(zap.DPanicLevel)).Error(errMsg, fields...)

	}
}

// Recovery returns a middleware that recovers from any panics and writes a 500
// if there was one. Uses zap logger.
func (m *middlewares) Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(logger.Log, true)
}
