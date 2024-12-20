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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userService struct {
	svc.UnimplementedUsersServer
}

func (s *userService) GetUser(ctx context.Context, in *svc.UserGetRequest) (*svc.UserGetReply, error) {
	log.Printf("Received request for user with Email: %s Id: %s\n", in.Email, in.Id)
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

func (s *userService) GetHelp(stream svc.Users_GetHelpServer) error {
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
		response := svc.UserHelpReply{
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

type repoService struct {
	svc.UnimplementedRepoServer
}

func (s *repoService) GetRepos(in *svc.RepoGetRequest, stream svc.Repo_GetReposServer) error {
	log.Printf("Received request for repo with CreateId: %s Id: %s\n", in.CreatorId, in.Id)
	cnt := 1

	for {
		repo := svc.Repository{
			Id:    in.Id,
			Owner: &svc.User{Id: in.CreatorId, FirstName: "Huang"},
			Name:  fmt.Sprintf("repo-%d", cnt),
		}
		repo.Url = fmt.Sprintf("https://github/%s", repo.Name)

		r := svc.RepoGetReply{
			Repo: &repo,
		}
		if err := stream.Send(&r); err != nil {
			return err
		}
		if cnt >= 5 {
			break
		}
		cnt++
	}

	return nil
}

func (s *repoService) CreateBuild(in *svc.Repository, stream svc.Repo_CreateBuildServer) error {
	log.Printf("Received build request for repo: %s\n", in.Name)

	logLine := svc.RepoBuildLog{
		LogLine:   fmt.Sprintf("Starting build for repository:%s", in.Name),
		Timestamp: timestamppb.Now(),
	}
	if err := stream.Send(&logLine); err != nil {
		return err
	}

	cnt := 1
	for {
		logLine := svc.RepoBuildLog{
			LogLine:   fmt.Sprintf("Build log line  - %d", cnt),
			Timestamp: timestamppb.Now(),
		}
		if err := stream.Send(&logLine); err != nil {
			return err
		}
		if cnt >= 3 {
			break
		}
		time.Sleep(300 * time.Millisecond)
		cnt++
	}
	logLine = svc.RepoBuildLog{
		LogLine:   fmt.Sprintf("Finished build for repository:%s", in.Name),
		Timestamp: timestamppb.Now(),
	}
	if err := stream.Send(&logLine); err != nil {
		return err
	}
	return nil
}

func (s *repoService) CreateRepo(stream svc.Repo_CreateRepoServer) error {
	var repoContext *svc.RepoContext
	var data []byte
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		switch t := r.Body.(type) {
		case *svc.RepoCreateRequest_Context:
			repoContext = r.GetContext()
		case *svc.RepoCreateRequest_Data:
			b := r.GetData()
			data = append(data, b...)
		case nil:
			return status.Error(codes.InvalidArgument, "Message doesn't contain context or data")
		default:
			return status.Errorf(codes.FailedPrecondition, "Unexpected message type:%s", t)
		}
	}

	repo := svc.Repository{
		Name: repoContext.Name,
		Url:  fmt.Sprintf("https://git.example.com/%s/%s", repoContext.CreatorId, repoContext.Name),
	}
	r := svc.RepoCreateReply{
		Repo: &repo,
		Size: int32(len(data)),
	}
	return stream.SendAndClose(&r)
}

func registerServices(s *grpc.Server) {
	svc.RegisterUsersServer(s, &userService{})
	svc.RegisterRepoServer(s, &repoService{})
	reflection.Register(s)
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

	s := grpc.NewServer()
	registerServices(s)
	log.Fatal(startServer(s, lis))
}
