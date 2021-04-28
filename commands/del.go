package commands

import (
	"lmdb-cli/core"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

type Del struct {
}

func (cmd Del) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 1)
	if err != nil {
		return err
	}
	err = context.WithinWrite(func(txn *lmdb.Txn) error {
		return txn.Del(context.DBI, args[0], nil)
	})
	if lmdb.IsNotFound(err) {
		context.Output(FALSE)
		return nil
	}
	if err != nil {
		return err
	}
	context.Output(OK)
	return nil
}
