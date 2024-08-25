package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var errInvalidPosArgSpecified = errors.New("more than one positional argument specified")

type config struct {
	NumTimes int
	Name     string
}

func getName(r io.Reader, w io.Writer) (string, error) {
	msg := "Your name Please?\n"
	fmt.Fprint(w, msg)

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	name := scanner.Text()
	if len(name) == 0 {
		return "", errors.New("you didn't enter your name")
	}

	return name, nil
}

func parseArgs(w io.Writer, args []string) (config, error) {
	c := config{}
	fs := flag.NewFlagSet("greeter", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.IntVar(&c.NumTimes, "n", 0, "Number of times to greet")

	fs.Usage = func() {
		var usageString = `
A greeter application which prints the name you entered a specified number of times.
 
Usage of %s: <option> [name]`

		fmt.Fprintf(w, usageString, fs.Name())
		fmt.Fprintln(w)
		fs.PrintDefaults()
	}
	err := fs.Parse(args)
	if err != nil {
		return c, err
	}
	if fs.NArg() > 1 {
		return c, errInvalidPosArgSpecified
	}
	if fs.NArg() == 1 {
		c.Name = fs.Arg(0)
	}
	return c, nil
}

func validateArgs(c config) error {
	if !(c.NumTimes > 0) {
		return errors.New("must specify a number greater than 0")
	}
	return nil
}

func runCmd(r io.Reader, w io.Writer, c config) error {

	var err error
	if len(c.Name) == 0 {
		c.Name, err = getName(r, w)
		if err != nil {
			return err
		}
	}
	greetUser(c, w)

	return nil
}

func greetUser(c config, w io.Writer) {
	msg := fmt.Sprintf("Nice to meet you %s\n", c.Name)
	for i := 0; i < c.NumTimes; i++ {
		fmt.Fprint(w, msg)
	}
}

func main() {
	c, err := parseArgs(os.Stderr, os.Args[1:])
	if err != nil {
		if errors.Is(err, errInvalidPosArgSpecified) {
			fmt.Fprintln(os.Stdout, err)
		}
		os.Exit(1)
	}

	err = validateArgs(c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	err = runCmd(os.Stdin, os.Stdout, c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
