package main

import (
	"database/sql"
	"log"
	"order-service/internal/app"
	"order-service/internal/repository"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5434/orderdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresOrderRepository(db)
	paymentGateway := app.NewHTTPPaymentGateway("http://localhost:8081")
	uc := usecase.NewOrderUseCase(repo, paymentGateway)
	handler := http.NewOrderHandler(uc)

	router := gin.Default()
	router.POST("/orders", handler.CreateOrder)
	router.GET("/orders/:id", handler.GetOrder)
	router.PATCH("/orders/:id/cancel", handler.CancelOrder)

	router.Run(":8080")
}
