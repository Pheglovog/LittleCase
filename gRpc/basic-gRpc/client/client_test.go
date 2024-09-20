package main

import (
	"context"
	"log"
	"net"
	users "service"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type dummyUserService struct {
	users.UnimplementedUsersServer
}

func (s *dummyUserService) GetUser(ctx context.Context, in *users.UserGetRequest) (*users.UserGetReply, error) {
	components := strings.Split(in.Email, "@")
	u := users.User{
		Id:        in.Id,
		FirstName: components[0],
		LastName:  components[1],
		Age:       36,
	}
	return &users.UserGetReply{User: &u}, nil
}

func startServer(s *grpc.Server, l net.Listener) error {
	return s.Serve(l)
}

func startTestGrpcServer() (*grpc.Server, *bufconn.Listener) {
	l := bufconn.Listen(10)
	s := grpc.NewServer()
	users.RegisterUsersServer(s, &dummyUserService{})

	go func() {
		err := startServer(s, l)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return s, l
}

func Test_getUser(t *testing.T) {
	s, l := startTestGrpcServer()
	defer s.GracefulStop()

	bufconnDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return l.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(bufconnDialer))

	if err != nil {
		t.Fatal(err)
	}

	c := getUserServiceClient(conn)
	resp, err := getUser(c, &users.UserGetRequest{Email: "abc@google.com"})
	if err != nil {
		log.Fatal(err)
	}

	if resp.User.FirstName != "abc" {
		t.Fatalf("Expected : abc, Got %s", resp.User.FirstName)
	}
}
