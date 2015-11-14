package commands

import (
	"bytes"

	"git.2nd.io/matt/lmdb-cli/core"
	"github.com/szferi/gomdb"
)

type Iterate struct {
}

func (cmd Iterate) Execute(context *core.Context, input []byte) (err error) {
	return cmd.execute(context, false)
}

func (cmd Iterate) execute(context *core.Context, first bool) (err error) {
	cursor := context.Cursor
	if cursor == nil {
		return nil
	}
	for i := 0; i < 10; i++ {
		var err error
		var key, value []byte
		if first && cursor.Prefix != nil {
			key, value, err = cursor.Get(cursor.Prefix, nil, mdb.SET_RANGE)
			first = false
		} else {
			key, value, err = cursor.Get(nil, nil, mdb.NEXT)
		}

		if err == mdb.NotFound || (cursor.Prefix != nil && !bytes.HasPrefix(key, cursor.Prefix)) {
			context.CloseCursor()
			return nil
		}
		if err != nil {
			context.CloseCursor()
			return err
		}
		context.Output(key)
		context.Output(value)
	}
	context.Output(SCAN_MORE)
	return nil
}
