// payment-service/internal/repository/repository.go
package repository

import (
	"database/sql"
	"payment-service/internal/domain"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Create(payment *domain.Payment) error {
	_, err := r.db.Exec("INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1, $2, $3, $4, $5)",
		payment.ID, payment.OrderID, payment.TransactionID, payment.Amount, payment.Status)
	return err
}

func (r *PostgresPaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.db.QueryRow("SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id = $1", orderID).
		Scan(&payment.ID, &payment.OrderID, &payment.TransactionID, &payment.Amount, &payment.Status)
	return payment, err
}
