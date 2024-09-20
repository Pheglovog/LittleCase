package main

import (
	"context"
	"log"
	"net"
	users "service"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func startTestGrpcServer() (*grpc.Server, *bufconn.Listener) {
	l := bufconn.Listen(10)

	s := grpc.NewServer()
	registerServices(s)
	go func() {
		err := startServer(s, l)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return s, l
}

func Test_userService(t *testing.T) {
	s, l := startTestGrpcServer()
	defer s.GracefulStop()

	bufconnDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return l.Dial()
	}

	client, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(bufconnDialer),
	)
	if err != nil {
		t.Fatal(err)
	}

	UsersClient := users.NewUsersClient(client)
	resp, err := UsersClient.GetUser(
		context.Background(),
		&users.UserGetRequest{
			Email: "abc@google.com",
			Id:    "test",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if resp.User.FirstName != "abc" {
		t.Errorf("Expected FirstName to be: abc, Got: %s", resp.User.FirstName)
	}

}
