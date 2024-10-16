package main

import (
	"bytes"
	"testing"
)

func TestReadFromSource(t *testing.T) {
	t.Run("it read single line command from source", func(t *testing.T) {
		buf := bytes.Buffer{}
		buf.WriteString("echo \"Hello, world\"")

		expected := "echo \"Hello, world\""
		cmd, _ := readFromSource(&buf)

		if expected != cmd {
			t.Errorf("Expected %s, but got %s\n", expected, cmd)
		}
	})

	t.Run("it read multi lines command from a source", func(t *testing.T) {
		buf := bytes.Buffer{}
		buf.WriteString("echo \"This is where all start.\\\n I named this shell cish, but I could rename it somthing else.\\\n Maybe cheh, I don't know.\"")

		expected := "echo \"This is where all start. I named this shell cish, but I could rename it somthing else. Maybe cheh, I don't know.\""
		cmd, _ := readFromSource(&buf)

		if expected != cmd {
			t.Errorf("Expected %s, but got %s\n", expected, cmd)
		}
	})
}
