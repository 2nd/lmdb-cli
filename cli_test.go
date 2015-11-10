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

func (t CLITests) DeletesAMissingKey() {
	t.withinShell("del nowaythiskeyexists")
	t.recorder.assert("MDB_NOTFOUND: No matching key/data pair found")
}

func (t CLITests) DeletesAKey() {
	del(t.context, "over")
	exists(t.context, "over")
	t.recorder.assert("ok", "false")
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

func (t CLITests) withinShell(commands ...string) {
	in := new(bytes.Buffer)
	go func() {
		runShell(t.context, in)
	}()
	for _, command := range commands {
		in.WriteString(command + "\n")
	}
	in.WriteString("exit\n")
	time.Sleep(time.Millisecond * 5)
}
