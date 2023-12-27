package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/gin-gonic/gin"
)

// Balance - получение текущего баланса счёта баллов лояльности пользователя.
//
// Route: GET /api/user/balance
func (h *handlers) Balance(c *gin.Context) {
	balance, err := h.storage.Balance().Get(readContextUserID(c))
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, balance)
}

type requestWithdraw struct {
	Order model.OrderNumber `json:"order"`
	Sum   float64           `json:"sum"`
}

// Withdraw - запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа.
//
// Route: POST /api/user/balance/withdraw
func (h *handlers) Withdraw(c *gin.Context) {
	userID := readContextUserID(c)

	var (
		err error
		req requestWithdraw
	)

	if err = json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = req.Order.Validate(); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// check if user has enough points to withdraw
	balance, err := h.storage.Balance().Get(userID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if balance.Balance < req.Sum {
		// "insufficient funds"
		c.AbortWithStatus(http.StatusPaymentRequired)
		return
	}

	err = h.storage.Balance().Withdraw(req.Sum, userID, req.Order)
	if err != nil {
		if errors.Is(err, storage.ErrNegativeBalance) {
			// "insufficient funds"
			c.AbortWithStatus(http.StatusPaymentRequired)
			return
		}

		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// Withdrawals - получение информации о выводе средств с накопительного счёта пользователем.
//
// Route: GET /api/user/withdrawals
func (h *handlers) Withdrawals(c *gin.Context) {
	history, err := h.storage.Balance().Withdrawals(readContextUserID(c))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(history) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, history)
}
