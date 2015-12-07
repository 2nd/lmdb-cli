package commands

import (
	"errors"

	"git.2nd.io/matt/lmdb-cli/core"
	"github.com/szferi/gomdb"
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
	var name *string
	if len(args) == 1 {
		n := string(args[0])
		name = &n
	}
	err = context.SwitchDB(name)
	if err == mdb.DbsFull {
		return DbsFullErr
	}
	return err
}
