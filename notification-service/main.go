package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time" // Добавили библиотеку времени

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

var processedMessages sync.Map

func main() {
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
		log.Fatalf("Failed to connect to RabbitMQ after multiple attempts: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {
			if _, loaded := processedMessages.LoadOrStore(d.MessageId, true); loaded {
				d.Ack(false)
				continue
			}

			var event PaymentEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				d.Nack(false, false)
				continue
			}

			fmt.Printf("[Notification] Sent email to %s for Order %s. Amount: $%.2f\n", event.CustomerEmail, event.OrderID, event.Amount)

			d.Ack(false)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
