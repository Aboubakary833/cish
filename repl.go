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

const NULChar = '\x00'

const (
	KeyCtrlC     = 3
	KeyCtrlD     = 4
	KeyEnter     = '\r'
	KeyNewLine   = '\n'
	keyArrow     = '\033'
	keyBackSlace = '\\'
	KeyBackspace = 127
	KeyUnknown   = 256 + iota
	KeyAltLeft   = 259 + iota
	KeyAltRight
)

const (
	KeyArrowUp = 65 + iota
	KeyArrowBottom
	KeyArrowRight
	KeyArrowLeft
)

// Shell prompts
const (
	PS1 = iota + 1
	PS2
)

var (
	Quotes   = []byte{'"', '\''}
	quitKeys = []byte{KeyCtrlC, KeyCtrlD}
)

type Input struct {
	reader       *bufio.Reader
	quotesOpened bool
	openedQuote  byte
	shouldEscape bool
	cursorPos    uint64
	prompt       int
	buffer       string
	sourceFd	 int
	termState *term.State
}

func newInput(source io.Reader, sourceFd int, state *term.State) *Input {
	return &Input{
		reader:       bufio.NewReader(source),
		quotesOpened: false,
		openedQuote:  NULChar,
		shouldEscape: false,
		cursorPos:    uint64(0),
		prompt:       PS1,
		buffer:       "",
		sourceFd: sourceFd,
		termState: state,
	}
}

func (input *Input) read() (err error) {
	input.printPS1Prompt()

L:
	for {
		key, b_err := input.reader.ReadByte()

		if b_err != nil {
			err = b_err
			break
		}

		// Escape arrow keys when printing to stdout
		if key != keyArrow {
			fmt.Print(string(key))
		}

		switch true {

		case slices.Contains(quitKeys, key):
			quitRawMode(input.sourceFd, input.termState, EXIT_ERROR)

		case slices.Contains(Quotes, key):
			input.handleQuote(key)

		case key == KeyBackspace:
			input.handleBackspace()

		case key == keyBackSlace:
			input.handleBackSlace()

		case key == keyArrow:
			if b_err := input.moveCursor(); b_err != nil {
				err = b_err
				break L
			}

		case key == KeyEnter:
			if input.handleKeyEnter() {
				break L
			}

		default:
			input.appendToBuffer(key)
			if input.shouldEscape {
				input.shouldEscape = false
			}
		}
	}

	return
}

func (input *Input) hasSuffix(str string) bool {
	return strings.HasSuffix(input.buffer, str)
}

func (input *Input) appendToBuffer(char byte) {
	bufferLen := input.bufferLen()

	if bufferLen == 0 || input.cursorPos == bufferLen {
		input.buffer += string(char)
		input.cursorPos++
		return
	}

	firstChunk := input.buffer[:input.cursorPos]
	lastChunk := input.buffer[input.cursorPos:bufferLen]

	input.buffer = firstChunk + string(char) + lastChunk
	input.cursorPos++
}

func (input *Input) handleQuote(char byte) {
	if input.shouldEscape {
		input.appendToBuffer(char)
		input.shouldEscape = false
		return
	}

	if input.bufferLen() == 0 || (!input.quotesOpened && !input.shouldEscape) {
		input.appendToBuffer(char)
		input.quotesOpened = true
		input.openedQuote = char
		return
	}

	if !input.quotesOpened && input.shouldEscape {
		input.appendToBuffer(char)
		return
	}

	if char == input.openedQuote && !input.shouldEscape {
		input.appendToBuffer(char)
		input.quotesOpened = false
		input.openedQuote = NULChar
	} else {
		input.appendToBuffer(char)
	}
}

func (input *Input) handleBackSlace() {
	if input.shouldEscape {
		input.appendToBuffer(keyBackSlace)
		input.shouldEscape = false
		return
	}

	input.appendToBuffer(keyBackSlace)
	input.shouldEscape = true
}

func (input *Input) handleKeyEnter() bool {
	if input.quotesOpened {
		input.printPS2Prompt()
		return false
	}

	if input.bufferLen() == 0 {
		return true
	}

	buffer := input.buffer
	backSlace := string(keyBackSlace)

	if input.hasSuffix(backSlace) {

		prevChar := string(buffer[input.cursorPos-1])

		if input.bufferLen() == 1 || input.shouldEscape {
			input.buffer, _ = strings.CutSuffix(buffer, backSlace)
			input.cursorPos--
			input.shouldEscape = false
			input.printPS2Prompt()
			return false
		} else if strings.EqualFold(prevChar, backSlace) {
			input.appendToBuffer(KeyNewLine)
			return true
		}
	}

	input.appendToBuffer(KeyNewLine)

	return true
}

func (input *Input) handleBackspace() {
	if len(input.buffer) == 0 {
		return
	}

	if input.quotesOpened && input.hasSuffix(string(input.openedQuote)) {
		input.quotesOpened = false
		input.openedQuote = NULChar
	}

	input.buffer = input.buffer[:input.bufferLen()-1]
	fmt.Print("\b\033[K")
	input.cursorPos--
}

func (input *Input) bufferLen() uint64 {
	return uint64(len(input.buffer))
}

func (input *Input) cursorIsPeak() bool {
	return input.cursorPos == uint64(len(input.buffer))
}

func (input *Input) moveCursor() (err error) {
	input.reader.ReadByte()
	key, b_err := input.reader.ReadByte()
	if b_err != nil {
		err = b_err
		return
	}

	if len(input.buffer) == 0 || !slices.Contains([]byte{KeyArrowLeft, KeyArrowRight}, key) {
		return
	}
	
	if key == KeyArrowLeft && input.cursorPos > 0 {
		input.cursorPos--
	} else if key == KeyArrowRight && !input.cursorIsPeak() {
		input.cursorPos++
	} else {
		return
	}

	fmt.Printf("\033[%s", string(key))

	return
}

func (input *Input) printPS1Prompt() {
	if input.prompt != PS1 {
		input.prompt = PS1
	}

	fmt.Fprint(os.Stdout, "\r$ ")
}

func (input *Input) printPS2Prompt() {
	if input.prompt != PS2 {
		input.prompt = PS2
	}

	fmt.Fprint(os.Stdout, "\n> ")
}

// Repl is the acronym for Read Eval Print and Loop.
// So, it's the orchestrator of this shell
func Repl(rd io.Reader) {
	exitCommands := []string{"exit\n", "quit\n"}
	stdinFd := int(os.Stdin.Fd())

	state := enterRawMode(stdinFd)

	for {
		input := newInput(rd, stdinFd, state)

		if err := input.read(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			quitRawMode(stdinFd, state, EXIT_ERROR)
		}

		if slices.Contains(exitCommands, input.buffer) {
			fmt.Print("\n")
			break
		}

		fmt.Printf("\n%s", input.buffer)
	}

	quitRawMode(stdinFd, state, EXIT_SUCCESS)
}

func enterRawMode(sourceFd int) (state *term.State) {
	state, err := term.MakeRaw(sourceFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	
	return
}

func quitRawMode(sourceFd int, state *term.State, status int) {
	if t_err := term.Restore(sourceFd, state); t_err != nil {
		fmt.Fprintln(os.Stderr, t_err.Error())
		os.Exit(EXIT_ERROR)
	}

	os.Exit(status)
}
