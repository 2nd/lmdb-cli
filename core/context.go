package core

import (
	"io"
	"log"
	"path"

	"github.com/szferi/gomdb"
)

type Context struct {
	*mdb.Env
	DBI          mdb.DBI
	path         string
	prompt       []byte
	writer       io.Writer
	promptWriter io.Writer
	pathName     string
	cursor       *Cursor
}

type Cursor struct {
	*mdb.Cursor
	txn    *mdb.Txn
	prefix []byte
}

func NewContext(dbPath string, size uint64, ro bool, writer io.Writer) *Context {
	env, _ := mdb.NewEnv()
	env.SetMapSize(size)
	var openFlags uint
	if ro {
		openFlags |= mdb.RDONLY
	}
	if err := env.Open(dbPath, openFlags, 0664); err != nil {
		log.Fatal("failed to open environment: ", err)
	}
	return &Context{
		Env:          env,
		path:         dbPath,
		writer:       writer,
		promptWriter: writer,
		pathName:     path.Base(dbPath),
	}
}

func (c *Context) Prompt() {
	c.promptWriter.Write(c.prompt)
}

func (c *Context) SwitchDB(name *string) error {
	err := c.WithinRead(func(txn *mdb.Txn) error {
		dbi, err := txn.DBIOpen(name, 0)
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
	c.prompt = []byte(c.pathName + ":" + n + "> ")
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

func (c *Context) PrepareCursor(prefix []byte) error {
	txn, err := c.BeginTxn(nil, mdb.RDONLY)
	if err != nil {
		return err
	}
	cursor, err := txn.CursorOpen(c.DBI)
	if err != nil {
		txn.Abort()
		return err
	}
	c.cursor = &Cursor{txn: txn, Cursor: cursor, prefix: prefix}
	return nil
}

func (c *Context) CloseCursor() {
	if c.cursor != nil {
		c.cursor.Cursor.Close()
		c.cursor.txn.Commit()
		c.cursor = nil
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
