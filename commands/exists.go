package commands

import (
	"lmdb-cli/core"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

type Exists struct {
}

func (cmd Exists) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 1)
	if err != nil {
		return err
	}
	err = context.WithinRead(func(txn *lmdb.Txn) error {
		_, err = txn.Get(context.DBI, args[0])
		return err
	})

	if lmdb.IsNotFound(err) {
		context.Output(FALSE)
		return nil
	}
	if err != nil {
		return err
	}
	context.Output(TRUE)
	return nil
}
