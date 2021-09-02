# golog
[![GoDoc](https://godoc.org/github.com/keakon/golog?status.svg)](https://godoc.org/github.com/keakon/golog)
[![Build Status](https://app.travis-ci.com/keakon/golog.svg?branch=master)](https://www.travis-ci.com/keakon/golog)
[![Go Report Card](https://goreportcard.com/badge/github.com/keakon/golog)](https://goreportcard.com/report/github.com/keakon/golog)

## Features

1. Unstructured
2. Leveled
3. With caller (file path / name and line number)
4. Customizable output layout
5. Rotating by size, date or hour
6. Cross platform, tested on Linux, macOS and Windows
7. No 3rd party dependance
8. Fast

## Installation

```
go get -u github.com/keakon/golog
```

## Benchmarks

```
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkBufferedFileLogger-12    4364552	       268.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkDiscardLogger-12         9694594	       125.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkNopLog-12             1000000000	       0.2551 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiLevels-12           2174760	       560.6 ns/op	       0 B/op	       0 allocs/op
```

Example output of the benchmarks:
```
[I 2018-11-20 17:05:37 log_test:118] test
```
