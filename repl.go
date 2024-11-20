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

const DELETE = "\b\033[K"
const ARROW_CHUNK = "\033["

// The four constant are A, B, C & D in decimal.
// They are combine with keyArrow constant to move
// the cursor inside the command.
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

type Command struct {
	reader       *bufio.Reader
	outputStream io.Writer
	quotesOpened bool
	openedQuote  byte
	shouldEscape bool
	cursorPos    uint64
	prompt       int
	buffer       string
	sourceFd     int
	termState    *term.State
}

func newCommand(source io.Reader, sourceFd int, state *term.State) *Command {
	return &Command{
		reader:       bufio.NewReader(source),
		outputStream: os.Stdout,
		quotesOpened: false,
		openedQuote:  NULChar,
		shouldEscape: false,
		cursorPos:    uint64(0),
		prompt:       PS1,
		buffer:       "",
		sourceFd:     sourceFd,
		termState:    state,
	}
}

func (cmd *Command) read() (err error) {
	cmd.printPS1Prompt()

L:
	for {
		key, b_err := cmd.reader.ReadByte()

		if b_err != nil {
			err = b_err
			break
		}

		// Print out the key
		cmd.printKey(key)

		switch true {

		case slices.Contains(quitKeys, key):
			quitRawMode(cmd.sourceFd, cmd.termState, EXIT_ERROR)

		case slices.Contains(Quotes, key):
			cmd.handleQuote(key)

		case key == KeyBackspace:
			cmd.handleBackspace()

		case key == keyBackSlace:
			cmd.handleBackSlace()

		case key == keyArrow:
			if b_err := cmd.moveCursor(); b_err != nil {
				err = b_err
				break L
			}

		case key == KeyEnter:
			if cmd.handleKeyEnter() {
				break L
			}

		default:
			cmd.appendToBuffer(key)
			if cmd.shouldEscape {
				cmd.shouldEscape = false
			}
		}
	}

	return
}

func (cmd *Command) hasSuffix(str string) bool {
	return strings.HasSuffix(cmd.buffer, str)
}

// appendToBuffer append the typed key to the buffer
// at the current cursor position
func (cmd *Command) appendToBuffer(char byte) {
	bufferLen := cmd.bufferLen()

	if bufferLen == 0 || cmd.cursorIsPeak() {
		cmd.buffer += string(char)
		cmd.cursorPos++
		return
	}

	firstChunk := cmd.buffer[:cmd.cursorPos]
	lastChunk := cmd.buffer[cmd.cursorPos:bufferLen]

	cmd.buffer = firstChunk + string(char) + lastChunk
	cmd.cursorPos++
}

func (cmd *Command) handleQuote(char byte) {
	if cmd.shouldEscape {
		cmd.appendToBuffer(char)
		cmd.shouldEscape = false
		return
	}

	if cmd.bufferLen() == 0 || (!cmd.quotesOpened && !cmd.shouldEscape) {
		cmd.appendToBuffer(char)
		cmd.quotesOpened = true
		cmd.openedQuote = char
		return
	}

	if !cmd.quotesOpened && cmd.shouldEscape {
		cmd.appendToBuffer(char)
		return
	}

	if char == cmd.openedQuote && !cmd.shouldEscape {
		cmd.appendToBuffer(char)
		cmd.quotesOpened = false
	} else {
		cmd.appendToBuffer(char)
	}
}

func (cmd *Command) handleBackSlace() {
	if cmd.shouldEscape {
		cmd.appendToBuffer(keyBackSlace)
		cmd.shouldEscape = false
		return
	}

	cmd.appendToBuffer(keyBackSlace)
	cmd.shouldEscape = true
}

// handleKeyEnter execute when the Enter key is hit.
// It return false if the command is a multiline command.
// Otherwise, it return true
func (cmd *Command) handleKeyEnter() bool {
	if cmd.quotesOpened {
		cmd.printPS2Prompt()
		return false
	}

	if cmd.bufferLen() == 0 {
		return true
	}

	buffer := cmd.buffer
	backSlace := string(keyBackSlace)

	if cmd.hasSuffix(backSlace) {

		prevChar := string(buffer[cmd.cursorPos-1])

		if cmd.bufferLen() == 1 || cmd.shouldEscape {
			cmd.buffer, _ = strings.CutSuffix(buffer, backSlace)
			cmd.cursorPos--
			cmd.shouldEscape = false
			cmd.printPS2Prompt()
			return false
		} else if strings.EqualFold(prevChar, backSlace) {
			
			if !cmd.cursorIsPeak() {
				cmd.clearAndPrint()
			}
			
			cmd.appendToBuffer(KeyNewLine)
			return true
		}
	}

	if !cmd.cursorIsPeak() {
		cmd.clearAndPrint()
	}

	// put the newline key to the end of the cmd
	cmd.buffer += string(KeyNewLine)

	return true
}

