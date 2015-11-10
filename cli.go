package lmdbcli

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/szferi/gomdb"
)

var (
	pathFlag = flag.String("db", "", "Relative path to lmdb file")
	sizeFlag = flag.Float64("size", 2, "factor to allocate for growth or shrinkage")
	roFlag   = flag.Bool("ro", false, "open the database in read-only mode")
	minArgs  = map[string]int{"scan": 0, "stat": 0, "expand": 0, "exists": 1, "get": 1, "del": 1, "put": 2, "exit": 0, "quit": 0}
)

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
	for {
		fmt.Print(context.prompt)
		var fn, key, val string
		fmt.Fscanln(in, &fn, &key, &val)

		if _, ok := minArgs[fn]; !ok {
			context.Write([]byte("error: invalid command"))
		} else if !checkNumArgs(fn, key, val) {
			context.Write([]byte("error: not enough arguments"))
		} else if fn == "get" {
			err = get(context, key)
		} else if fn == "exists" {
			err = exists(context, key)
		} else if fn == "del" {
			err = del(context, key)
		} else if fn == "put" {
			err = put(context, key, val)
		} else if fn == "scan" {
			err = scan(context)
		} else if fn == "quit" || fn == "exit" {
			return
		}
		if err != nil {
			context.Write([]byte(err.Error()))
		}
	}
}

func get(context *Context, key string) error {
	return context.WithinRead(func(txn *mdb.Txn) error {
		data, err := txn.Get(context.dbi, []byte(key))
		if err != nil {
			return err
		}
		context.Write(data)
		return nil
	})
}

func exists(context *Context, key string) error {
	return context.WithinRead(func(txn *mdb.Txn) error {
		_, err := txn.Get(context.dbi, []byte(key))
		if err != nil {
			context.Write([]byte("false"))
		} else {
			context.Write([]byte("true"))
		}
		return nil
	})
}

func del(context *Context, key string) error {
	return context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Del(context.dbi, []byte(key), nil)
	})
}

func put(context *Context, key, val string) error {
	return context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Put(context.dbi, []byte(key), []byte(val), 0)
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

func checkNumArgs(fn, key, val string) bool {
	if fn == "" {
		return false
	}
	n := 0
	if key != "" {
		n++
	}
	if val != "" {
		n++
	}
	if expected, ok := minArgs[fn]; ok {
		return n >= expected
	}
	return false
}
