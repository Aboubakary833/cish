package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecreasePointer(t *testing.T) {
	input := "echo 'Hello, world!'"
	line := Line{
		input,
		int64(len(input)),
		2,
	}

	line.DecreasePointer()
	got := line.pointer

	assert.Equal(t, int64(1), got)
}

func TestNextChar(t *testing.T) {
	t.Run("It should return l", func(t *testing.T) {
		input := "ls"
		line := Line{
			input,
			2,
			INIT_POSITION,
		}

		expected := 'l'
		got := line.NextChar()

		assert.Equal(t, expected, got)
	})

	t.Run("It should return s", func(t *testing.T) {
		input := "ls -a"
		line := Line{
			input,
			5,
			0,
		}

		expected := 's'
		got := line.NextChar()

		assert.Equal(t, expected, got, "Expected %q, but got %q", expected, got)
	})

	t.Run("It should return EOF code", func(t *testing.T) {
		input := "ls -l"
		line := Line{
			input,
			5,
			5,
		}

		got := line.NextChar()

		assert.Equal(t, EOF, got)
	})

	t.Run("It should return 0", func(t *testing.T) {
		line := Line{
			"",
			0,
			INIT_POSITION,
		}

		got := line.NextChar()

		assert.Equal(t, RUNE_ERROR, got)
	})
}

func TestFurtherChar(t *testing.T) {
  t.Run("It should return s", func(t *testing.T) {
    input := "ls -la"
    line := Line{
      input,
      6,
      INIT_POSITION,
    }

    expected := 's'
    got := line.FurtherChar()
	
	assert.Equal(t, expected, got, "Expected %q, but got %q", expected, got)
  })

  t.Run("It should return EOF code", func(t *testing.T) {
	line := Line{
		"cd",
		2,
		2,
	}

	got := line.FurtherChar()

	assert.Equal(t, EOF, got)
  })

  t.Run("It should return RUNE ERROR code", func(t *testing.T) {
	input := ""
	line := Line{
		input,
		int64(len(input)),
		INIT_POSITION,
	}

	got := line.FurtherChar()

	assert.Equal(t, RUNE_ERROR, got)
  })
}
