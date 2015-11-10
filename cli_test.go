package lmdbcli

import (
	"bytes"
	"testing"
	"time"

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

func (t CLITests) VerifyExistingKey() {
	exists(t.context, "over")
	t.recorder.assert("true")
}

func (t CLITests) VerifyMissingKey() {
	exists(t.context, "nowaythiskeyexists")
	t.recorder.assert("false")
}

func (t CLITests) Exits() {
	for _, input := range []string{"exit", "quit"} {
		done := false
		in := bytes.NewBufferString(input + "\n")
		go func() {
			runShell(t.context, in)
			done = true
		}()
		time.Sleep(time.Millisecond * 5)
		Expect(done).To.Equal(true)
	}
}
