package commands

import "lmdb-cli/core"

var helpText = []byte(`
  del KEY        - removes the key/value
  get KEY FORMAT - gets the value. FORMAT is optional or 'json' or 'hex'
  set KEY        - creates or overwrites the key with the specified value
                   (aliases: put)
  exists KEY     - checks if the key exists

  scan PREFIX    - returns key & values where keys match the optional prefix
  keys PREFIX    - returns keys where keys match the optional prefix
  it             - goes to the next page of a scan/keys result

  info           - returns information on the database
                   (aliases: stat, stats)
  use DB         - switches to a named database. If DB is omitted, switches back
                   to the default database.

  exit           - exits the program
                  (aliases: quit, CTRL-C)
`)

type Help struct {
}

func (cmd Help) Execute(context *core.Context, input []byte) (err error) {
	context.Output(helpText)
	return nil
}
