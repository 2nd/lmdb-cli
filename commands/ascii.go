package commands

import "lmdb-cli/core"

var ascii = []byte(`____________  __________________       ______________________
___  /___   |/  /__  __ \__  __ )      __  ____/__  /____  _/
__  / __  /|_/ /__  / / /_  __  |_______  /    __  /  __  /
_  /___  /  / / _  /_/ /_  /_/ /_/_____/ /___  _  /____/ /
/_____/_/  /_/  /_____/ /_____/        \____/  /_____/___/
`)

type Ascii struct {
}

func (cmd Ascii) Execute(context *core.Context, input []byte) (err error) {
	context.Output(ascii)
	return nil
}
