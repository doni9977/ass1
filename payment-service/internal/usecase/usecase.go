// payment-service/internal/usecase/usecase.go
package usecase

import (
	"errors"
	"payment-service/internal/domain"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (u *PaymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	if err := u.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *PaymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}
