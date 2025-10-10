package main

import (
	"fmt"
	"grpcapi/internals/api/handlers"
	pb "grpcapi/proto/gen"
	"log"
	"net"
	"os"

	godotenv "github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	s := grpc.NewServer()
	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

	reflection.Register(s)
	port := os.Getenv("SERVER_PORT")
	fmt.Println("gRPC Server is running on port", port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error listening on specified port: ", err)
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatal("Failed to serve: ", err)
	}

}
