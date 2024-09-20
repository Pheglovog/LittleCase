package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	svc "service"
	"strings"
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
	stream, err := repoClient.GetRepos(context.Background(), &svc.RepoGetRequest{
		CreatorId: "user-123",
		Id:        "repo-123",
	})
	if err != nil {
		t.Fatal(err)
	}

	var repos []*svc.Repository
	for {
		repo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, repo.Repo)
	}

	if len(repos) != 5 {
		t.Fatalf("Expected to get back 5 repo, got back: %d repos", len(repos))
	}

	for idx, repo := range repos {
		gotRepoName := repo.Name
		expectedRepoName := fmt.Sprintf("repo-%d", idx+1)
		if gotRepoName != expectedRepoName {
			t.Errorf("Expected Repo Name to be: %s, Got: %s", expectedRepoName, gotRepoName)
		}
	}
}

func TestRepoBuildMethod(t *testing.T) {

	s, l := startTestGrpcServer()
	defer s.GracefulStop()

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
	repoClient := svc.NewRepoClient(client)
	stream, err := repoClient.CreateBuild(
		context.Background(),
		&svc.Repository{Name: "myrepo"},
	)
	if err != nil {
		t.Fatal(err)
	}
	var logLines []*svc.RepoBuildLog
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		logLines = append(logLines, line)
	}
	if len(logLines) != 5 {
		t.Fatalf("Expected to get back 3 lines in the log, got back: %d repos", len(logLines))
	}

	expectedFirstLine := "Starting build for repository:myrepo"
	if logLines[0].LogLine != expectedFirstLine {
		t.Fatalf("Expected first line to be:%s, Got:%s", expectedFirstLine, logLines[0].LogLine)
	}
	expectedLastLine := "Finished build for repository:myrepo"
	if logLines[4].LogLine != expectedLastLine {
		t.Fatalf("Expected last line to be:%s,Got:%s", expectedLastLine, logLines[4].LogLine)
	}

	logLine := logLines[0]
	if err := logLine.Timestamp.CheckValid(); err != nil {
		t.Fatalf("Logline timestamp invalid: %#v", logLine)
	}
}

func TestCreateRepo(t *testing.T) {
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
	stream, err := repoClient.CreateRepo(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	c := svc.RepoCreateRequest_Context{
		Context: &svc.RepoContext{
			CreatorId: "user-123",
			Name:      "test-repo",
		},
	}
	r := svc.RepoCreateRequest{
		Body: &c,
	}
	err = stream.Send(&r)
	if err != nil {
		t.Fatal("StreamSend", err)
	}

	data := "Arbitrary Data Bytes"
	repoData := strings.NewReader(data)
	for {
		b, err := repoData.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal("StreamSend", err)
		}
		bData := svc.RepoCreateRequest_Data{
			Data: []byte{b},
		}
		r := svc.RepoCreateRequest{
			Body: &bData,
		}
		err = stream.Send(&r)
		if err != nil {
			t.Fatal("StreamSend", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatal("CloseAndRecv", err)
	}
	expectedSize := int32(len(data))
	if resp.Size != expectedSize {
		t.Errorf(
			"Expected Repo Created to be: %d bytes Got back: %d",
			expectedSize,
			resp.Size,
		)
	}
	expectedRepoUrl := "https://git.example.com/user-123/test-repo"
	if resp.Repo.Url != expectedRepoUrl {
		t.Errorf(
			"Expected Repo URL to be: %s, Got: %s",
			expectedRepoUrl,
			resp.Repo.Url,
		)
	}
}
