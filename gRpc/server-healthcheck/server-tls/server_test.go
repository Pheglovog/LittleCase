package main

import (
	"context"
	"errors"
	"log"
	"net"
	svc "service"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthz "google.golang.org/grpc/health"
	healthsvc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var h *healthz.Server

func startTestGrpcServer() *bufconn.Listener {
	h = healthz.NewServer()
	l := bufconn.Listen(10)
	s := grpc.NewServer()
	registerServices(s, h)
	updateServiceHealth(
		h,
		svc.Users_ServiceDesc.ServiceName,
		healthsvc.HealthCheckResponse_SERVING,
	)
	go func() {
		log.Fatal(startServer(s, l))
	}()
	return l
}

func TestUserService(t *testing.T) {

	l := startTestGrpcServer()

	bufconnDialer := func(
		ctx context.Context, addr string,
	) (net.Conn, error) {
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
	usersClient := svc.NewUsersClient(client)
	resp, err := usersClient.GetUser(
		context.Background(),
		&svc.UserGetRequest{
			Email: "jane@doe.com",
			Id:    "foo-bar",
		},
	)

	if err != nil {
		t.Fatal(err)
	}
	if resp.User.FirstName != "jane" {
		t.Errorf(
			"Expected FirstName to be: jane, Got: %s",
			resp.User.FirstName,
		)
	}
}

func getHealthSvcClient(l *bufconn.Listener) (healthsvc.HealthClient, error) {
	bufconnDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return l.Dial()
	}

	client, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(bufconnDialer),
	)
	if err != nil {
		return nil, err
	}
	return healthsvc.NewHealthClient(client), nil
}

func TestHealthService(t *testing.T) {
	l := startTestGrpcServer()
	healthClient, err := getHealthSvcClient(l)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := healthClient.Check(context.Background(), &healthsvc.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}

	serviceHealthStatus := resp.Status.String()
	if serviceHealthStatus != "SERVING" {
		t.Fatalf("Expected health: SERVING, Got: %s", serviceHealthStatus)
	}
}

func TestHealthServiceUsers(t *testing.T) {
	l := startTestGrpcServer()
	healthClient, err := getHealthSvcClient(l)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := healthClient.Check(context.Background(), &healthsvc.HealthCheckRequest{Service: "Users"})
	if err != nil {
		t.Fatal(err)
	}
	serviceHealthStatus := resp.Status.String()
	if serviceHealthStatus != "SERVING" {
		t.Fatalf("Expected health: SERVING, Got: %s", serviceHealthStatus)
	}
}

func TestHealthServiceUnknown(t *testing.T) {

	l := startTestGrpcServer()
	healthClient, err := getHealthSvcClient(l)
	if err != nil {
		t.Fatal(err)
	}

	_, err = healthClient.Check(
		context.Background(),
		&healthsvc.HealthCheckRequest{
			Service: "Repo",
		},
	)
	if err == nil {
		t.Fatalf("Expected non-nil error, Got nil error")
	}
	expectedError := status.Errorf(
		codes.NotFound, "unknown service",
	)
	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"Expected error %v, Got; %v",
			err,
			expectedError,
		)
	}
}

func TestHealthServiceWatch(t *testing.T) {

	l := startTestGrpcServer()
	healthClient, err := getHealthSvcClient(l)
	if err != nil {
		t.Fatal(err)
	}

	clientStream, err := healthClient.Watch(
		context.Background(),
		&healthsvc.HealthCheckRequest{
			Service: "Users",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := clientStream.Recv()
	if err != nil {
		t.Fatalf("Error in Watch: %#v\n", err)
	}
	if resp.Status != healthsvc.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING, Got: %#v", resp.Status.String())
	}

	updateServiceHealth(
		h,
		"Users",
		healthsvc.HealthCheckResponse_NOT_SERVING,
	)

	resp, err = clientStream.Recv()
	if err != nil {
		t.Fatalf("Error in Watch: %#v\n", err)
	}
	if resp.Status != healthsvc.HealthCheckResponse_NOT_SERVING {
		t.Errorf("Expected NOT_SERVING, Got: %#v", resp.Status.String())
	}
}
