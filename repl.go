package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const (
	EXIT_SUCCESS = iota
	EXIT_ERROR
)


// Repl as the acronym for Read Eval Print and Loop
// is the orchestrator of this shell
func Repl(rd io.Reader) {

	exitCommands := []string{"exit\n", "quit\n"}

	for {
		printSingleLineCmdPrompt()
		input, err := readFromSource(rd)

		if (input == "\n") {
			continue
		}

		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
			os.Exit(EXIT_ERROR)
		}

		if slices.Contains(exitCommands, input) {
			break
		}

	}

	os.Exit(EXIT_SUCCESS)
}

// readFromSource read a rune at a time from a reader
// and return a string and an error if there is one.
//
// It also have the ability to read a multiline from a stdin.
func readFromSource(source io.Reader) (str string, err error) {
	reader := bufio.NewReader(source)

	for {
		char, _, r_err := reader.ReadRune()
		
		if r_err != nil {
			err = r_err
			break
		}

		if char == '\n' {
			if strings.HasSuffix(str, "\\") {
				str, _ = strings.CutSuffix(str, "\\")
				printMultiLinesCmdPrompt()
				continue
			} else {
				str += "\n"
				break
			}
		}

		str += string(char)

	}

	return;
}

// printSingleLineCmdPrompt is the first prompt string.
// It wait the one that wait for you to enter a command
func printSingleLineCmdPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

// printMultiLinesCmdPrompt is the second prompt string.
// It's the one showed when entering a multi line command.
func printMultiLinesCmdPrompt() {
	fmt.Fprint(os.Stdout, "> ")
}
