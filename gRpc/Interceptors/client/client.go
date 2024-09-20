package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	users "service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupGrpcConnnection(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			loggingUnaryInterceptor,
			metadataUnaryInterceptor,
		),
		grpc.WithChainStreamInterceptor(
			loggingStreamingInterceptor,
			metadataStreamInterceptor,
		),
	)
}

func getUserServiceClient(conn *grpc.ClientConn) users.UsersClient {
	return users.NewUsersClient(conn)
}

func getUser(client users.UsersClient, u *users.UserGetRequest) (*users.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func setupChat(r io.Reader, w io.Writer, c users.UsersClient) error {
	stream, err := c.GetHelp(context.Background())
	if err != nil {
		return err
	}
	for {
		scanner := bufio.NewScanner(r)
		prompt := "Request: "
		fmt.Fprint(w, prompt)

		scanner.Scan()
		if err := scanner.Err(); err != nil {
			return err
		}
		msg := scanner.Text()
		if msg == "quit" {
			break
		}
		request := users.UserHelpRequest{
			Request: msg,
		}

		err := stream.Send(&request)
		if err != nil {
			return err
		}

		resp, err := stream.Recv()
		if err != nil {
			return err
		}
		fmt.Printf("Response: %s\n", resp.Response)
	}
	return stream.CloseSend()
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Must specify a gRPC server address")
	}

	serverAddr := os.Args[1]
	methodName := os.Args[2]

	conn, err := setupGrpcConnnection(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := getUserServiceClient(conn)

	switch methodName {
	case "GetUser":
		result, err := getUser(
			c,
			&users.UserGetRequest{Email: os.Args[3]},
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "User: %s %s\n", result.User.FirstName, result.User.LastName)
	case "GetHelp":
		err = setupChat(os.Stdin, os.Stdout, c)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unrecognized method name")
	}
}
