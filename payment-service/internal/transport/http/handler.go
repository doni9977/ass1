// payment-service/internal/transport/http/handler.go
package http

import (
	"net/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	useCase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req paymentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	payment, err := h.useCase.ProcessPayment(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	payment, err := h.useCase.GetPaymentStatus(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, payment)
}
