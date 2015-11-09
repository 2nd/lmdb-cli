package main

import (
	"flag"
	"fmt"

	"github.com/szferi/gomdb"
)

var (
	dbPath      = *flag.String("db", "", "Relative path to lmdb file")
	shellPrompt = "golmdb>"
)

// Run golmdb using the directory containing the data as dbPath

func main() {
	flag.Parse()
	if dbPath == "" {
		dbPath = flag.Args()[0]
	}
	env, _ := mdb.NewEnv()
	env.SetMapSize(1 << 20)
	if err := env.Open(dbPath, 0, 0664); err != nil {
		fmt.Println("open environment failed")
		return
	}
	defer env.Close()

	txn, _ := env.BeginTxn(nil, 0)
	dbi, _ := txn.DBIOpen(nil, 0)
	defer env.DBIClose(dbi)
	txn.Commit()

	runShell(env, txn, dbi)
}

func runShell(env *mdb.Env, txn *mdb.Txn, dbi mdb.DBI) {
	running := true
	for running {
		fmt.Print(shellPrompt)
		var f, i, d string
		fmt.Scanln(&f, &i, &d)

		if f == "" || i == "" {
			fmt.Println("invalid command")
		} else if f == "get" {
			fmt.Println(mdbGet(env, dbi, i))
		} else if f == "del" {
			fmt.Println(mdbDel(env, dbi, i))
		} else if f == "scan" {
			scanMdb(env, dbi)
		} else {
			running = false
		}
	}
	return
}

func mdbGet(env *mdb.Env, dbi mdb.DBI, key string) string {
	txn, _ := env.BeginTxn(nil, mdb.RDONLY)
	defer txn.Reset()
	data, err := txn.Get(dbi, []byte(key))
	if err != nil {
		return "get failed"
	}
	return string(data)
}

func mdbDel(env *mdb.Env, dbi mdb.DBI, key string) string {
	txn, _ := env.BeginTxn(nil, 0)
	if err := txn.Del(dbi, []byte(key), nil); err != nil {
		txn.Abort()
		return "error when deleting entry"
	}
	txn.Commit()
	return "entry successfully deleted"
}

func scanMdb(env *mdb.Env, dbi mdb.DBI) {
	txn, _ := env.BeginTxn(nil, mdb.RDONLY)
	defer txn.Abort()
	cursor, _ := txn.CursorOpen(dbi)
	defer cursor.Close()
	for {
		key, val, err := cursor.Get(nil, nil, mdb.NEXT)
		if err == mdb.NotFound {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(string(key), string(val))
	}
}
