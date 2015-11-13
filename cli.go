// a command line interface to lmdb
package lmdbcli

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"path"

	"git.2nd.io/matt/lmdb-cli/commands"
	"git.2nd.io/matt/lmdb-cli/core"
)

var (
	pathFlag = flag.String("db", "", "Relative path to lmdb file")
	sizeFlag = flag.Float64("size", 2, "factor to allocate for growth or shrinkage")
	roFlag   = flag.Bool("ro", false, "open the database in read-only mode")

	cmds  = make(map[string]Command)
	units = []string{"KB", "MB", "GB", "TB", "PB"}

	OK              = []byte("OK")
	SCAN_MORE       = []byte(`"it" for more`)
	INVALID_COMMAND = []byte("invalid command")
)

type Command interface {
	Execute(context *core.Context, arguments []byte) error
}

func init() {
	cmds["get"] = commands.Get{}
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

	context := core.NewContext(*pathFlag, size, *roFlag, os.Stdout)
	defer context.Close()
	if err := context.SwitchDB(nil); err != nil {
		log.Fatal("could not select default database: ", err)
	}
	runShell(context, os.Stdin)
}

func runShell(context *core.Context, in io.Reader) {
	reader := bufio.NewReader(in)
	for {
		context.Prompt()
		var arguments []byte
		input, _ := reader.ReadSlice('\n')
		input = bytes.TrimSpace(input)

		if index := bytes.IndexByte(input, ' '); index != -1 {
			arguments = input[index+1:]
			input = input[:index]
		}

		cmd := cmds[string(input)]
		if cmd == nil {
			context.Output(INVALID_COMMAND)
			continue
		}
		if err := cmd.Execute(context, arguments); err != nil {
			context.OutputErr(err)
		}
	}
}
