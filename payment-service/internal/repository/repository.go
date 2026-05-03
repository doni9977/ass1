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

func (r *PostgresPaymentRepository) ListByStatus(status string) ([]*domain.Payment, error) {
	var rows *sql.Rows
	var err error

	if status == "" {
		rows, err = r.db.Query("SELECT id, order_id, transaction_id, amount, status FROM payments")
	} else {
		rows, err = r.db.Query("SELECT id, order_id, transaction_id, amount, status FROM payments WHERE status = $1", status)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
