# A command line utility for LMDB

A command line utility for [LMDB](http://symas.com/mdb/) aimed at providing a simple way to explore an LMDB database.

## Usage

The simplest usage is to launch the program passing it the path to your database:

```
lmdb-cli /path/to/db
```

Alternatively, the `-db` flag can be specified.

### Other Options

- `size BYTES`: Size of memory map to use for a new database only. This value is ignored if the database already exists. [33554432 (32MB)]
- `growth #`: Grow (or shrink) the memory mapped size by the specified filter. [1]
- `db PATH`: Path to the folder containing the database.
- `ro`: Opens the database in read-only mode
- `dbs #`: Sets the maximum number of named databases
- `c COMMAND`: executes the specified command without entering the shell
- `dir=false`: Open the database with `MDB_NOSUBDIR`

## Shell Commands

Typing `help` will provide a list of available commands.
