package main

import (
	pb "Transaction/proto"
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"strconv"
)

func init() {
	DatabaseConnection()
}

var DB *gorm.DB
var err error

var (
	port = flag.Int("port", 50051, "gRPC server port")
)

type server struct {
	pb.UnimplementedTransactionServer
}

type User struct {
	Login    string
	Password string
	Balance  string
}

func DatabaseConnection() {
	host := "localhost"
	port := "5432"
	dbName := "transactiondb"
	dbUser := "nikolaychernikov"
	password := "123"
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host,
		port,
		dbUser,
		dbName,
		password,
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to the database...", err)
	}
	fmt.Println("Database connection successful...")
}

func (*server) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	fmt.Println("Update balance")
	var user User
	reqUser := req.GetUser()

	DB.FirstOrInit(&user, map[string]interface{}{"login": reqUser.Login, "password": reqUser.Password})
	userB1, _ := strconv.Atoi(user.Balance)
	userB2, _ := strconv.Atoi(reqUser.Balance)
	newBalance := userB1 + userB2
	if newBalance < 0 {
		return nil, errors.New("not less then 0")
	}
	res := DB.Table("users").Where("login=? AND password=?", reqUser.Login, reqUser.Password).Updates(User{Balance: strconv.Itoa(newBalance)})
	if res.RowsAffected == 0 {
		return nil, errors.New("transaction false")
	}
	return &pb.UpdateTransactionResponse{
		User: &pb.User{
			Login:    user.Login,
			Password: user.Password,
			Balance:  user.Balance,
		},
	}, nil
}

func (*server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	fmt.Println("Create User")
	user := req.GetUser()

	userData := User{
		Login:    user.GetLogin(),
		Password: user.GetPassword(),
		Balance:  "0",
	}

	res := DB.Table("users").Create(&userData)
	if res.Error != nil {
		return nil, errors.New("user could not be created")
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Login:    user.GetLogin(),
			Password: user.GetPassword(),
			Balance:  user.GetBalance(),
		},
	}, nil
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
