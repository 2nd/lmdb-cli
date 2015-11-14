package commands

import (
	"errors"
	"unicode"
)

const (
	STATE_NONE int = iota
	STATE_WORD
	STATE_QUOTE
	STATE_ESCAPED
)

var (
	NotEnoughArgumentsErr = errors.New("not enough arguments")
	TooManyArgumentsErr   = errors.New("too many arguments")
)

func parseRange(in []byte, min int, max int) ([][]byte, error) {
	args := parse(in)
	if len(args) < min {
		return nil, NotEnoughArgumentsErr
	}
	if len(args) > max {
		return nil, TooManyArgumentsErr
	}
	return args, nil
}

// handle both space delimiters and arguments in quotations
// arguments are defined as contained by spaces ' arg ' or quotations '"arg"'
// forward slash escapes for nested quotations
func parse(in []byte) [][]byte {
	var results [][]byte
	var arg []byte
	state := STATE_NONE
	for _, b := range in {
		switch state {
		case STATE_NONE:
			if isQuote(b) {
				state = STATE_QUOTE
			} else if !isWhiteSpace(b) {
				arg = append(arg, b)
				state = STATE_WORD
			}
		case STATE_ESCAPED:
			arg = append(arg, b)
			state = STATE_QUOTE
		case STATE_WORD:
			if isWhiteSpace(b) {
				results = append(results, arg)
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		case STATE_QUOTE:
			if b == '\\' {
				state = STATE_ESCAPED
			} else if isQuote(b) {
				results = append(results, arg)
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		}
	}
	if len(arg) != 0 {
		results = append(results, arg)
	}
	return results
}

func isWhiteSpace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func isQuote(b byte) bool {
	return b == '"' || b == '\'' || b == '`'
}
