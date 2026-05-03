// payment-service/internal/domain/payment.go
package domain

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}

type PaymentRepository interface {
	Create(payment *Payment) error
	GetByOrderID(orderID string) (*Payment, error)
	ListByStatus(status string) ([]*Payment, error)
}
