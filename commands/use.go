package commands

import (
	"errors"

	"lmdb-cli/core"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

var (
	DbsFullErr = errors.New("DBs full. Launch with -dbs X to allow X number of databases to be opened")
)

type Use struct {
}

func (cmd Use) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 0, 1)
	if err != nil {
		return err
	}
	var name string
	if len(args) == 1 {
		name = string(args[0])
	}
	err = context.SwitchDB(name)
	if lmdb.IsErrno(err, lmdb.DBsFull) {
		return DbsFullErr
	}
	return err
}
