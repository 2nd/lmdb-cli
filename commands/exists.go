package commands

import (
	"git.2nd.io/matt/lmdb-cli/core"
	"github.com/szferi/gomdb"
)

type Exists struct {
}

func (cmd Exists) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 1)
	if err != nil {
		return err
	}
	err = context.WithinRead(func(txn *mdb.Txn) error {
		_, err = txn.Get(context.DBI, args[0])
		return err
	})

	if err == mdb.NotFound {
		context.Output(FALSE)
		return nil
	}
	if err != nil {
		return err
	}
	context.Output(TRUE)
	return nil
}
