package main

import (
	"context"
	"log"
	"time"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/SternKater/carsharing/pkg/auth"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Specify the command. Available: signin signout refresh signup")
	}
	command := os.Args[1]

// create connect to gRPC
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := auth.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

// for test only
	switch command {
	case "signin":
		req := &auth.SignInRequest{
			UserName: 		"user_name",
			UserPassword: 	"user_pwd",
		}
		resp, err := client.SignIn(ctx, req)
		if err != nil {
			log.Fatalf("could not send event: %v", err)
		}

		log.Printf("Response:\n Success=%v\n Message=%v\n token=%v\n refresh=%v", resp.Success, resp.Message, resp.AccessToken, resp.RefreshToken)

	case "signup":
		req := &auth.SignUpRequest{
			UserName: "user_name",
			UserPwd: "user_pwd",
		}
		resp, err := client.SignUp(ctx, req)
		if err != nil {
			log.Fatalf("could not send event: %v", err)
		}

		log.Printf("Response:\n Success=%v\n Message=%v\n token=%v\n refresh=%v", resp.Success, resp.Message, resp.AccessToken, resp.RefreshToken)

	case "signout":
		req := &auth.SignOutRequest{
		}
		resp, err := client.SignOut(ctx, req)
		if err != nil {
			log.Fatalf("could not send event: %v", err)
		}

		log.Printf("Response:\n Success=%v\n Message=%v\n", resp.Success, resp.Message)

	case "refresh":
		req := &auth.AuthRefreshRequest{
			RefreshToken: "213",
		}
		resp, err := client.AuthRefresh(ctx, req)
		if err != nil {
			log.Fatalf("could not send event: %v", err)
		}

		log.Printf("Response:\n Success=%v\n Message=%v\n token=%v\n refresh=%v", resp.Success, resp.Message, resp.AccessToken, resp.RefreshToken)

		
	default:
		log.Fatalf("Unknown command")
	}

}
