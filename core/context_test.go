package core

import (
	"bytes"
	"testing"

	. "github.com/karlseguin/expect"
)

type ContextTests struct{}

func Test_Context(t *testing.T) {
	Expectify(new(ContextTests), t)
}

func (_ ContextTests) WriteAppendsNewLine() {
	buffer := new(bytes.Buffer)
	c := &Context{writer: buffer}
	c.Output([]byte("over 9000!"))
	Expect(buffer.String()).To.Equal("over 9000!\n")
}
