package commands

import (
	"bytes"
	"encoding/json"
	"errors"

	"git.2nd.io/matt/lmdb-cli/core"
	"github.com/szferi/gomdb"
)

var (
	GetFormatErr = errors.New("second argument must be 'json'")
)

type Get struct {
}

func (cmd Get) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 2)
	if err != nil {
		return err
	}
	if len(args) == 2 && bytes.Equal(args[1], []byte("json")) == false {
		return GetFormatErr
	}

	var value []byte
	context.WithinRead(func(txn *mdb.Txn) error {
		value, err = txn.Get(context.DBI, args[0])
		return nil
	})

	if err != nil {
		return err
	}

	if len(args) == 2 {
		var prettyData bytes.Buffer
		if err := json.Indent(&prettyData, value, "", "  "); err != nil {
			return err
		}
		value = prettyData.Bytes()
	}
	context.Output(value)
	return nil
}
