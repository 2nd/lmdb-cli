// a command line interface to lmdb
package lmdbcli

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"git.2nd.io/matt/lmdb-cli/commands"
	"git.2nd.io/matt/lmdb-cli/core"
)

var (
	pathFlag    = flag.String("db", "", "Relative path to lmdb file")
	sizeFlag    = flag.Float64("size", 2, "factor to allocate for growth or shrinkage")
	roFlag      = flag.Bool("ro", false, "open the database in read-only mode")
	dbsFlag     = flag.Int("dbs", 0, "number of additional databases to allow")
	commandFlag = flag.String("c", "", "command to run")

	cmds = make(map[string]Command)

	INVALID_COMMAND = []byte("invalid command")
)

type Command interface {
	Execute(context *core.Context, arguments []byte) error
}

func init() {
	cmds["del"] = commands.Del{}
	cmds["exists"] = commands.Exists{}
	cmds["get"] = commands.Get{}
	cmds["info"] = commands.Stats{}
	cmds["it"] = commands.Iterate{}
	cmds["put"] = commands.Put{}
	cmds["scan"] = commands.Scan{}
	cmds["set"] = commands.Put{}
	cmds["stat"] = commands.Stats{}
	cmds["stats"] = commands.Stats{}
	cmds["use"] = commands.Use{}
}

func Run() {
	flag.Parse()

	if len(*pathFlag) == 0 && len(flag.Args()) == 1 {
		pathFlag = &flag.Args()[0]
	}
	if len(*pathFlag) == 0 {
		log.Fatal("-db must be specified")
	}

	size := uint64(1024 * 1024 * 32)
	if stat, err := os.Stat(path.Join(*pathFlag, "data.mdb")); err != nil {
		if os.IsNotExist(err) == false {
			log.Fatal("failed to stat data.mdb file: ", err)
		}
	} else {
		size = uint64(float64(stat.Size()) * *sizeFlag)
	}
	runOne := len(*commandFlag) != 0

	var promptWriter io.Writer = os.Stdout
	if runOne {
		promptWriter = ioutil.Discard
	}

	context := core.NewContext(*pathFlag, size, *roFlag, *dbsFlag, os.Stdout, promptWriter)
	defer context.Close()
	if err := context.SwitchDB(nil); err != nil {
		log.Fatal("could not select default database: ", err)
	}
	if runOne {
		process(context, []byte(*commandFlag))
	} else {
		runShell(context, os.Stdin)
	}
}

func runShell(context *core.Context, in io.Reader) {
	reader := bufio.NewReader(in)
	for {
		context.Prompt()
		input, _ := reader.ReadSlice('\n')
		if process(context, input) == false {
			break
		}
	}
}

func process(context *core.Context, input []byte) bool {

	var arguments []byte
	input = bytes.TrimSpace(input)

	if index := bytes.IndexByte(input, ' '); index != -1 {
		arguments = input[index+1:]
		input = input[:index]
	}

	if bytes.Equal(input, []byte("exit")) || bytes.Equal(input, []byte("quit")) {
		return false
	}

	if bytes.Equal(input, []byte("it")) == false {
		context.CloseCursor()
	}

	cmd := cmds[string(input)]
	if cmd == nil {
		context.Output(INVALID_COMMAND)
	} else if err := cmd.Execute(context, arguments); err != nil {
		context.OutputErr(err)
	}
	return true
}
