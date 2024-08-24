package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type config struct {
	NumTimes   int
	PrintUsage bool
}

var usageString = fmt.Sprintf(`Usage : %s <integer> [-h|--help]
A greeter application which prints the name you entered <integer> number of times.
`, os.Args[0])

func printUsage(w io.Writer) {
	fmt.Fprint(w, usageString)
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

func parseArgs(args []string) (config, error) {
	var numTimes int
	var err error
	c := config{}
	if len(args) != 1 {
		return c, errors.New("invalid number of arguments")
	}

	if args[0] == "-h" || args[0] == "--help" {
		c.PrintUsage = true
		return c, nil
	}

	numTimes, err = strconv.Atoi(args[0])
	if err != nil {
		return c, err
	}
	c.NumTimes = numTimes

	return c, nil
}

func validateArgs(c config) error {
	if !(c.NumTimes > 0) && !c.PrintUsage {
		return errors.New("must specify a number greater than 0")
	}
	return nil
}

func runCmd(r io.Reader, w io.Writer, c config) error {
	if c.PrintUsage {
		printUsage(w)
		return nil
	}

	name, err := getName(r, w)
	if err != nil {
		return err
	}

	greetUser(c, name, w)

	return nil
}

func greetUser(c config, name string, w io.Writer) {
	msg := fmt.Sprintf("Nice to meet you %s\n", name)
	for i := 0; i < c.NumTimes; i++ {
		fmt.Fprint(w, msg)
	}
}

func main() {
	c, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
	}

	err = validateArgs(c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		printUsage(os.Stdout)
		os.Exit(1)
	}

	err = runCmd(os.Stdin, os.Stdout, c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
