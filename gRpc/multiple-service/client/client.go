package main

import (
	"context"
	"fmt"
	"log"
	"os"
	users "service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func setupGrpcConnnection(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func getUserServiceClient(conn *grpc.ClientConn) users.UsersClient {
	return users.NewUsersClient(conn)
}

func getUser(client users.UsersClient, u *users.UserGetRequest) (*users.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func createUserRequest(jsonQuery string) (*users.UserGetRequest, error) {
	u := users.UserGetRequest{}
	input := []byte(jsonQuery)
	return &u, protojson.Unmarshal(input, &u)
}

func getUserResponseJson(resp *users.UserGetReply) ([]byte, error) {
	return protojson.Marshal(resp)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Must specify a gRPC server address and search query")
	}

	conn, err := setupGrpcConnnection(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	u, err := createUserRequest(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	client := getUserServiceClient(conn)

	resp, err := getUser(client, u)
	if err != nil {
		log.Fatal()
	}
	data, err := getUserResponseJson(resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, string(data))
}
