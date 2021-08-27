# golog
[![GoDoc](https://godoc.org/github.com/keakon/golog?status.svg)](https://godoc.org/github.com/keakon/golog)
[![Build Status](https://www.travis-ci.org/keakon/golog.svg?branch=master)](https://www.travis-ci.org/keakon/golog)
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
BenchmarkBufferedFileLogger-12      4186982	       267.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkDiscardLogger-12          11162383	       108.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkNopLog-12               1000000000	       0.2531 ns/op	       0 B/op	       0 allocs/op
BenchmarkMultiLevels-12             2543925	       476.6 ns/op	       0 B/op	       0 allocs/op
```

Example output of the benchmarks:
```
[I 2018-11-20 17:05:37 log_test:118] test
```
