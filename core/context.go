package core

import (
	"errors"
	"io"
	"log"
	"path"

	"github.com/szferi/gomdb"
)

var (
	NoPromptErr = errors.New("no prompt has been configured")
)

type Prompter interface {
	Prompt(string) (string, error)
	AppendHistory(string)
}

type Context struct {
	*mdb.Env
	DBI      mdb.DBI
	path     string
	prompt   string
	writer   io.Writer
	prompter Prompter
	pathName string
	Cursor   *Cursor
}

type Cursor struct {
	*mdb.Cursor
	txn           *mdb.Txn
	Prefix        []byte
	IncludeValues bool
}

func NewContext(dbPath string, size uint64, ro bool, dbs int, writer io.Writer) *Context {
	env, _ := mdb.NewEnv()
	env.SetMapSize(size)
	var openFlags uint
	if ro {
		openFlags |= mdb.RDONLY
	}

	if dbs > 0 {
		if err := env.SetMaxDBs(mdb.DBI(dbs)); err != nil {
			env.Close()
			log.Fatal("failed to set max dbs", err)
		}
	}

	if err := env.Open(dbPath, openFlags, 0664); err != nil {
		env.Close()
		log.Fatal("failed to open environment: ", err)
	}
	return &Context{
		Env:      env,
		path:     dbPath,
		writer:   writer,
		pathName: path.Base(dbPath),
	}
}

func (c *Context) SetPrompter(Prompter Prompter) {
	c.prompter = Prompter
}

func (c *Context) Prompt() (string, error) {
	if c.prompter == nil {
		return "", NoPromptErr
	}
	input, err := c.prompter.Prompt(c.prompt)
	if err == nil {
		c.prompter.AppendHistory(input)
	}
	return input, err
}

func (c *Context) SwitchDB(name *string) error {
	err := c.WithinWrite(func(txn *mdb.Txn) error {
		dbi, err := txn.DBIOpen(name, mdb.CREATE)
		if err != nil {
			return err
		}
		c.DBI = dbi
		return nil
	})
	if err != nil {
		return err
	}

	var n string
	if name == nil {
		n = "0"
	} else {
		n = *name
	}
	c.prompt = c.pathName + ":" + n + "> "
	return nil
}

func (c *Context) WithinRead(f func(*mdb.Txn) error) error {
	txn, err := c.BeginTxn(nil, mdb.RDONLY)
	if err != nil {
		return err
	}
	defer txn.Commit()
	return f(txn)
}

func (c *Context) WithinWrite(f func(*mdb.Txn) error) error {
	txn, err := c.BeginTxn(nil, 0)
	if err != nil {
		return err
	}
	defer txn.Commit()
	return f(txn)
}

func (c *Context) PrepareCursor(prefix []byte, includeValues bool) error {
	txn, err := c.BeginTxn(nil, mdb.RDONLY)
	if err != nil {
		return err
	}
	cursor, err := txn.CursorOpen(c.DBI)
	if err != nil {
		txn.Abort()
		return err
	}
	c.Cursor = &Cursor{txn: txn, Cursor: cursor, Prefix: prefix, IncludeValues: includeValues}
	return nil
}

func (c *Context) CloseCursor() {
	if c.Cursor != nil {
		c.Cursor.Cursor.Close()
		c.Cursor.txn.Commit()
		c.Cursor = nil
	}
}

func (c *Context) Close() {
	c.CloseCursor()
	c.Env.Close()
}

func (c *Context) Output(data []byte) {
	c.writer.Write(data)
	c.writer.Write([]byte{'\n'})
}

func (c *Context) OutputErr(err error) {
	c.Output([]byte(err.Error()))
}
