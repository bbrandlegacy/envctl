package main

import (
	"errors"
	"fmt"
	"os"

	"envctl/internal/cli"
	"envctl/internal/runner"
)

func main() {
	root := cli.NewRootCommand()

	if err := root.Execute(); err != nil {
		var exitErr *runner.CommandExitError
		if errors.As(err, &exitErr) {
			if exitErr.Message != "" {
				fmt.Fprintln(os.Stderr, exitErr.Message)
			}
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
