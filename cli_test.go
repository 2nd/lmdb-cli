package lmdbcli

import (
	"testing"

	. "github.com/karlseguin/expect"
)

type CLITests struct {
	context  *Context
	recorder *Recorder
}

func Test_CLI(t *testing.T) {
	Expectify(new(CLITests), t)
}

func (c *CLITests) Each(test func()) {
	c.context = NewTestContext()
	c.recorder = c.context.writer.(*Recorder)
	defer c.context.Close()
	test()
}

func (t CLITests) GetsAnExistingKey() {
	get(t.context, "over")
	t.recorder.assert("9000!!")
}
