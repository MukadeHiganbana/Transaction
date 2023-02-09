package main

import (
	pb "Transaction/proto"
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 50051, "gRPC server port")
)

const (
	host     = "localhost"
	psqlport = 5432
	user     = "nikolaychernikov"
	password = "123"
	dbname   = "transactiondb"
)

type server struct {
	pb.UnimplementedTransactionServer
}

func (*server) Transaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	userData := req.GetUser()

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, psqlport, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	ctx = context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, "UPDATE users SET balance = $1 + balance WHERE login = $2 AND password = $3 AND (balance + $1 >= 0)", userData.GetBalance(), userData.GetLogin(), userData.GetPassword())
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return &pb.TransactionResponse{Response: "True"}, nil
}

func (*server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	userData := req.GetUser()

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, psqlport, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2)", userData.GetLogin(), userData.GetPassword())
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return &pb.CreateUserResponse{Response: "True"}, nil
}

func main() {
	fmt.Println("gRPC server running ...")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterTransactionServer(s, &server{})

	log.Printf("Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve : %v", err)
	}
}
