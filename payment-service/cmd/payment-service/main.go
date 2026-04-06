// payment-service/cmd/payment-service/main.go
package main

import (
	"database/sql"
	"log"
	"payment-service/internal/repository"
	"payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5435/paymentdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := http.NewPaymentHandler(uc)

	router := gin.Default()
	router.POST("/payments", handler.ProcessPayment)
	router.GET("/payments/:order_id", handler.GetPaymentStatus)

	router.Run(":8081")
}
