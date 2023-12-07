package handler

import "github.com/gin-gonic/gin"

const (
	ctxKeyUserID = "userID"
)

func setContextUserID(c *gin.Context, userID int64) {
	c.Set(ctxKeyUserID, userID)
}

func readContextUserID(c *gin.Context) (userID int64) {
	return c.GetInt64(ctxKeyUserID)
}
