package usecase

import (
	"errors"
	"order-service/internal/domain"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo           domain.OrderRepository
	paymentGateway domain.PaymentGateway
}

func NewOrderUseCase(repo domain.OrderRepository, pg domain.PaymentGateway) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentGateway: pg}
}

func (u *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	if idempotencyKey != "" {
		existingOrder, err := u.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil && existingOrder != nil {
			return existingOrder, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}

	if err := u.repo.Create(order); err != nil {
		return nil, err
	}

	status, err := u.paymentGateway.AuthorizePayment(order.ID, order.Amount)
	if err != nil {
		u.repo.UpdateStatus(order.ID, "Failed")
		return nil, err
	}

	if status == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	u.repo.UpdateStatus(order.ID, order.Status)

	return order, nil
}

func (u *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return u.repo.GetByID(id)
}

func (u *OrderUseCase) CancelOrder(id string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status == "Paid" {
		return errors.New("cannot cancel paid order")
	}
	return u.repo.UpdateStatus(id, "Cancelled")
}

func (u *OrderUseCase) GetOrdersByAmountRange(minAmountStr, maxAmountStr string) ([]*domain.Order, error) {
	if minAmountStr == "" && maxAmountStr == "" {
		return nil, errors.New("missing min_amount and max_amount")
	}

	if minAmountStr == "" || maxAmountStr == "" {
		return nil, errors.New("both min_amount and max_amount are required")
	}

	minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid min_amount")
	}

	maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid max_amount")
	}

	if minAmount < 0 {
		return nil, errors.New("min_amount less than 0")
	}

	if maxAmount > 1000000 {
		return nil, errors.New("max_amount greater than 1000000")
	}

	return u.repo.GetOrdersByAmountRange(minAmount, maxAmount)
}
