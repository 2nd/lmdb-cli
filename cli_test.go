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

func (t *CLITests) Each(test func()) {
	t.context = NewTestContext()
	t.recorder = t.context.writer.(*Recorder)
	defer t.context.Close()
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
	t.withinShell(`put test " over\" 9000"`, "get 'test'")
	t.recorder.assert("OK", " over\" 9000")
}

func (t CLITests) IterateNothing() {
	t.withinShell("it")
	t.recorder.assert()
}

func (t CLITests) IteratesAll() {
	t.withinShell("scan iter:", "it", "it", "it", "it")
	t.recorder.assert("iter:0", "value-0", "iter:1", "value-1", "iter:10", "value-10", "iter:11", "value-11", "iter:12", "value-12", "iter:13", "value-13", "iter:14", "value-14", "iter:15", "value-15", "iter:16", "value-16", "iter:17", "value-17", "\"it\" for more", "iter:18", "value-18", "iter:19", "value-19", "iter:2", "value-2", "iter:20", "value-20", "iter:21", "value-21", "iter:22", "value-22", "iter:23", "value-23", "iter:3", "value-3", "iter:4", "value-4", "iter:5", "value-5", "\"it\" for more", "iter:6", "value-6", "iter:7", "value-7", "iter:8", "value-8", "iter:9", "value-9")
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
