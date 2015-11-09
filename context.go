package lmdbcli

import (
	"io"
	"log"
	"path"

	"github.com/szferi/gomdb"
)

type Context struct {
	*mdb.Env
	dbi      mdb.DBI
	path     string
	prompt   string
	writer   io.Writer
	pathName string
}

func NewContext(dbPath string, size uint64, writer io.Writer) *Context {
	env, _ := mdb.NewEnv()
	env.SetMapSize(size)
	var openFlags uint
	if *roFlag {
		openFlags |= mdb.RDONLY
	}
	if err := env.Open(dbPath, openFlags, 0664); err != nil {
		log.Fatal("failed to open environment: ", err)
	}
	return &Context{
		Env:      env,
		path:     dbPath,
		writer:   writer,
		pathName: path.Base(dbPath),
	}
}

func (c *Context) SwitchDB(name *string) error {
	err := c.WithinRead(func(txn *mdb.Txn) error {
		dbi, err := txn.DBIOpen(name, 0)
		if err != nil {
			return err
		}
		c.dbi = dbi
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

func (c *Context) Write(data []byte) {
	c.writer.Write(data)
	c.writer.Write([]byte{'\n'})
}
