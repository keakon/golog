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
    w, _ := golog.NewBufferedFileWriter("test.log")
    l := golog.NewLoggerWithWriter(w)
    defer l.Close()

    l.Infof("hello world")
}
```

### Rotating

```go
func main() {
    w, _ := golog.NewTimedRotatingFileWriter("test", golog.RotateByDate, 30)
    l := golog.NewLoggerWithWriter(w)
    defer l.Close()

    l.Infof("hello world")
}
```

### Formatting

```go
func main() {
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

### Fast timer

```go
func main() {
    golog.StartFastTimer()
    defer golog.StopFastTimer()

    l := golog.NewStdoutLogger()
    defer l.Close()

    l.Infof("hello world")
}
```

The fast timer is about 30% faster than calling time.Time() for each logging record. But it's not thread-safe which may cause some problems (I think those are neglectable in most cases):
1. The timer updates every 1 second, so the logging time can be at most 1 second behind the real time.
2. Each thread will notice the changes of timer in a few milliseconds, so the concurrent logging messages may get different logging time (less than 2% probability). eg:
```
[I 2021-09-13 14:31:25 log_test:206] test
[I 2021-09-13 14:31:24 log_test:206] test
[I 2021-09-13 14:31:25 log_test:206] test
```
3. When the day changing, the logging date and time might from different day. eg:
```
[I 2021-09-12 23:59:59 log_test:206] test
[I 2021-09-13 23:59:59 log_test:206] test
[I 2021-09-12 00:00:00 log_test:206] test
```

## Benchmarks

```
go1.17 darwin/amd64
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

BenchmarkDiscardLogger-12                 13788436       77.83 ns/op       0 B/op      0 allocs/op
BenchmarkDiscardLoggerWithoutTimer-12     10472464       112.1 ns/op       0 B/op      0 allocs/op
BenchmarkNopLog-12                      1000000000      0.1918 ns/op       0 B/op      0 allocs/op
BenchmarkMultiLevels-12                    3922504       304.8 ns/op       0 B/op      0 allocs/op
BenchmarkBufferedFileLogger-12             4937521       251.7 ns/op       0 B/op      0 allocs/op

BenchmarkDiscardZerolog-12                 4074019       289.8 ns/op     280 B/op      3 allocs/op
BenchmarkDiscardZap-12                     2908678       409.8 ns/op     321 B/op      7 allocs/op
```

Example output of the benchmarks:
```
[I 2018-11-20 17:05:37 log_test:118] test
```
