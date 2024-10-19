package scanner

import (
	"slices"
	"strings"
)

type Token struct {
	text string
	Len int
	isEndOfLine bool
}

//Append appends a new char to the token
func (token *Token) Append(char rune) {
	token.text += string(char)
	token.Len++
}

//Tokenize create tokens from a line struct
func Tokenize(line *Line) []Token {
	var tokens []Token

	for {
		c := line.NextChar()
		if c == EOF || c == RUNE_ERROR {
			break;
		}
		token := CreateToken(line)
		tokens = append(tokens, token)
	}

	tokens = append(tokens, Token{
		"",
		0,
		true,
	})

	return tokens
}


func CreateToken(line *Line) (token Token) {

	for {
		c := line.NextChar()

		if c == EOF {
			break
		}

		if slices.Contains([]rune{'\'', '"'}, c) {
			if token.Len == 0 {
				token.Append(c)
				continue
			}

			if (c == '"' && strings.HasPrefix(token.text, "\"")) ||
			   (c == '\'' && strings.HasPrefix(token.text, "'")) {
				token.Append(c)
				if line.FurtherChar() == ' ' {
					break
				}
			}
		}

		token.Append(c)
	}

	return
}

/* func handleQuote(quote rune, nextChar rune, token *Token) int {
	
	if token.Len == 0 {
		token.Append(quote)
		return 0
	}

	if strings.HasPrefix(token.text, string(quote)) && nextChar == ' ' {
		token.Append(quote)
		return -1
	}

	return 0
} */
