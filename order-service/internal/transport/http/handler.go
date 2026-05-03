package http

import (
	"net/http"
	"order-service/internal/domain"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	useCase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: uc}
}

type createOrderReq struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	idemKey := c.GetHeader("Idempotency-Key")

	order, err := h.useCase.CreateOrder(req.CustomerID, req.ItemName, req.Amount, idemKey)
	if err != nil {
		if err.Error() == "service unavailable" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.useCase.GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.CancelOrder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *OrderHandler) GetOrdersByAmountRange(c *gin.Context) {
	minAmountStr := c.Query("min_amount")
	maxAmountStr := c.Query("max_amount")

	orders, err := h.useCase.GetOrdersByAmountRange(minAmountStr, maxAmountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if orders == nil {
		orders = []*domain.Order{}
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) ListPayments(c *gin.Context) {
	reqStatus := c.Query("status")

	res, err := h.useCase.ListPayments(c.Request.Context(), reqStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payments: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
