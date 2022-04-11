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

// Globals
var currentID int

func run(args []string) error {
	fs := flag.NewFlagSet("sqlite-bench", flag.ContinueOnError)
	journalMode := fs.String("journal-mode", "delete", "journal mode (delete, truncate, persist, memory, wal, off)")
	synchronous := fs.String("synchronous", "full", "synchronous mode (off, normal, full, extra)")
	batchSize := fs.Int("batch-size", 1000, "rows per batch")
	batchCount := fs.Int("batch-count", 1000, "number of batches")
	rowSize := fs.Int("row-size", 100, "row size, in bytes")
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

	// Set journal settings.
	if _, err = db.Exec(`PRAGMA journal_mode = ` + *journalMode); err != nil {
		return err
	} else if _, err = db.Exec(`PRAGMA synchronous = ` + *synchronous); err != nil {
		return err
	}

	// Initialize schema.
	if _, err = db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		return err
	}

	// Generate name string once. SQLite doesn't compress values so it shouldn't matter.
	name := strings.Repeat("x", *rowSize)

	t := time.Now()
	for i := 0; i < *batchCount; i++ {
		if err := insertBatch(db, *batchSize, name); err != nil {
			return err
		}
	}

	rowN := (*batchSize) * (*batchCount)
	elapsed := time.Since(t)

	// Checkpoint if using WAL.
	if _, err := db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return err
	}

	// Check file size.
	fi, err := os.Stat(fs.Arg(0))
	if err != nil {
		return err
	}

	fmt.Printf("Inserts:   %d rows\n", rowN)
	fmt.Printf("Elapsed:   %0.03fs\n", elapsed.Seconds())
	fmt.Printf("Rate:      %0.03f insert/sec\n", float64(rowN)/elapsed.Seconds())
	fmt.Printf("File size: %d bytes\n", fi.Size())

	return nil
}

func insertBatch(db *sql.DB, batchSize int, name string) error {
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
