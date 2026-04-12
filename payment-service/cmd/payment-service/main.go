package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"payment-service/internal/repository"
	grpctransport "payment-service/internal/transport/grpc"
	httptransport "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	pb "github.com/doni9977/ass2go-gen/payment/v1"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5435/paymentdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctransport.LoggingInterceptor),
	)
	grpcHandler := grpctransport.NewPaymentHandler(uc)
	pb.RegisterPaymentServiceServer(grpcServer, grpcHandler)

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	handler := httptransport.NewPaymentHandler(uc)

	router := gin.Default()
	router.POST("/payments", handler.ProcessPayment)
	router.GET("/payments/:order_id", handler.GetPaymentStatus)

	router.Run(":8081")
}
