package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Flags
var (
	useWAL     bool
	batchSize  int
	batchCount int
)

// Globals
var currentID int

func run(args []string) error {
	fs := flag.NewFlagSet("sqlite-bench", flag.ContinueOnError)
	fs.BoolVar(&useWAL, "use-wal", false, "use WAL")
	fs.IntVar(&batchSize, "batch-size", 1000, "batch size")
	fs.IntVar(&batchCount, "batch-count", 1000, "batch count")
	if err := fs.Parse(args); err != nil {
		return err
	} else if fs.NArg() == 0 {
		return fmt.Errorf("path required")
	} else if fs.NArg() > 1 {
		return fmt.Errorf("too many args")
	}

	filename := fs.Arg(0)

	// Remove db file, if it already exists.
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Open & initialize database.
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	defer db.Close()

	if useWAL {
		if _, err = db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
			return err
		}
	}

	if _, err = db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		return err
	}

	t := time.Now()
	if err := insertBatches(db); err != nil {
		return err
	}
	fmt.Printf("TOTAL: %d inserts over %0.03fs; %0.03f insert/sec \n", batchSize*batchCount, time.Since(t).Seconds(), float64(batchSize*batchCount)/time.Since(t).Seconds())

	time.Sleep(3 * time.Second)

	return nil
}

func insertBatches(db *sql.DB) error {
	for i := 0; i < batchCount; i++ {
		if err := insertBatch(db); err != nil {
			return err
		}
	}
	return nil
}

func insertBatch(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO t (id, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	name := strings.Repeat("x", 100)
	for i := 0; i < batchSize; i++ {
		if _, err := stmt.Exec(currentID, name); err != nil {
			return err
		}
		currentID++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
