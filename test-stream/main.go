package main

import (
	"context"
	"fmt"
	"io"
	"log"

	orderpb "github.com/doni9977/ass2go-gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := orderpb.NewOrderServiceClient(conn)

	req := &orderpb.OrderRequest{
		OrderId: "dd8f86b6-0d9a-4371-8aed-7fcb3c18b0ce",
	}

	stream, err := client.SubscribeToOrderUpdates(context.Background(), req)
	if err != nil {
		log.Fatalf("error on subscribe: %v", err)
	}

	fmt.Println("Subscribed to order updates...")

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error receiving stream: %v", err)
		}
		fmt.Printf("Order ID: %s | New Status: %s\n", res.OrderId, res.Status)
	}
}
