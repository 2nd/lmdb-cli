package commands

import "lmdb-cli/core"

type Stats struct {
}

func (cmd Stats) Execute(context *core.Context, input []byte) (err error) {
	_, err = parseRange(input, 0, 0)
	if err != nil {
		return err
	}
	info, err := context.Info()
	if err != nil {
		return err
	}
	stats, err := context.Stat()
	if err != nil {
		return err
	}
	context.Output(labelInt("map size", info.MapSize))
	if readable := readableBytes(info.MapSize); len(readable) != 0 {
		context.Output(labelString("map size (human)", readable))
	}
	context.Output(labelUint("num entries", stats.Entries))
	context.Output(labelUint("max readers", uint64(info.MaxReaders)))
	context.Output(labelUint("num readers", uint64(info.NumReaders)))

	context.Output(labelUint("db page size", uint64(stats.PSize)))
	context.Output(labelUint("non-leaf pages", stats.BranchPages))
	context.Output(labelUint("leaf pages", stats.LeafPages))
	context.Output(labelUint("overflow pages", stats.OverflowPages))
	context.Output(labelInt("last page id", info.LastPNO))
	context.Output(labelInt("map tx id", info.LastTxnID))
	return nil
}
