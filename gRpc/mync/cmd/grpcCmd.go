package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	svc "service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type grpcConfig struct {
	server      string
	method      string
	request     string
	service     string
	prettyPrint bool
}

func setupGrpcConn(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func getUserServiceClient(conn *grpc.ClientConn) svc.UsersClient {
	return svc.NewUsersClient(conn)
}

func getUser(client svc.UsersClient, u *svc.UserGetRequest) (*svc.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func createUserRequest(jsonQuery string) (*svc.UserGetRequest, error) {
	u := svc.UserGetRequest{}
	input := []byte(jsonQuery)
	err := protojson.Unmarshal(input, &u)
	return &u, err
}

func getUserResponseJson(c grpcConfig, resp *svc.UserGetReply) ([]byte, error) {
	if c.prettyPrint {
		return []byte(protojson.Format(resp)), nil
	}
	return protojson.Marshal(resp)
}

func getRepoServiceClient(conn *grpc.ClientConn) svc.RepoClient {
	return svc.NewRepoClient(conn)
}

func createGetRepoRequest(jsonQuery string) (*svc.RepoGetRequest, error) {
	u := svc.RepoGetRequest{}
	input := []byte(jsonQuery)
	return &u, protojson.Unmarshal(input, &u)
}

func getRepos(client svc.RepoClient, r *svc.RepoGetRequest) (*svc.RepoGetReply, error) {
	return client.GetRepos(context.Background(), r)
}

func getReposResponseJson(c grpcConfig, result *svc.RepoGetReply) ([]byte, error) {
	if c.prettyPrint {
		return []byte(protojson.Format(result)), nil
	}
	return protojson.Marshal(result)
}

func callUserMethod(usersClient svc.UsersClient, c grpcConfig) ([]byte, error) {
	switch c.method {
	case "GetUser":
		req, err := createUserRequest(c.request)
		if err != nil {
			return nil, InvalidInputError{Err: err}
		}
		result, err := getUser(usersClient, req)
		if err != nil {
			return nil, err
		}
		respData, err := getUserResponseJson(c, result)
		return respData, err
	case "":
		return nil, ErrInvalidGrpcMethod
	default:
		return nil, ErrInvalidGrpcMethod
	}
}

func callRepoMethod(repoClient svc.RepoClient, c grpcConfig) ([]byte, error) {
	switch c.method {
	case "GetRepos":
		req, err := createGetRepoRequest(c.request)
		if err != nil {
			return nil, InvalidInputError{Err: err}
		}
		result, err := getRepos(repoClient, req)
		if err != nil {
			return nil, err
		}
		respData, err := getReposResponseJson(c, result)
		return respData, err
	case "":
		return nil, ErrInvalidGrpcMethod
	default:
		return nil, ErrInvalidGrpcMethod
	}
}

func validateGrpcConfig(c grpcConfig) error {
	if len(c.service) == 0 {
		return errors.New("unrecognized service")
	}
	if len(c.method) == 0 {
		return ErrInvalidGrpcMethod
	}

	return nil
}

func HandleGrpc(w io.Writer, args []string) error {
	c := grpcConfig{}
	fs := flag.NewFlagSet("grpc", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.method, "method", "", "Method to call")
	fs.StringVar(&c.request, "request", "", "Request to send")
	fs.StringVar(&c.service, "service", "", "gRpc service to send the request to")
	fs.BoolVar(&c.prettyPrint, "pretty-print", false, "Pretty print the JSON output")

	fs.Usage = func() {
		var usageString = `
grpc: A gRPC client.
 
grpc: <options> server`
		fmt.Fprint(w, usageString)
		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return ErrNoServerSpecified
	}
	c.server = fs.Arg(0)

	err = validateGrpcConfig(c)
	if err != nil {
		return err
	}

	conn, err := setupGrpcConn(c.server)
	if err != nil {
		return err
	}

	var usersClient svc.UsersClient

	switch c.service {
	case "Users":
		usersClient = getUserServiceClient(conn)
		respJson, err := callUserMethod(usersClient, c)
		if err != nil {
			return err
		}
		fmt.Fprint(w, string(respJson))
	case "Repo":
		repoClient := getRepoServiceClient(conn)
		respJson, err := callRepoMethod(repoClient, c)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(respJson))
	default:
		return errors.New("unrecognized service")
	}
	return nil
}
