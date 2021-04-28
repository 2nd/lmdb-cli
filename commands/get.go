package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"

	"lmdb-cli/core"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

var (
	GetFormatErr = errors.New("second argument must be 'json' or 'hex'")
)

type Get struct {
}

func (cmd Get) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 1, 2)
	if err != nil {
		return err
	}

	var value []byte
	context.WithinRead(func(txn *lmdb.Txn) error {
		value, err = txn.Get(context.DBI, args[0])
		return nil
	})

	if err != nil {
		return err
	}

	if len(args) == 2 {
		if bytes.Equal(args[1], []byte("json")) {
			var prettyData bytes.Buffer
			if err := json.Indent(&prettyData, value, "", "  "); err != nil {
				return err
			}
			value = prettyData.Bytes()
		} else if bytes.Equal(args[1], []byte("hex")) {
			value = []byte(hex.Dump(value))
		} else {
			return GetFormatErr
		}
	}

	context.Output(value)
	return nil
}
