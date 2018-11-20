# golog [![Build Status](https://www.travis-ci.org/keakon/golog.svg?branch=master)](https://www.travis-ci.org/keakon/golog)

## Features

1. Unstructured
2. Leveled
3. With caller (file path / name and line number)
4. Rotating by size, date or hour
5. No 3rd party dependancy
6. Fast

## Benchmarks

```
BenchmarkBufferedFileLogger-8   	 5000000	       304 ns/op	       0 B/op	       0 allocs/op
BenchmarkDiscardLogger-8        	 5000000	       257 ns/op	       0 B/op	       0 allocs/op
```

Example output:
```
[I 2018-11-20 17:05:37 log_test.go:118] test
```
