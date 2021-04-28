package main

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"lmdb-cli/core"

	"github.com/bmatsuo/lmdb-go/lmdb"
	. "github.com/karlseguin/expect"
)

type IntegrationTests struct {
	context  *core.Context
	recorder *Recorder
}

func Test_LmdbCli(t *testing.T) {
	Expectify(new(IntegrationTests), t)
}

func (t *IntegrationTests) Each(test func()) {
	t.context, t.recorder = NewTestContext()
	defer t.context.Close()
	test()
}

func (t IntegrationTests) GetsAnExistingKey() {
	t.withinShell("get over")
	t.recorder.assert("9000!!")
}

func (t IntegrationTests) VerifyExistingKey() {
	t.withinShell("exists over")
	t.recorder.assert("true")
}

func (t IntegrationTests) VerifyMissingKey() {
	t.withinShell("exists nowaythiskeyexists")
	t.recorder.assert("false")
}

func (t IntegrationTests) DeletesAMissingKey() {
	t.withinShell("del nowaythiskeyexists")
	t.recorder.assert("false")
}

func (t IntegrationTests) DeletesAKey() {
	t.withinShell("del over", "exists over")
	t.recorder.assert("ok", "false")
}

func (t IntegrationTests) PutsAKey() {
	t.withinShell("put paul atreides")
	t.recorder.assert("ok")
	t.assert("paul", "atreides")
}

func (t IntegrationTests) OverwritesAKey() {
	t.withinShell("put over ninethousand")
	t.recorder.assert("ok")
	t.assert("over", "ninethousand")
}

func (t IntegrationTests) HandlesQuotes() {
	t.withinShell(`put test " over\" 9000"`, "get 'test'")
	t.recorder.assert("ok", " over\" 9000")
}

func (t IntegrationTests) IterateNothing() {
	t.withinShell("it")
	t.recorder.assert()
}

func (t IntegrationTests) IteratesAll() {
	t.withinShell("scan iter:", "it", "it", "it", "it")
	t.recorder.assert("iter:0", "value-0", "", "iter:1", "value-1", "", "iter:10", "value-10", "", "iter:11", "value-11", "", "iter:12", "value-12", "", "iter:13", "value-13", "", "iter:14", "value-14", "", "iter:15", "value-15", "", "iter:16", "value-16", "", "iter:17", "value-17", "", "\"it\" for more", "iter:18", "value-18", "", "iter:19", "value-19", "", "iter:2", "value-2", "", "iter:20", "value-20", "", "iter:21", "value-21", "", "iter:22", "value-22", "", "iter:23", "value-23", "", "iter:3", "value-3", "", "iter:4", "value-4", "", "iter:5", "value-5", "", "\"it\" for more", "iter:6", "value-6", "", "iter:7", "value-7", "", "iter:8", "value-8", "", "iter:9", "value-9", "")
}

func (t IntegrationTests) Keys() {
	t.withinShell("keys iter:1", "it", "it", "it", "it")
	t.recorder.assert("iter:1", "iter:10", "iter:11", "iter:12", "iter:13", "iter:14", "iter:15", "iter:16", "iter:17", "iter:18", "\"it\" for more", "iter:19")
}

func (t IntegrationTests) Stats() {
	t.withinShell("stats")
	t.recorder.assert("map size: 4194304", "map size (human): 4MB", "num entries: 25", "max readers: 126", "num readers: 0", "db page size: 4096", "non-leaf pages: 0", "leaf pages: 1", "overflow pages: 0", "last page id: 7", "map tx id: 25")
}

func (t IntegrationTests) UseErrorIfNoSize() {
	t.withinShell("use leto", "use paul")
	t.recorder.assert("DBs full. Launch with -dbs X to allow X number of databases to be opened")
}

func (t IntegrationTests) UsesDifferentDatabase() {
	t.withinShell("use leto", "set spice flow", "use", "exists spice", "use leto", "exists spice")
	t.recorder.assert("ok", "false", "true")
}

func (t IntegrationTests) withinShell(commands ...string) {
	t.context.SetPrompter(NewMockPrompter(commands...))
	runShell(t.context)
}

func (t IntegrationTests) assert(key string, expected string) {
	t.context.WithinRead(func(txn *lmdb.Txn) error {
		actual, err := txn.Get(t.context.DBI, []byte(key))
		if err != nil {
			panic(err)
		}
		Expect(expected).To.Eql(actual)
		return nil
	})
}

func NewTestContext() (*core.Context, *Recorder) {
	root, _ := os.Getwd()
	root = path.Join(root, "test")
	dbPath := path.Join(root, "sample")
	os.RemoveAll(dbPath)
	if err := exec.Command("cp", "-r", path.Join(root, "template"), dbPath).Run(); err != nil {
		panic(err)
	}
	recorder := NewRecorder()
	c := core.NewContext(dbPath, 4194304, false, 1, recorder)
	if err := c.SwitchDB(""); err != nil {
		c.Close()
		panic(err)
	}
	return c, recorder
}

type MockPrompter struct {
	c chan string
}

func NewMockPrompter(commands ...string) core.Prompter {
	c := make(chan string, len(commands)+1)
	for _, command := range commands {
		c <- command + "\n"
	}
	c <- "exit\n"
	return &MockPrompter{c}
}

func (m *MockPrompter) AppendHistory(line string) {

}

func (m *MockPrompter) Prompt(p string) (string, error) {
	return <-m.c, nil
}

type Recorder struct {
	values []string
}

func NewRecorder() *Recorder {
	return &Recorder{values: make([]string, 0, 5)}
}

func (r *Recorder) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == '\n' {
		//don't bother writing command output separators
		return 1, nil
	}
	r.values = append(r.values, string(b))
	return len(b), nil
}

func (r *Recorder) assert(values ...string) {
	Expect(len(values)).To.Equal(len(r.values))
	for i, expected := range values {
		Expect(expected).To.Equal(r.values[i])
	}
}
