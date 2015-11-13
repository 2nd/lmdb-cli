package core

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

func NewTestContext() *Context {
	root, _ := os.Getwd()
	root = path.Join(root, "test")
	dbPath := path.Join(root, "sample")
	os.RemoveAll(dbPath)
	if err := exec.Command("cp", "-r", path.Join(root, "template"), dbPath).Run(); err != nil {
		panic(err)
	}
	c := NewContext(dbPath, 4194304, false, NewRecorder())
	c.promptWriter = ioutil.Discard
	if err := c.SwitchDB(nil); err != nil {
		c.Close()
		panic(err)
	}
	return c
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
