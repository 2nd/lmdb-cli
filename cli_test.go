package lmdbcli

import (
	"bytes"
	"testing"
	"time"

	. "github.com/karlseguin/expect"
	"github.com/szferi/gomdb"
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
	t.withinShell("get over")
	t.recorder.assert("9000!!")
}

func (t CLITests) VerifyExistingKey() {
	t.withinShell("exists over")
	t.recorder.assert("true")
}

func (t CLITests) VerifyMissingKey() {
	t.withinShell("exists nowaythiskeyexists")
	t.recorder.assert("false")
}

func (t CLITests) DeletesAMissingKey() {
	t.withinShell("del nowaythiskeyexists")
	t.recorder.assert("MDB_NOTFOUND: No matching key/data pair found")
}

func (t CLITests) DeletesAKey() {
	t.withinShell("del over", "exists over")
	t.recorder.assert("OK", "false")
}

func (t CLITests) PutsAKey() {
	t.withinShell("put paul atreides")
	t.recorder.assert("OK")
	t.assert("paul", "atreides")
}

func (t CLITests) OverwritesAKey() {
	t.withinShell("put over ninethousand")
	t.recorder.assert("OK")
	t.assert("over", "ninethousand")
}

func (t CLITests) HandlesQuotes() {
	t.withinShell(`put test " over\" 9000`, "get 'test'")
	t.recorder.assert("OK", ` over\" 9000`)
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

func (t CLITests) assert(key string, expected string) {
	t.context.WithinRead(func(txn *mdb.Txn) error {
		actual, err := txn.Get(t.context.dbi, []byte(key))
		if err != nil {
			panic(err)
		}
		Expect(expected).To.Eql(actual)
		return nil
	})
}
