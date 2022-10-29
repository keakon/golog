# golog
[![GoDoc](https://pkg.go.dev/badge/github.com/keakon/golog)](https://pkg.go.dev/github.com/keakon/golog)
[![Build Status](https://github.com/keakon/golog/actions/workflows/go.yml/badge.svg)](https://github.com/keakon/golog/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/keakon/golog)](https://goreportcard.com/report/github.com/keakon/golog)
[![codecov](https://codecov.io/gh/keakon/golog/branch/master/graph/badge.svg?token=dmE2GC9in2)](https://codecov.io/gh/keakon/golog)

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
3. When the day changing, the logging date and time might belong to different day. eg:
```
[I 2021-09-12 23:59:59 log_test:206] test
[I 2021-09-13 23:59:59 log_test:206] test
[I 2021-09-12 00:00:00 log_test:206] test
```

## Benchmarks

### Platform

```
go1.19.2 darwin/amd64
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
```

### Result

| Name | Time/op | Time (x) | Alloc/op | allocs/op |
| :--- | :---: | :---: | :---: | :---: |
| DiscardLogger-12 | 514ns ± 1% | 1.0 | 0B | 0 |
| DiscardLoggerParallel-12 | 103ns ± 5%  | 1.0 | 0B | 0 |
| DiscardLoggerWithoutTimer-12 | 715ns ± 1% | 1.39 | 0B | 0 |
| DiscardLoggerWithoutTimerParallel-12 | 148ns ± 2% | 1.44  | 0B | 0 |
| NopLog-12 | 514ns ± 1% | 1.0 | 0B | 0 |
| NopLogParallel-12 | 109ns ± 4% | 1.06 | 0B | 0 |
| MultiLevels-12 | 2.91µs ± 2% | 5.66 | 0B | 0 |
| MultiLevelsParallel-12 | 602ns ± 7% | 5.84 | 0B | 0 |
| BufferedFileLogger-12 |  587ns ± 2% | 1.14 | 0B | 0 |
| BufferedFileLoggerParallel-12 | 311ns ± 2% | 3.02 | 0B | 0 |
| | | | | |
| DiscardZerolog-12 | 2.32µs ± 2% | 4.51  | 280B | 3 |
| DiscardZerologParallel-12 | 442ns ± 5% | 4.29 | 280B | 3 |
| DiscardZap-12 | 2.41µs ± 2% | 4.69 | 313B | 6 |
| DiscardZapParallel-12 | 652ns ±15% | 6.03 | 314B | 6 |

### Example output of the benchmarks

```
[I 2018-11-20 17:05:37 log_test:118] test
```
