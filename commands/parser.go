package commands

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strconv"
	"unicode"
)

const (
	STATE_NONE int = iota
	STATE_WORD
	STATE_QUOTE
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

	pushArg := func() {
		if state == STATE_WORD && bytes.HasPrefix(arg, []byte("0x")) {
			if n, err := hex.Decode(arg[2:], arg[2:]); err == nil {
				arg = arg[2 : 2+n]
			}
		}

		results = append(results, arg)
	}

	for _, b := range in {
		switch state {
		case STATE_NONE:
			if isQuote(b) {
				state = STATE_QUOTE
				arg = append(arg, '"')
			} else if !isWhiteSpace(b) {
				arg = append(arg, b)
				state = STATE_WORD
			}
		case STATE_WORD:
			if isWhiteSpace(b) {
				pushArg()
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		case STATE_QUOTE:
			if isQuote(b) {
				arg = append(arg, '"')
				unquoted, err := strconv.Unquote(string(arg))
				if err == nil {
					arg = []byte(unquoted)
				}
				pushArg()
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		}
	}
	if len(arg) != 0 {
		pushArg()
	}
	return results
}

func isWhiteSpace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func isQuote(b byte) bool {
	return b == '"' || b == '\'' || b == '`'
}
