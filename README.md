sqlite-bench
============

This repository contains a small Go program for performing simple,
microbenchmarks of writes to a SQLite database. It is not meant to be
representative of real world use but is meant to give a general ballpark of 
performance numbers given different journaling settings.

## Results

The following results were obtained on a MacBook Pro (13-inch, 2017), 2.3 GHz
Dual-Core Intel Core i5, with an APPLE SSD SM0128L disk.

### Journal mode DELETE

```sh
$ sqlite-bench -batch-count 1000 -batch-size 1000 -row-size 100 -journal-mode delete ~/bench.db
Inserts:   1000000 rows
Elapsed:   3.153s
Rate:      317175.758 insert/sec
File size: 110993408 bytes
```

### Journal mode WAL, Synchronous FULL

```sh
sqlite-bench -batch-count 1000 -batch-size 1000 -row-size 100 -journal-mode wal ~/bench.db
Inserts:   1000000 rows
Elapsed:   3.131s
Rate:      319398.103 insert/sec
File size: 110993408 bytes
```

### Journal mode WAL, Synchronous NORMAL

```sh
sqlite-bench -batch-count 1000 -batch-size 1000 -row-size 100 -journal-mode wal -synchronous normal ~/bench.db
Inserts:   1000000 rows
Elapsed:   2.544s
Rate:      393045.303 insert/sec
File size: 110993408 bytes
```

### Journal mode WAL, Synchronous OFF

This mode is completely unsafe. Please do not actually run this.

```sh
sqlite-bench -batch-count 1000 -batch-size 1000 -row-size 100 -journal-mode wal -synchronous off ~/bench.db
Inserts:   1000000 rows
Elapsed:   2.093s
Rate:      477827.165 insert/sec
File size: 110993408 bytes
```

