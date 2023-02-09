package main

import (
	"flag"
	"log"
	"net/http"

	pb "Transaction/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Balance  string `json:"balance"`
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewTransactionClient(conn)

	r := gin.Default()
	r.POST("/user/create", func(ctx *gin.Context) {
		var user User

		err := ctx.ShouldBind(&user)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		userData := &pb.User{
			Login:    user.Login,
			Password: user.Password,
			Balance:  user.Balance,
		}
		res, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			User: userData,
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{
			"user": res.User,
		})
	})
	r.POST("/transaction", func(ctx *gin.Context) {
		var user User
		err := ctx.ShouldBind(&user)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		res, err := client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
			User: &pb.User{
				Login:    user.Login,
				Password: user.Password,
				Balance:  user.Balance,
			},
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"balance": res.User,
		})
		return
	})
	r.Run(":5000")

}
