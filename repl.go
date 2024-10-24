package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

const (
	EXIT_SUCCESS = iota
	EXIT_ERROR
)

const (
	KeyCtrlC     = 3
	KeyCtrlD     = 4
	KeyEnter     = '\r'
	KeyNewLine   = '\n'
	KeyBackspace = 127
	KeyUnknown   = 256 + iota
	KeyAltLeft = 259 + iota
	KeyAltRight
)

const (
	KeyArrowUp = 65 + iota
	KeyArrowBottom
	KeyArrowRight
	KeyArrowLeft
)

// Repl as the acronym for Read Eval Print and Loop
// is the orchestrator of this shell
func Repl(rd io.Reader) {
	exitCommands := []string{"exit\n", "quit\n"}

	if rd == os.Stdin {
		fd := int(os.Stdin.Fd())
		termState, err := term.MakeRaw(fd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(EXIT_ERROR)
		}
		defer term.Restore(fd, termState)
	}

	for {
		printSingleLineCmdPrompt()
		input, err := readFromSource(rd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(EXIT_ERROR)
		}

		if slices.Contains(exitCommands, input) {
			input = strings.Replace(input, "\n", "\r", 1)
			fmt.Printf("\n%s", input)
			break
		}

		fmt.Printf("\n%s", input)
	}

	os.Exit(EXIT_SUCCESS)
}

// readFromSource read a rune at a time from a reader
// and return a string and an error if there is one.
//
// It also have the ability to read a multiline from a stdin.
func readFromSource(source io.Reader) (line string, err error) {
	reader := bufio.NewReader(source)
	quotes := []byte{'"', '\''}
	quoteOpened := false
	cursorPos := uint64(0)

	for {
		char, r_err := reader.ReadByte()

		if r_err != nil {
			err = r_err
			break
		}

		if slices.Contains([]byte{KeyCtrlC, KeyCtrlD}, char) {
			os.Exit(EXIT_ERROR)
		}

		if char == KeyBackspace {
			if len(line) == 0 {
				continue
			} else {
				line = line[:len(line)-1]
				fmt.Print("\b\033[K")
				cursorPos--
				continue
			}
		}

		if char == '\033' {
			if b_err := moveCursor(reader, &line, &cursorPos); b_err != nil {
				err = b_err
				break
			}
			continue
		}

		fmt.Print(string(char))
		cursorPos++

		if slices.Contains(quotes, char) {
			handleQuote(&line, &quoteOpened, char)
			continue
		}

		if char == KeyEnter {
			if quoteOpened {
				printMultiLinesCmdPrompt()
				continue
			} else if strings.HasSuffix(line, "\\") {
				line, _ = strings.CutSuffix(line, "\\")
				printMultiLinesCmdPrompt()
				continue
			} else {
				line += string(KeyNewLine)
				break
			}
		}

		line += string(char)
	}

	return
}

// moveCursor handle arrow key hit and move the cursor position accordingly
func moveCursor(reader *bufio.Reader, line *string, cursorPos *uint64) (err error) {

	reader.ReadByte()
	b, b_err := reader.ReadByte()
	if b_err != nil {
		err = b_err
		return
	}

	if int64(len(*line)) == 0 {
		return
	}

	if b == KeyArrowLeft && *cursorPos > 0 {
		*cursorPos--
		fmt.Printf("\033[%s", string(b))
	} else if b == KeyArrowRight && *cursorPos < uint64(len(*line)) {
		*cursorPos++
		fmt.Printf("\033[%s", string(b))
	}

	return;
}

// handleQuote deal with quote contains in the input
func handleQuote(line *string, state *bool, char byte) {
	c := string(char)

	if len(*line) == 0 {
		*line += c
		*state = true
		return
	}

	if *state {
		*line += c
		*state = false
		return
	}

	if !*state && strings.HasSuffix(*line, c) {
		*line, _ = strings.CutSuffix(*line, c)
	} else {
		*line += c
	}

	*state = true
}

// printSingleLineCmdPrompt is the first prompt lineing.
// It wait the one that wait for you to enter a command
func printSingleLineCmdPrompt() {
	fmt.Fprint(os.Stdout, "\r$ ")
}

// printMultiLinesCmdPrompt is the second prompt string.
// It's the one showed when entering a multi line command.
func printMultiLinesCmdPrompt() {
	fmt.Fprint(os.Stdout, "\n> ")
}
