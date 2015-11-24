package commands

import "git.2nd.io/matt/lmdb-cli/core"

type Use struct {
}

func (cmd Use) Execute(context *core.Context, input []byte) (err error) {
	args, err := parseRange(input, 0, 1)
	if err != nil {
		return err
	}
	var name *string
	if len(args) == 1 {
		n := string(args[0])
		name = &n
	}
	return context.SwitchDB(name)
}
