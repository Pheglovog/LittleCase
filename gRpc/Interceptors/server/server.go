package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	users "service"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	users.UnimplementedUsersServer
}

func (s *userService) GetUser(ctx context.Context, in *users.UserGetRequest) (*users.UserGetReply, error) {
	log.Printf("Received request for user with Email: %s Id: %s\n", in.Email, in.Id)
	components := strings.Split(in.Email, "@")
	if len(components) != 2 {
		return nil, status.Error(codes.InvalidArgument, "Invalid email address specified")
	}
	if components[0] == "panic" {
		panic("I was asked to panic")
	}
	u := users.User{
		Id:        in.Id,
		FirstName: components[0],
		LastName:  components[1],
		Age:       36,
	}
	return &users.UserGetReply{User: &u}, nil
}

func (s *userService) GetHelp(stream users.Users_GetHelpServer) error {
	log.Println("Client connected")
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fmt.Printf("Request received: %s\n", request.Request)
		if request.Request == "panic" {
			panic("I was asked to panic")
		}
		response := users.UserHelpReply{
			Response: request.Request,
		}
		err = stream.Send(&response)
		if err != nil {
			return err
		}
	}
	log.Println("Client disconnected")
	return nil
}

func registerServices(s *grpc.Server) {
	users.RegisterUsersServer(s, &userService{})
}

func startServer(s *grpc.Server, l net.Listener) error {
	return s.Serve(l)
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":50051"
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metricUnaryInterceptor,
			loggingUnaryInterceptor,
			panicUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			metricStreamInterceptor,
			loggingStreamInterceptor,
			panicStreamInterceptor,
		),
	)
	registerServices(s)
	log.Fatal(startServer(s, lis))
}
