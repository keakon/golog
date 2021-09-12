# golog
[![GoDoc](https://godoc.org/github.com/keakon/golog?status.svg)](https://godoc.org/github.com/keakon/golog)
[![Build Status](https://app.travis-ci.com/keakon/golog.svg?branch=master)](https://app.travis-ci.com/github/keakon/golog)
[![Go Report Card](https://goreportcard.com/badge/github.com/keakon/golog)](https://goreportcard.com/report/github.com/keakon/golog)

## Features

1. Unstructured
2. Leveled
3. With caller (file path / name and line number)
4. Customizable output format
5. Rotating by size, date or hour
6. Cross platform, tested on Linux, macOS and Windows
7. No 3rd party dependance
8. Fast

## Installation

```
go get -u github.com/keakon/golog
```

## Examples

### Logging to console

```go
package main

import (
    "github.com/keakon/golog"
    "github.com/keakon/golog/log"
)

func main() {
    golog.StartFastTimer()
    defer golog.StopFastTimer()

    l := golog.NewStdoutLogger()
    defer l.Close()

    l.Infof("hello %d", 1)

    log.SetDefaultLogger(l)
    test()
}

func test() {
    log.Infof("hello %d", 2)
}
```

### Logging to file

```go
func main() {
    golog.StartFastTimer()
    defer golog.StopFastTimer()

    w, _ := golog.NewBufferedFileWriter("test.log")
    l := golog.NewLoggerWithWriter(w)
    defer l.Close()

    l.Infof("hello world")
}
```

### Rotating

```go
func main() {
    golog.StartFastTimer()
    defer golog.StopFastTimer()

    w, _ := golog.NewTimedRotatingFileWriter("test", golog.RotateByDate, 30)
    l := golog.NewLoggerWithWriter(w)
    defer l.Close()

    l.Infof("hello world")
}
```

### Formatting

```go
func main() {
    golog.StartFastTimer()
    defer golog.StopFastTimer()

    w := golog.NewStdoutWriter()

    f := golog.ParseFormat("[%l %D %T %S] %m")
    h := golog.NewHandler(golog.InfoLevel, f)
    h.AddWriter(w)

    l := golog.NewLogger(golog.InfoLevel)
    l.AddHandler(h)
    defer l.Close()

    l.Infof("hello world")
}
```

Check [document](https://pkg.go.dev/github.com/keakon/golog#Formatter.Format) for more format directives.

## Benchmarks

```
go1.17 darwin/amd64
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

BenchmarkBufferedFileLogger-12      4470165         281.8 ns/op        0 B/op          0 allocs/op
BenchmarkDiscardLogger-12          14886738         79.74 ns/op        0 B/op          0 allocs/op
BenchmarkNopLog-12               1000000000        0.2131 ns/op        0 B/op          0 allocs/op
BenchmarkMultiLevels-12             3647791         332.0 ns/op        0 B/op          0 allocs/op

BenchmarkDiscardZerolog-12          4112203         293.5 ns/op      280 B/op          3 allocs/op
BenchmarkDiscardZap-12              3086234         398.6 ns/op      321 B/op          7 allocs/op
```

Example output of the benchmarks:
```
[I 2018-11-20 17:05:37 log_test:118] test
```
