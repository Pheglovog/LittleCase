package cmd

import (
	"bytes"
	"context"
	"flag"
	"log"
	"net"
	svc "service"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type dummyUserService struct {
	svc.UnimplementedUsersServer
}

func (s *dummyUserService) GetUser(ctx context.Context, in *svc.UserGetRequest) (*svc.UserGetReply, error) {
	components := strings.Split(in.Email, "@")
	u := svc.User{
		Id:        in.Id,
		FirstName: components[0],
		LastName:  components[1],
		Age:       36,
	}
	return &svc.UserGetReply{User: &u}, nil
}

type dummyReposService struct {
	svc.UnimplementedRepoServer
}

func (s *dummyReposService) GetRepos(ctx context.Context, in *svc.RepoGetRequest) (*svc.RepoGetReply, error) {

	repos := []*svc.Repository{
		{
			Id:    "repo-123",
			Name:  "hsh",
			Url:   "github.com",
			Owner: &svc.User{Id: "user-123"},
		},
	}
	return &svc.RepoGetReply{Repo: repos}, nil
}

func startTestGrpcServer() (*grpc.Server, *bufconn.Listener) {
	l := bufconn.Listen(10)
	s := grpc.NewServer()
	svc.RegisterUsersServer(s, &dummyUserService{})
	svc.RegisterRepoServer(s, &dummyReposService{})
	go func() {
		err := s.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return s, l
}

func Test_callUserSvc(t *testing.T) {
	tests := []struct {
		name     string
		c        grpcConfig
		respJson string
		errMsg   string
	}{
		{
			name:     "test1",
			c:        grpcConfig{},
			respJson: "",
			errMsg:   ErrInvalidGrpcMethod.Error(),
		},
		{
			name:     "test2",
			c:        grpcConfig{method: "GetUser", request: `{"email":"john@doe.com","id":"user-123"}`},
			errMsg:   "",
			respJson: `{"user":{"id":"user-123","firstName":"john","lastName":"doe.com","age":36}}`,
		},
		{
			name:     "test3",
			c:        grpcConfig{method: "GetUser", request: "foo-bar"},
			errMsg:   "invalid value",
			respJson: "",
		},
	}

	s, l := startTestGrpcServer()
	defer s.GracefulStop()

	bufconnDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return l.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(bufconnDialer),
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			usersClient := getUserServiceClient(conn)
			respJson, err := callUserMethod(usersClient, tc.c)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if !strings.Contains(errMsg, tc.errMsg) {
				t.Fatalf("Expected error: %v, got: %v", tc.errMsg, errMsg)
			}

			sanitizedRespJson := strings.Replace(string(respJson), " ", "", -1)
			sanitizedRespJson = strings.Replace(string(sanitizedRespJson), "\n", "", -1)

			if sanitizedRespJson != tc.respJson {
				t.Fatalf("Expected result: %v Got: %v", tc.respJson, sanitizedRespJson)
			}

		})
	}
}

func TestCallReposSvc(t *testing.T) {
	tests := []struct {
		name     string
		c        grpcConfig
		respJson string
		errMsg   string
	}{
		{
			name:     "test1",
			c:        grpcConfig{},
			respJson: "",
			errMsg:   ErrInvalidGrpcMethod.Error(),
		},
		{
			name:     "test2",
			c:        grpcConfig{method: "GetRepos", request: `{"id":"1"}`},
			errMsg:   "",
			respJson: `{"repo":[{"id":"repo-123","name":"hsh","url":"github.com","owner":{"id":"user-123"}}]}`,
		},
		{
			name:     "test3",
			c:        grpcConfig{method: "GetRepos", request: "foo-bar"},
			errMsg:   "invalid value",
			respJson: "",
		},
	}
	s, l := startTestGrpcServer()
	defer s.GracefulStop()

	bufconnDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return l.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(bufconnDialer),
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoClient := getRepoServiceClient(conn)
			respJson, err := callRepoMethod(repoClient, tc.c)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if !strings.Contains(errMsg, tc.errMsg) {
				t.Fatalf("Expected error: %v, got: %v", tc.errMsg, errMsg)
			}

			sanitizedRespJson := strings.Replace(string(respJson), " ", "", -1)
			sanitizedRespJson = strings.Replace(string(sanitizedRespJson), "\n", "", -1)

			if sanitizedRespJson != tc.respJson {
				t.Fatalf("Expected result: %v Got: %v", tc.respJson, sanitizedRespJson)
			}

		})
	}
}

func TestHandleGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	s := grpc.NewServer()
	defer s.Stop()

	svc.RegisterUsersServer(s, &dummyUserService{})
	svc.RegisterRepoServer(s, &dummyReposService{})

	go func() {
		s.Serve(l)
	}()

	tests := []struct {
		name     string
		args     []string
		output   string
		errMsg   string
		respJson string
	}{
		{
			name:     "test1",
			args:     []string{"-service", "Gopher", "-method", "Hello", "-request", "{}", l.Addr().String()},
			output:   "",
			errMsg:   "unrecognized service",
			respJson: "",
		},
		{
			name:     "test2",
			args:     []string{"-service", "Users", "-method", "GetUser1", "-request", `{"email":"john@doe.com","id":"user-123"}`, l.Addr().String()},
			output:   "",
			errMsg:   "Invalid gRPC method",
			respJson: "",
		},
		{
			name:     "test3",
			args:     []string{"-service", "Users", "-method", "GetUser", "-request", `{"email":"john@doe.com","id":"user-123"}`, l.Addr().String()},
			output:   "",
			errMsg:   "",
			respJson: `{"user":{"id":"user-123","firstName":"john","lastName":"doe.com","age":36}}`,
		},
		{
			name:     "test4",
			args:     []string{"-service", "Repo", "-method", "GetFoo", "-request", `{"email":"john@doe.com","id":"user-123"}`, l.Addr().String()},
			errMsg:   "Invalid gRPC method",
			output:   "",
			respJson: "",
		},
		{
			name:     "test5",
			args:     []string{"-service", "Repo", "-method", "GetRepos", "-request", `{"id":"1"}`, l.Addr().String()},
			errMsg:   "",
			output:   "",
			respJson: `{"repo":[{"id":"repo-123","name":"hsh","url":"github.com","owner":{"id":"user-123"}}]}`,
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := HandleGrpc(w, tc.args)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tc.errMsg {
				t.Errorf("Expected error message `%s`, got `%s`", tc.errMsg, errMsg)
			}

			if len(tc.output) != 0 {
				output := w.String()
				if diff := cmp.Diff(output, tc.output); diff != "" {
					t.Errorf("Expected output to be: %#v, Got: %#v", tc.output, output)
				}
			}

			if len(tc.respJson) != 0 {
				respJson := w.String()
				sanitizedRespJson := strings.Replace(string(respJson), " ", "", -1)
				sanitizedRespJson = strings.Replace(string(sanitizedRespJson), "\n", "", -1)

				if sanitizedRespJson != tc.respJson {
					t.Errorf("Expected result: %v Got: %v", tc.respJson, sanitizedRespJson)
				}
			}
		})
		w.Reset()
	}
}

func TestGrpcCmdFlagParsing(t *testing.T) {
	usageMessage := `
grpc: A gRPC client.
 
grpc: <options> server

Options:
  -method string
    	Method to call
  -pretty-print
    	Pretty print the JSON output
  -request string
    	Request to send
  -service string
    	gRpc service to send the request to
`
	tests := []struct {
		name   string
		args   []string
		output string
		errMsg string
	}{
		{
			name:   "test1",
			args:   []string{},
			output: "",
			errMsg: ErrNoServerSpecified.Error(),
		},
		{
			name:   "test2",
			args:   []string{"-h"},
			output: usageMessage,
			errMsg: flag.ErrHelp.Error(),
		},
		{
			name:   "test3",
			args:   []string{"-service", "Users", "localhost:50051"},
			output: "",
			errMsg: ErrInvalidGrpcMethod.Error(),
		},
	}

	w := new(bytes.Buffer)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := HandleGrpc(w, tc.args)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if tc.errMsg != errMsg {
				t.Fatalf("Expected error %v, got %v", tc.errMsg, errMsg)
			}

			if len(tc.output) != 0 {
				gotoutput := w.String()
				if tc.output != gotoutput {
					t.Fatalf("Expected output to be: %#v, Got: %#v", tc.output, gotoutput)
				}
			}
			w.Reset()
		})
	}
}