// handleBackspace is executed when the backspace key is press
// and depending on the cmd states, determine what
// action should be done.
func (cmd *Command) handleBackspace() {
	if len(cmd.buffer) == 0 || cmd.cursorPos == 0 {
		return
	}

	if cmd.quotesOpened && cmd.hasSuffix(string(cmd.openedQuote)) {
		cmd.quotesOpened = false
		cmd.openedQuote = NULChar
	}

	bufferLen := cmd.bufferLen()

	if cmd.cursorIsPeak() {
		lastChar := cmd.buffer[bufferLen-1]
		cmd.buffer = cmd.buffer[:bufferLen-1]

		if bufferLen >= 2 && cmd.buffer[bufferLen - 2] == keyBackSlace {
			cmd.shouldEscape = true
		} else if lastChar == cmd.openedQuote {
			cmd.quotesOpened = true
		} 


		cmd.defaultPrint(DELETE)
		cmd.cursorPos--
		return
	}

	// Remove the char from buffer
	firstChunk := cmd.buffer[:cmd.cursorPos - 1]
	lastChunk := cmd.buffer[cmd.cursorPos:bufferLen]
	cmd.buffer = firstChunk + lastChunk
	cmd.cursorPos--
	
	cmd.defaultPrint(DELETE)
	cmd.defaultPrint(lastChunk)

	// Replace the cursor in the stdout
	for i := len(lastChunk); i > 0; i-- {
		cmd.defaultPrint(ARROW_CHUNK + string(rune(KeyArrowLeft)))
	}
}

// bufferLen return the length of the cmd buffer
func (cmd *Command) bufferLen() uint64 {
	return uint64(len(cmd.buffer))
}

// cursorIsPeak determine wether the cursor
// is at the end of the buffer or not
func (cmd *Command) cursorIsPeak() bool {
	return cmd.cursorPos == cmd.bufferLen()
}

// moveCursor handle the shell navigation through
// the arrows keys. It all modify the cmd cursor position.
func (cmd *Command) moveCursor() (err error) {

	var key byte
	var b_err error
	verticalKeys := []byte{KeyArrowLeft, KeyArrowRight}

	for i := 0; i < 2; i++ {
		if i == 0 {
			_, b_err = cmd.reader.ReadByte()
		} else {
			key, b_err = cmd.reader.ReadByte()
		}

		if b_err != nil {
			err = b_err
			return
		}
	}

	if len(cmd.buffer) == 0 || !slices.Contains(verticalKeys, key) {
		return
	}

	// Increase or decrease cursor depending on the key pressed.
	// Quit function if the key is one of the vertical keys.
	if key == KeyArrowLeft && cmd.cursorPos > 0 {
		cmd.cursorPos--
	} else if key == KeyArrowRight && !cmd.cursorIsPeak() {
		cmd.cursorPos++
	} else {
		return
	}

	cmd.defaultPrint(ARROW_CHUNK + string(key))

	return
}

func (cmd *Command) printPS1Prompt() {
	if cmd.prompt != PS1 {
		cmd.prompt = PS1
	}

	cmd.defaultPrint("\r$ ")
}

func (cmd *Command) printPS2Prompt() {
	if cmd.prompt != PS2 {
		cmd.prompt = PS2
	}

	cmd.defaultPrint("\n> ")
}

// defaultPrint is similar to `fmt.Print`, but redirect the output
// to Command output stream.
func (cmd *Command) defaultPrint(a ...any) (n int, err error) {
	n, err = fmt.Fprint(cmd.outputStream, a...)

	return
}

// printKey print out the typed key
func (cmd *Command) printKey(key byte) {

	previousChar := " "

	// Escape arrow keys when printing to stdout
	if key == keyArrow {
		return
	}
	
	if cmd.cursorIsPeak() {
		cmd.defaultPrint(string(key))
		return
	}

	firstChunk := cmd.buffer[:cmd.cursorPos]
	
	if len(firstChunk) != 0 {
		previousChar = firstChunk[cmd.cursorPos - 1:cmd.cursorPos]
	}
	lastChunk := cmd.buffer[cmd.cursorPos:cmd.bufferLen()]
	
	cmd.defaultPrint(DELETE)
	cmd.defaultPrint(previousChar + string(key))
	cmd.defaultPrint(lastChunk)

	// Replace the cursor in the stdout
	for i := len(lastChunk); i > 0; i-- {
		cmd.defaultPrint(ARROW_CHUNK + string(rune(KeyArrowLeft)))
	}
}

// Clear stdout and print out the command.
// This function also set the cursor position to peak.
func (cmd *Command) clearAndPrint() {
	cmd.defaultPrint(DELETE)
	cmd.printPS1Prompt()
	cmd.defaultPrint(cmd.buffer)
	cmd.defaultPrint(string(KeyEnter))
	cmd.cursorPos = cmd.bufferLen() - 1
}

// Repl is the acronym for Read Eval Print and Loop.
// So, it's the orchestrator of this shell
func Repl(rd io.Reader) {
	exitCommands := []string{"exit\n", "quit\n"}
	stdinFd := int(os.Stdin.Fd())

	state := enterRawMode(stdinFd)

	for {
		cmd := newCommand(rd, stdinFd, state)

		if err := cmd.read(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			quitRawMode(stdinFd, state, EXIT_ERROR)
		}

		if slices.Contains(exitCommands, cmd.buffer) {
			fmt.Print("\n")
			break
		}

		fmt.Printf("\n%s", cmd.buffer)
	}

	quitRawMode(stdinFd, state, EXIT_SUCCESS)
}

// enterRawMode put the terminal into the raw mode
// by disabling the default mode called canonical/cooked mode
func enterRawMode(sourceFd int) (state *term.State) {
	state, err := term.MakeRaw(sourceFd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	return
}

// quitRawMode restore the default canonical mode of the terminal
// This function is generally called when quiting the cish shell
func quitRawMode(sourceFd int, state *term.State, status int) {
	if t_err := term.Restore(sourceFd, state); t_err != nil {
		fmt.Fprintln(os.Stderr, t_err.Error())
		os.Exit(EXIT_ERROR)
	}

	os.Exit(status)
}
