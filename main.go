// a command line interface to lmdb
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path"
	"strings"

	"github.com/peterh/liner"

	"lmdb-cli/commands"
	"lmdb-cli/core"
)

var (
	pathFlag    = flag.String("db", "", "Relative path to lmdb file")
	nameFlag    = flag.String("name", "", "database name to open")
	sizeFlag    = flag.Int("size", 32*1024*1024, "size in bytes to allocate for new database")
	growthFlag  = flag.Float64("growth", 1, "factor to grow/shrink an existing database")
	roFlag      = flag.Bool("ro", false, "open the database in read-only mode")
	dir         = flag.Bool("dir", true, "path given is the directory (when false, uses MDB_NOSUBDIR)")
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
	cmds["keys"] = commands.Keys{}
	cmds["help"] = commands.Help{}
	cmds["ascii"] = commands.Ascii{}
}

func main() {
	flag.Parse()

	if len(*pathFlag) == 0 && len(flag.Args()) == 1 {
		pathFlag = &flag.Args()[0]
	}
	if len(*pathFlag) == 0 {
		log.Fatal("-db must be specified")
	}

	dataPath := *pathFlag
	size := int64(*sizeFlag)

	if *dir {
		dataPath = path.Join(dataPath, "data.mdb")
	}
	if stat, err := os.Stat(dataPath); err != nil {
		if os.IsNotExist(err) == false {
			log.Fatal("failed to stat data.mdb file: ", err)
		}
		if *dir {
			if err := os.Mkdir(*pathFlag, 0744); err != nil {
				log.Fatal("failed to make directory", err)
			}
		}
	} else {
		size = int64(float64(stat.Size()) * *growthFlag)
	}
	runOne := len(*commandFlag) != 0

	context := core.NewContext(*pathFlag, size, *roFlag, *dir, *dbsFlag, os.Stdout)
	defer context.Close()
	if err := context.SwitchDB(*nameFlag); err != nil {
		log.Fatal("could not select default database: ", err)
	}
	if runOne {
		process(context, []byte(*commandFlag))
		return
	}

	cmds["ascii"].Execute(context, nil)
	context.Output([]byte("stats>"))
	cmds["stats"].Execute(context, nil)
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) (c []string) {
		line = strings.ToLower(line)
		for cmd := range cmds {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		return c
	})

	context.SetPrompter(line)
	runShell(context)
}

func runShell(context *core.Context) {
	for {
		input, err := context.Prompt()
		if err != nil {
			if err == liner.ErrPromptAborted {
				break
			}
			context.OutputErr(err)
		} else if process(context, []byte(input)) == false {
			break
		}
	}
}

func process(context *core.Context, input []byte) bool {
	if len(input) == 0 {
		return true
	}

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
