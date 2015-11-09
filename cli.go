package lmdbcli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/szferi/gomdb"
)

var (
	pathFlag = flag.String("db", "", "Relative path to lmdb file")
	sizeFlag = flag.Float64("size", 1, "factor to allocate for growth or shrinkage")
	roFlag   = flag.Bool("ro", false, "open the database in read-only mode")
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
	runShell(context)
}

func runShell(context *Context) {
	var err error
	for {
		fmt.Print(context.prompt)
		var f, i, d string
		fmt.Scanln(&f, &i, &d)

		if f == "" || i == "" {
			context.Write([]byte("invalid command"))
		} else if f == "get" {
			err = get(context, i)
		} else if f == "del" {
			err = del(context, i)
		} else if f == "scan" {
			err = scan(context)
		} else {
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

func del(context *Context, key string) error {
	return context.WithinWrite(func(txn *mdb.Txn) error {
		return txn.Del(context.dbi, []byte(key), nil)
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
