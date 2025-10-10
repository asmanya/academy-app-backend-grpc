package main

import (
	"embed"
	"fmt"
	"grpcapi/internals/api/handlers"
	"grpcapi/internals/api/interceptors"
	"grpcapi/pkg/utils"
	pb "grpcapi/proto/gen"
	"log"
	"net"
	"os"

	godotenv "github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating .env file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp file: %v", err)
	}

	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing to temp file: %v", err)
	}

	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// TLS ENABLED
func main() {
	loadEnvFromEmbeddedFile()
	
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")
	
	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		log.Fatalf("failed to load TLS certificates")
	}
	
	// NOT USED WHILE BENCHMARKING
	// r := interceptors.NewRateLimiter(50, time.Minute)
	// s := grpc.NewServer(grpc.ChainUnaryInterceptor(r.RateLimiterInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))
	
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))
	
	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})
	
	reflection.Register(s) // used during development phase
	port := os.Getenv("SERVER_PORT")
	
	go utils.JwtStore.CleanUpExpiredTokens()
	
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

// WITHOUT TLS

// func main() {
// 	loadEnvFromEmbeddedFile()

// 	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor))

// 	pb.RegisterExecsServiceServer(s, &handlers.Server{})
// 	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
// 	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

// 	reflection.Register(s) // used during development phase
// 	port := os.Getenv("SERVER_PORT")

// 	go utils.JwtStore.CleanUpExpiredTokens()

// 	fmt.Println("gRPC Server is running on port", port)

// 	lis, err := net.Listen("tcp", port)
// 	if err != nil {
// 		log.Fatal("Error listening on specified port: ", err)
// 	}

// 	err = s.Serve(lis)
// 	if err != nil {
// 		log.Fatal("Failed to serve: ", err)
// 	}

// }