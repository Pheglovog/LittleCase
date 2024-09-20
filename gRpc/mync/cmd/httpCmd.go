package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mync/middleware"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type httpConfig struct {
	url             string
	verb            string
	postBody        string
	disableRedirect bool
	headers         []string
	basicAuth       string
	report          bool
	numRequests     int
	maxIdleConns    int
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

func addHeaders(c httpConfig, req *http.Request) {
	for _, h := range c.headers {
		kv := strings.Split(h, "=")
		req.Header.Add(kv[0], kv[1])
	}
}

func addBasicAuth(c httpConfig, req *http.Request) {
	if len(c.basicAuth) != 0 {
		up := strings.Split(c.basicAuth, "=")
		req.SetBasicAuth(up[0], up[1])
	}
}

func HandleHttp(w io.Writer, args []string) error {
	var outputFile string
	var postBodyFile string
	var responseBody []byte
	var req *http.Request
	var httpClient http.Client
	var ctx context.Context
	var redirectPolicyFunc func(req *http.Request, via []*http.Request) error
	c := httpConfig{}

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&outputFile, "output", "", "File path to write the response into")
	fs.StringVar(&c.postBody, "body", "", "JSON data for HTTP POST request")
	fs.StringVar(&postBodyFile, "body-file", "", "File containing JSON data for HTTP POST request")
	fs.BoolVar(&c.disableRedirect, "disable-redirect", false, "Do not follow redirection request")
	fs.StringVar(&c.basicAuth, "basicAuth", "", "Add basic auth (username:password) credentials to the outgoing request")
	fs.BoolVar(&c.report, "report", false, "report this http request's latency")
	fs.IntVar(&c.numRequests, "num-requests", 1, "Number of requests to make")
	fs.IntVar(&c.maxIdleConns, "max-idle-conns", 0, "Maximum number of idle connections for the connection pool")

	fs.Func("header", "Add one or more headers to the outgoing request (key=value)", func(s string) error {
		c.headers = append(c.headers, s)
		return nil
	})

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
		return FlagParsingError{err}
	}

	if fs.NArg() != 1 {
		return InvalidInputError{ErrNoServerSpecified}
	}

	if postBodyFile != "" && c.postBody != "" {
		return InvalidInputError{ErrInvalidHTTPPostCommand}
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
		return InvalidInputError{err}
	}

	c.url = fs.Arg(0)

	if c.disableRedirect {
		redirectPolicyFunc = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 1 {
				return errors.New("stopped after 1 redirect")
			}
			return nil
		}
	}
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          c.maxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	httpLatencyMiddleWare := middleware.HttpLatencyClient{
		Logger:    log.New(os.Stdout, "", log.LstdFlags),
		Transport: t,
	}
	httpClient = http.Client{
		CheckRedirect: redirectPolicyFunc,
		Transport:     httpLatencyMiddleWare,
	}
	if c.report {
		l := log.New(w, "", log.LstdFlags)
		httpClient.Transport = middleware.HttpLatencyClient{Logger: l}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	switch c.verb {
	case http.MethodGet:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
		if err != nil {
			return err
		}
	case http.MethodPost:
		postBodyReader := strings.NewReader(c.postBody)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.url, postBodyReader)
		if err != nil {
			return err
		}
		c.headers = append(c.headers, "Content-Type=application/json")
	}

	addHeaders(c, req)
	addBasicAuth(c, req)

	for i := 0; i < c.numRequests; i++ {
		r, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		responseBody, err = io.ReadAll(r.Body)
		if err != nil {
			return err
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
	}
	return nil
}
