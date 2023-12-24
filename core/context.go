package core

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"path"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

var NoPromptErr = errors.New("no prompt has been configured")

type Prompter interface {
	Prompt(string) (string, error)
	AppendHistory(string)
}

type Context struct {
	*lmdb.Env
	DBI      lmdb.DBI
	path     string
	prompt   string
	writer   io.Writer
	prompter Prompter
	pathName string
	Cursor   *Cursor
}

type Cursor struct {
	*lmdb.Cursor
	txn           *lmdb.Txn
	Prefix        []byte
	IncludeValues bool
}

func NewContext(dbPath string, size int64, ro bool, dir bool, dbs int, writer io.Writer) *Context {
	env, err := lmdb.NewEnv()
	if err != nil {
		log.Fatal("failed to create env", err)
	}
	env.SetMapSize(size)
	var openFlags uint
	if ro {
		openFlags |= lmdb.Readonly
	}
	if !dir {
		openFlags |= lmdb.NoSubdir
	}

	if dbs > 0 {
		if err := env.SetMaxDBs(dbs); err != nil {
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

func (c *Context) SwitchDB(name string) error {
	err := c.WithinWrite(func(txn *lmdb.Txn) error {
		var dbi lmdb.DBI
		var err error
		if len(name) == 0 {
			dbi, err = txn.OpenRoot(lmdb.Create)
		} else {
			dbi, err = txn.OpenDBI(name, lmdb.Create)
		}

		if err != nil {
			return err
		}
		c.DBI = dbi
		return nil
	})
	if err != nil {
		return err
	}

	n := name
	if len(n) == 0 {
		n = "0"
	}
	c.prompt = c.pathName + ":" + n + "> "
	return nil
}

func (c *Context) WithinRead(f func(*lmdb.Txn) error) error {
	txn, err := c.BeginTxn(nil, lmdb.Readonly)
	if err != nil {
		return err
	}
	defer txn.Commit()
	return f(txn)
}

func (c *Context) WithinWrite(f func(*lmdb.Txn) error) error {
	txn, err := c.BeginTxn(nil, 0)
	if err != nil {
		return err
	}
	defer txn.Commit()
	return f(txn)
}

func (c *Context) PrepareCursor(prefix []byte, includeValues bool) error {
	txn, err := c.BeginTxn(nil, lmdb.Readonly)
	if err != nil {
		return err
	}
	cursor, err := txn.OpenCursor(c.DBI)
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
	n := len(data)
	readableCharacters := 0
	for _, b := range data {
		if b >= 32 && b <= 126 || b == 10 {
			readableCharacters++
		}
	}
	if readableCharacters > n*2/3 {
		c.writer.Write(data)
	} else {
		c.writer.Write([]byte("0x" + hex.EncodeToString(data)))
	}

	c.writer.Write([]byte{'\n'})
}

func (c *Context) OutputErr(err error) {
	c.Output([]byte(err.Error()))
}
