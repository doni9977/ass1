package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"order-service/internal/repository"
	grpctransport "order-service/internal/transport/grpc"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	orderpb "github.com/doni9977/ass2go-gen/order/v1"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5434/orderdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	paymentAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentAddr == "" {
		paymentAddr = "localhost:50051"
	}

	rawClient, err := grpctransport.NewPaymentClient(paymentAddr)
	if err != nil {
		log.Fatalf("failed to connect to payment grpc: %v", err)
	}

	paymentAdapter := grpctransport.NewGRPCPaymentAdapter(rawClient)

	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, paymentAdapter)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderServer := grpctransport.NewOrderServer(uc)
	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)

	go func() {
		log.Printf("Order gRPC server (Streaming) listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	handler := httptransport.NewOrderHandler(uc)
	router := gin.Default()
	router.POST("/orders", handler.CreateOrder)
	router.GET("/orders/:id", handler.GetOrder)

	log.Println("Order HTTP server listening on :8080")
	router.Run(":8080")
}
