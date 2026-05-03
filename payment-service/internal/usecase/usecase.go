package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"payment-service/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
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

	if status == "Authorized" {
		_ = u.PublishPaymentEvent(context.Background(), orderID, float64(amount), "user@example.com", status)
	}

	return payment, nil
}

func (u *PaymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}

func (u *PaymentUseCase) ListPayments(status string) ([]*domain.Payment, error) {
	return u.repo.ListByStatus(status)
}

func (u *PaymentUseCase) PublishPaymentEvent(ctx context.Context, orderID string, amount float64, email string, status string) error {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	event := map[string]interface{}{
		"order_id":       orderID,
		"amount":         amount,
		"customer_email": email,
		"status":         status,
	}
	body, _ := json.Marshal(event)

	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			MessageId:    fmt.Sprintf("%s-%d", orderID, time.Now().UnixNano()),
			Body:         body,
		})

	return err
}
