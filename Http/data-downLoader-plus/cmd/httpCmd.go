package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type httpConfig struct {
	url      string
	verb     string
	postBody string
}

func validateConfig(c httpConfig) error {
	var validMethod bool
	allowVerbs := []string{http.MethodGet, http.MethodPost, http.MethodHead}
	for _, v := range allowVerbs {
		if c.verb == v {
			validMethod = true
		}
	}
	if !validMethod {
		return ErrInvalidHTTPMethod
	}

	if c.verb == http.MethodPost && c.postBody == "" {
		return ErrInvalidHTTPPostRequest
	}

	if c.verb != http.MethodPost && c.postBody != "" {
		return ErrInvalidHTTPCommand
	}

	return nil
}

func fetchRemoteResource(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func createRemoteResource(url string, body io.Reader) ([]byte, error) {
	r, err := http.Post(url, "application/json", body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return io.ReadAll(r.Body)
}

func HandleHttp(w io.Writer, args []string) error {
	var outputFile string
	var postBodyFile string
	var responseBody []byte
	c := httpConfig{}

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&outputFile, "output", "", "File path to write the response into")
	fs.StringVar(&c.postBody, "body", "", "JSON data for HTTP POST request")
	fs.StringVar(&postBodyFile, "body-file", "", "File containing JSON data for HTTP POST request")

	fs.Usage = func() {
		var usageString = `
http: A HTTP client.
 
http: <options> server`
		fmt.Fprint(w, usageString)

		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options: ")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return ErrNoServerSpecified
	}

	if postBodyFile != "" && c.postBody != "" {
		return ErrInvalidHTTPPostCommand
	}

	if c.verb == http.MethodPost && postBodyFile != "" {
		data, err := os.ReadFile(postBodyFile)
		if err != nil {
			return err
		}
		c.postBody = string(data)
	}

	err = validateConfig(c)
	if err != nil {
		fmt.Fprint(w, err)
		return err
	}

	c.url = fs.Arg(0)

	switch c.verb {
	case http.MethodGet:
		responseBody, err = fetchRemoteResource(c.url)
		if err != nil {
			return err
		}
	case http.MethodPost:
		postBodyReader := strings.NewReader(c.postBody)
		responseBody, err = createRemoteResource(c.url, postBodyReader)
		if err != nil {
			return err
		}
	}

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(responseBody)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Data saved to: %s\n", outputFile)
		return err
	}

	fmt.Fprintln(w, string(responseBody))
	return nil
}
