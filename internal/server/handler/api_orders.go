package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/accrual"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/gin-gonic/gin"
)

// PostOrders handler func.
//
// Загрузка пользователем номера заказа для расчёта.
//
// Route: POST /api/user/orders
func (h *handlers) PostOrders(c *gin.Context) {
	userID := readContextUserID(c)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	orderNumber := model.OrderNumber(strings.TrimSpace(string(body)))
	if err = orderNumber.Validate(); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// check if order already exist
	orderFound, err := h.storage.Orders().Get(orderNumber)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// if smth found
	if orderFound != nil {
		if orderFound.UserID != userID {
			// order has already been loaded by someone else
			c.AbortWithStatus(http.StatusConflict)
			return
		} else {
			// already loaded by current user
			c.Status(http.StatusOK)
			return
		}
	}

	// store new order
	order := model.Order{
		ID:     orderNumber,
		Status: accrual.StatusOrderNew,
		UserID: userID,
	}

	_, err = h.storage.Orders().Create(order)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusAccepted)
}

// GetOrders handler func.
//
// Получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях.
//
// Route: GET /api/user/orders
func (h *handlers) GetOrders(c *gin.Context) {
	userID := readContextUserID(c)

	orders, err := h.storage.Orders().GetByUserID(userID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
