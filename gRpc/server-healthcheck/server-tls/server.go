package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	svc "service"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthz "google.golang.org/grpc/health"
	healthsvc "google.golang.org/grpc/health/grpc_health_v1"
)

type userService struct {
	svc.UnimplementedUsersServer
}

func (s *userService) GetUser(
	ctx context.Context,
	in *svc.UserGetRequest,
) (*svc.UserGetReply, error) {
	log.Printf(
		"Received request for user with Email: %s Id: %s\n",
		in.Email,
		in.Id,
	)
	components := strings.Split(in.Email, "@")
	if len(components) != 2 {
		return nil, errors.New("invalid email address")
	}
	u := svc.User{
		Id:        in.Id,
		FirstName: components[0],
		LastName:  components[1],
		Age:       36,
	}
	return &svc.UserGetReply{User: &u}, nil
}

func (s *userService) GetHelp(
	stream svc.Users_GetHelpServer,
) error {
	for {

		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("Request receieved: %s\n", request.Request)
		response := svc.UserHelpReply{
			Response: request.Request,
		}
		err = stream.Send(&response)
		if err != nil {
			return err
		}
	}
	return nil
}

func registerServices(s *grpc.Server, h *healthz.Server) {
	svc.RegisterUsersServer(s, &userService{})
	healthsvc.RegisterHealthServer(s, h)
}

func updateServiceHealth(h *healthz.Server, service string, status healthsvc.HealthCheckResponse_ServingStatus) {
	h.SetServingStatus(service, status)
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

	tlsCertFile := os.Getenv("TLS_CERT_FILE_PATH")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE_PATH")
	if len(tlsCertFile) == 0 || len(tlsKeyFile) == 0 {
		log.Fatal("TLS_CERT_FILE_PATH and TLS_KEY_FILE_PATH must both be specified")
	}

	creds, err := credentials.NewServerTLSFromFile(tlsCertFile, tlsKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(grpc.Creds(creds))
	h := healthz.NewServer()
	registerServices(s, h)
	updateServiceHealth(h, svc.Users_ServiceDesc.ServiceName, healthsvc.HealthCheckResponse_SERVING)
	log.Fatal(startServer(s, lis))
}
