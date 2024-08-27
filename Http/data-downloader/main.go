package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func fetchRemoteResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Must specify a HTTP URL to get data from")
	}

	body, err := fetchRemoteResource(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", body)
}
