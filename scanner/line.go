package scanner

const (
	INIT_POSITION = -2
	EOF = int32(-1)
	RUNE_ERROR = int32(0)
)

//Line represent the input from stdin
type Line struct {
	buffer string
	bufsize int64
	pointer int64
}

//DecreasePointer decrease the line struct pointer
func (line *Line) DecreasePointer() {
	if line.pointer < 0 {
		return
	}

	line.pointer--
}

//NextChar return the next char of the line
//by increasing the pointer by one.
//
//0 is return if the line is empty or the line size is 0. 
//
//If the pointer reach the end of the 
//line, -1 is return as an EOF code.
func (line *Line) NextChar() rune {
	if line.buffer == "" || line.bufsize == 0 {
		return RUNE_ERROR
	}

	if line.pointer >= 0 {
		line.pointer++
	}

	if line.pointer == INIT_POSITION {
		line.pointer = 0
	}

	if line.pointer >= line.bufsize {
		line.pointer = line.bufsize
		return EOF
	}

	return rune(line.buffer[line.pointer])
}

//FurtherChar is similar to NextChar, except that
// it return the next char and doesn't affect the line
// struct pointer.
func (line *Line) FurtherChar() rune {
	if line.buffer == "" || line.bufsize == 0 {
		return RUNE_ERROR
	}

	position := line.pointer

	if position == INIT_POSITION {
		position = 0
	}

	if position >= line.bufsize {
		return EOF
	}

	position++

	return rune(line.buffer[position])
}

//SkipWhiteSpace as it name denote it, skip whitespace
// and tabulation
func (line *Line) SkipWhiteSpace() {
	if line.buffer == "" || line.bufsize == 0 {
		return
	}

	for c := line.FurtherChar(); c != EOF && c == ' ' || c == '\t'; {
		line.NextChar()
	}
}

func CreateLine(text string, pointer int64) Line {
	return Line{
		text,
		int64(len(text)),
		pointer,
	}
}
