package commands

import (
	"git.2nd.io/matt/lmdb-cli/core"
	"github.com/szferi/gomdb"
)

type Put struct {
}

func (cmd Put) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 2, 2)
	if err != nil {
		return err
	}
	err = context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Put(context.DBI, args[0], args[1], 0)
	})
	if err != nil {
		return err
	}
	context.Output(OK)
	return nil
}
