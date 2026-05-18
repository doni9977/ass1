package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PaymentEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

type EmailSender interface {
	Send(email, orderID string, amount float64) error
}

type RealEmailSender struct{}

func (r *RealEmailSender) Send(email, orderID string, amount float64) error {
	fmt.Printf("[Real Sender] Successfully sent email to %s for Order %s\n", email, orderID)
	return nil
}

type MockEmailSender struct{}

func (m *MockEmailSender) Send(email, orderID string, amount float64) error {
	time.Sleep(500 * time.Millisecond)

	if rand.Float32() < 0.8 {
		return fmt.Errorf("mock network error timeout")
	}

	fmt.Printf("[Mock Sender] Successfully sent email to %s for Order %s\n", email, orderID)
	return nil
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis-cache:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	ctx := context.Background()

	var sender EmailSender
	if os.Getenv("PROVIDER_MODE") == "REAL" {
		sender = &RealEmailSender{}
		fmt.Println("[Notification] Using REAL Email Provider")
	} else {
		sender = &MockEmailSender{}
		fmt.Println("[Notification] Using MOCK Email Provider")
	}

	var conn *amqp.Connection
	var err error
	for i := 0; i < 15; i++ {
		conn, err = amqp.Dial(os.Getenv("RABBITMQ_URL"))
		if err == nil {
			fmt.Println("[Notification] Successfully connected to RabbitMQ!")
			break
		}
		fmt.Printf("[Notification] RabbitMQ not ready, retrying in 2 seconds... (Attempt %d/15)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("payment.completed", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {
			go func(d amqp.Delivery) {
				var event PaymentEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					d.Nack(false, false)
					return
				}

				idempotencyKey := "processed_notification:" + d.MessageId
				isNew, err := rdb.SetNX(ctx, idempotencyKey, true, 24*time.Hour).Result()

				if err != nil || !isNew {
					fmt.Printf("[Notification] Message %s already processed. Skipping.\n", d.MessageId)
					d.Ack(false)
					return
				}

				success := false
				backoff := 2 * time.Second
				maxRetries := 5

				for i := 1; i <= maxRetries; i++ {
					err := sender.Send(event.CustomerEmail, event.OrderID, event.Amount)
					if err == nil {
						success = true
						break
					}

					fmt.Printf("[Notification] Attempt %d failed: %v. Retrying in %v...\n", i, err, backoff)
					time.Sleep(backoff)
					backoff *= 2
				}

				if success {
					d.Ack(false)
				} else {
					fmt.Printf("[Notification] Failed to process message %s after %d retries.\n", d.MessageId, maxRetries)
					rdb.Del(ctx, idempotencyKey)
					d.Nack(false, true)
				}
			}(d)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Notification service shutting down...")
}
