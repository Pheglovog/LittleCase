package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
)

type httpConfig struct {
	url  string
	verb string
}

func validateConfig(c httpConfig) error {
	allowVerbs := []string{"GET", "POST", "HEAD"}
	for _, v := range allowVerbs {
		if c.verb == v {
			return nil
		}
	}
	return ErrInvalidHTTPMethod
}

func fetchResource(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func HandleHttp(w io.Writer, args []string) error {
	c := httpConfig{}
	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.verb, "verb", "GET", "HTTP method")

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

	err = validateConfig(c)
	if err != nil {
		if errors.Is(err, ErrInvalidHTTPMethod) {
			fmt.Fprint(w, err)
		}
		return err
	}

	c.url = fs.Arg(0)
	data, err := fetchResource(c.url)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
