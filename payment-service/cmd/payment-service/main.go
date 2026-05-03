package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"payment-service/internal/repository"
	grpcHandler "payment-service/internal/transport/grpc"
	httpHandler "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	pb "github.com/doni9977/ass2go-gen/payment/v1"
)

func main() {

	db, err := sql.Open("postgres", "postgres://postgres:postgres@payment-db:5432/payment_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)

	gHandler := grpcHandler.NewPaymentGRPCHandler(uc)
	hHandler := httpHandler.NewPaymentHandler(uc)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcHandler.LoggingInterceptor))
	pb.RegisterPaymentServiceServer(s, gHandler)

	router := gin.Default()
	router.POST("/payments", hHandler.ProcessPayment)
	router.GET("/payments/:order_id", hHandler.GetPaymentStatus)

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Println("Запуск REST сервера на порту :8081...")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
