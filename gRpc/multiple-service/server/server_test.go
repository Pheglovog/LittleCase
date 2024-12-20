package main

import (
	"context"
	"log"
	"net"
	svc "service"
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

	UsersClient := svc.NewUsersClient(client)
	resp, err := UsersClient.GetUser(
		context.Background(),
		&svc.UserGetRequest{
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

func TestRepoService(t *testing.T) {
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

	repoClient := svc.NewRepoClient(client)
	resp, err := repoClient.GetRepos(context.Background(), &svc.RepoGetRequest{
		CreatorId: "user-123",
		Id:        "repo-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repo) != 1 {
		t.Fatalf("Expected to get back 1 repo, got back: %d repos", len(resp.Repo))
	}
	gotId := resp.Repo[0].Id
	gotOwnerId := resp.Repo[0].Owner.Id

	if gotId != "repo-123" {
		t.Errorf("Expected Repo ID to be: repo-123, Got: %s", gotId)
	}

	if gotOwnerId != "user-123" {
		t.Errorf("Expected Creator ID to be: user-123, Got: %s", gotOwnerId)
	}
}
