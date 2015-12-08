package commands

import (
	"github.com/2nd/lmdb-cli/core"
	"github.com/szferi/gomdb"
)

type Del struct {
}

func (cmd Del) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 1)
	if err != nil {
		return err
	}
	err = context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Del(context.DBI, args[0], nil)
	})
	if err == mdb.NotFound {
		context.Output(FALSE)
		return nil
	}
	if err != nil {
		return err
	}
	context.Output(OK)
	return nil
}
