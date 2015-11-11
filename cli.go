package lmdbcli

import (
	"bufio"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"unicode"

	"github.com/szferi/gomdb"
)

var (
	pathFlag = flag.String("db", "", "Relative path to lmdb file")
	sizeFlag = flag.Float64("size", 2, "factor to allocate for growth or shrinkage")
	roFlag   = flag.Bool("ro", false, "open the database in read-only mode")
	minArgs  = map[string]int{"scan": 0, "stat": 0, "expand": 0, "exists": 1, "get": 1, "del": 1, "put": 2, "exit": 0, "quit": 0}
)

const (
	STATE_NONE int = iota
	STATE_WORD
	STATE_QUOTE
	STATE_ESCAPED
)

type Command struct {
	fn   string
	key  []byte
	val  []byte
	args [][]byte
}

// Run golmdb using the directory containing the data as dbPath

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

	context := NewContext(*pathFlag, size, os.Stdout)
	defer context.Close()
	if err := context.SwitchDB(nil); err != nil {
		log.Fatal("could not select default database: ", err)
	}
	runShell(context, os.Stdin)
}

func runShell(context *Context, in io.Reader) {
	var err error
	reader := bufio.NewReader(in)
	for {
		context.Prompt()
		input, _ := reader.ReadSlice('\n')

		args := parseInput(input)
		if cmd, err1 := getCommand(args); err1 != nil {
			context.Write([]byte(err1.Error()))
		} else if cmd.fn == "get" {
			err = get(context, cmd.key)
		} else if cmd.fn == "exists" {
			err = exists(context, cmd.key)
		} else if cmd.fn == "del" {
			err = del(context, cmd.key)
		} else if cmd.fn == "put" {
			err = put(context, cmd.key, cmd.val)
		} else if cmd.fn == "scan" {
			err = scan(context)
		} else if cmd.fn == "quit" || cmd.fn == "exit" {
			return
		}
		if err != nil {
			context.Write([]byte(err.Error()))
		}
	}
}

func get(context *Context, key []byte) error {
	return context.WithinRead(func(txn *mdb.Txn) error {
		data, err := txn.Get(context.dbi, key)
		if err != nil {
			return err
		}
		context.Write(data)
		return nil
	})
}

func exists(context *Context, key []byte) error {
	return context.WithinRead(func(txn *mdb.Txn) error {
		_, err := txn.Get(context.dbi, key)
		if err != nil {
			context.Write([]byte("false"))
		} else {
			context.Write([]byte("true"))
		}
		return nil
	})
}

func del(context *Context, key []byte) error {
	err := context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Del(context.dbi, key, nil)
	})
	if err != nil {
		return err
	}
	context.Write([]byte("ok"))
	return nil
}

func put(context *Context, key, val []byte) error {
	return context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Put(context.dbi, key, val, 0)
	})
}

func scan(context *Context) error {
	return context.WithinRead(func(txn *mdb.Txn) error {
		cursor, err := txn.CursorOpen(context.dbi)
		if err != nil {
			return err
		}
		defer cursor.Close()
		for {
			key, val, err := cursor.Get(nil, nil, mdb.NEXT)
			if err == mdb.NotFound {
				return nil
			}
			if err != nil {
				return err
			}
			context.Write(key)
			context.Write(val)
		}
	})
}

// handle both space delimiters and arguments in quotations
// arguments are defined as contained by spaces ' arg ' or quotations '"arg"'
// forward slash escapes for nested quotations
func parseInput(in []byte) [][]byte {
	var results [][]byte
	var arg []byte
	state := STATE_NONE
	for _, b := range in {
		switch state {
		case STATE_NONE:
			if isQuote(b) {
				state = STATE_QUOTE
			} else if !isWhiteSpace(b) {
				arg = append(arg, b)
				state = STATE_WORD
			}
		case STATE_ESCAPED:
			arg = append(arg, b)
			state = STATE_QUOTE
		case STATE_WORD:
			if isWhiteSpace(b) {
				results = append(results, arg)
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		case STATE_QUOTE:
			if b == '\\' {
				state = STATE_ESCAPED
			} else if isQuote(b) {
				results = append(results, arg)
				arg = make([]byte, 0)
				state = STATE_NONE
			} else {
				arg = append(arg, b)
			}
		}
	}
	return results
}

func isWhiteSpace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func isQuote(b byte) bool {
	return b == '"' || b == '\''
}

func getCommand(args [][]byte) (Command, error) {
	numArgs := 0
	var cmd Command
	if len(args) == 0 {
		return cmd, errors.New("empty command")
	}
	fn := string(args[0])
	if _, ok := minArgs[fn]; !ok {
		return cmd, errors.New("invalid command")
	}
	var key, value []byte
	if len(args) >= 2 && len(args[1]) > 0 {
		key = args[1]
		numArgs++
	}
	if len(args) >= 3 && len(args[2]) > 0 {
		value = args[2]
		numArgs++
	}
	if numArgs < minArgs[fn] {
		return cmd, errors.New("not enough arguments")
	}
	return Command{
		fn:  fn,
		key: key,
		val: value,
	}, nil
}
