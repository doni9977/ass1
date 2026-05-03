package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"order-service/internal/app"
	"order-service/internal/repository"
	grpcTransport "order-service/internal/transport/grpc"
	httpHandler "order-service/internal/transport/http"
	"order-service/internal/usecase"

	pbOrder "github.com/doni9977/ass2go-gen/order/v1"
	pbPayment "github.com/doni9977/ass2go-gen/payment/v1"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@order-db:5432/order_db?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	paymentServiceURL := os.Getenv("PAYMENT_GRPC_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "payment-service:50051"
	}
	conn, err := grpc.Dial(paymentServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	paymentClient := pbPayment.NewPaymentServiceClient(conn)
	paymentGateway := app.NewGRPCPaymentGateway(paymentClient)

	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, paymentGateway)

	restHandler := httpHandler.NewOrderHandler(uc)
	router := gin.Default()
	router.POST("/orders", restHandler.CreateOrder)
	router.GET("/orders/:id", restHandler.GetOrder)
	router.PATCH("/orders/:id/cancel", restHandler.CancelOrder)
	router.GET("/orders", restHandler.GetOrdersByAmountRange)
	router.GET("/payments", restHandler.ListPayments)

	go func() {
		restPort := os.Getenv("REST_PORT")
		if restPort == "" {
			restPort = ":8080"
		}
		router.Run(restPort)
	}()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	orderServer := grpcTransport.NewOrderServer(uc)
	pbOrder.RegisterOrderServiceServer(grpcServer, orderServer)
	grpcServer.Serve(lis)
}
