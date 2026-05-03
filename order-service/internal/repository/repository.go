package repository

import (
	"database/sql"
	"order-service/internal/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(order *domain.Order) error {
	_, err := r.db.Exec("INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt, order.IdempotencyKey)
	return err
}

func (r *PostgresOrderRepository) GetByID(id string) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRow("SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt, &order.IdempotencyKey)
	return order, err
}

func (r *PostgresOrderRepository) GetByIdempotencyKey(key string) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRow("SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE idempotency_key = $1", key).
		Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt, &order.IdempotencyKey)
	return order, err
}

func (r *PostgresOrderRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, id)
	return err
}
func (r *PostgresOrderRepository) GetOrdersByAmountRange(minAmount, maxAmount int64) ([]*domain.Order, error) {
	rows, err := r.db.Query("SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE amount >= $1 AND amount <= $2", minAmount, maxAmount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		var idempotencyKey sql.NullString

		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.ItemName,
			&order.Amount,
			&order.Status,
			&order.CreatedAt,
			&idempotencyKey,
		); err != nil {
			return nil, err
		}

		if idempotencyKey.Valid {
			order.IdempotencyKey = idempotencyKey.String
		}

		orders = append(orders, order)
	}
	return orders, nil
}
