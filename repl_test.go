package main

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestCommand(source io.Reader, output io.Writer) *Command {
	return &Command{
		reader: bufio.NewReader(source),
		outputStream: output,
		quotesOpened: false,
		openedQuote:  NULChar,
		shouldEscape: false,
		cursorPos:    uint64(0),
		prompt:       PS1,
		buffer:       "",
		sourceFd: 1,
		termState: nil,
	}
}

func TestAppendToBuffer(t *testing.T) {

	t.Run("it should set the buffer first char", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})

		cmd.appendToBuffer('A')

		assert.Equal(t, "A", cmd.buffer)
	})

	t.Run("it append the char to the buffer after the last char", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("Hello, Worl")
		text := "Hello, World"
		cmd.appendToBuffer('d')

		assert.Equal(t, text, cmd.buffer)
	})

	t.Run("it append the char to the buffer at the current cursor position", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("Hell, world")
		cmd.cursorPos = 4
		text := "Hello, world"

		cmd.appendToBuffer('o')

		assert.Equal(t, text, cmd.buffer)
	})
}

func TestHasSuffix(t *testing.T) {

	t.Run("it should return true", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		text := "Alola"

		for i := 0; i < len(text); i++ {
			cmd.appendToBuffer(text[i])
		}

		assert.True(t, cmd.hasSuffix("a"))
	})

	t.Run("it should return false", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("My code suck")

		assert.False(t, cmd.hasSuffix("b"))
	})

}

func TestHandleQuote(t *testing.T) {
	t.Run("it should enable quote status", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("echo ")
		cmd.handleQuote('"')

		assert.True(t, cmd.quotesOpened)
	})

	t.Run("it should disable quote status", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("echo \"Hello")
		cmd.openedQuote = '"'
		cmd.quotesOpened = true
		cmd.handleQuote('"')

		assert.False(t, cmd.quotesOpened)
	})

	t.Run("quote state should stay enabled because of escape state", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("echo \"Hello\\")
		cmd.shouldEscape = true
		cmd.openedQuote = '"'
		cmd.quotesOpened = true
		cmd.handleQuote('"')

		assert.True(t, cmd.quotesOpened)
	})

	t.Run("quote state should stay enabled because of different quote", func(t *testing.T) {
		cmd := newTestCommand(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.setBuffer("echo \"Hello")
		cmd.openedQuote = '"'
		cmd.quotesOpened = true
		cmd.handleQuote('\'')

		assert.True(t, cmd.quotesOpened)
	})
}

func TestHandleBackspace(t *testing.T) {
	
	t.Run("it should delete last char from buffer", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := newTestCommand(&bytes.Buffer{}, output)
		cmd.setBuffer("echo Hello, World")
		cmd.defaultPrint(cmd.buffer)
		cmd.handleBackspace()

		expectedOutput := "echo Hello, World" + string(DELETE)

		assert.Equal(t, cmd.buffer, "echo Hello, Worl")
		assert.Equal(t, expectedOutput, output.String())
	})

	t.Run("it should delete char from current cursor position", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := newTestCommand(&bytes.Buffer{}, output)
		cmd.setBuffer("Ocaml")
		cmd.cursorPos = 1
		cmd.handleBackspace()

		expectedOutput :=  string(DELETE) + "caml"
		for i := 0; i < 4; i++ {
			expectedOutput += (ARROW_CHUNK + string(rune(KeyArrowLeft)))
		}

		assert.Equal(t, cmd.buffer, "caml")
		assert.Equal(t, expectedOutput, output.String())
	})
}
