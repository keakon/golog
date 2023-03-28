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
9. Thread safe

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

The fast timer is about 30% faster than calling time.Time() for each logging record. But it's not thread-safe which may cause some problems (I think those are negligible in most cases):
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

go1.19.2 darwin/amd64
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

### Result

| Name | Time/op | Time (x) | Alloc/op | allocs/op |
| :--- | :---: | :---: | :---: | :---: |
| DiscardLogger | 483ns ± 1% | 1.00 | 0B | 0 |
| DiscardLoggerParallel | 89.0ns ± 6% | 1.00 | 0B | 0 |
| DiscardLoggerWithoutTimer | 691ns ± 7% | 1.43 | 0B | 0 |
| DiscardLoggerWithoutTimerParallel | 129ns ± 5% | 1.45 | 0B | 0 |
| NopLog | 1.5ns ±1% | 0.003 | 0B | 0 |
| NopLogParallel | 0.22ns ± 3% | 0.002 | 0B | 0 |
| MultiLevels | 2.77µs ± 7% | 5.73 | 0B | 0 |
| MultiLevelsParallel | 532ns ±15% | 5.98 | 0B | 0 |
| BufferedFileLogger | 576ns ± 5% | 1.19 | 0B | 0 |
| BufferedFileLoggerParallel | 260ns ±11% | 2.92 | 0B | 0 |
| | | | | |
| DiscardZerolog | 2.24µs ± 1% | 4.64 | 280B | 3 |
| DiscardZerologParallel | 408ns ±10% | 4.58 | 280B | 3 |
| DiscardZap | 2.13µs ±0% | 4.41 | 272B | 5 |
| DiscardZapParallel | 465ns ±5% | 5.22 | 274B | 5 |

* DiscardLogger: writes logs to ioutil.Discard
* DiscardLoggerWithoutTimer: the same as above but without fast timer
* NopLog: skips logs with lower level than the logger or handler
* MultiLevels: writes 5 levels of logs to 5 levels handlers of a warning level logger
* BufferedFileLogger: writes logs to a disk file
* DiscardZerolog: writes logs to ioutil.Discard with [zerolog](https://github.com/rs/zerolog)
* DiscardZap: writes logs to ioutil.Discard with [zap](https://github.com/uber-go/zap) using `zap.NewProductionEncoderConfig()`

All the logs include 4 parts: level, time, caller and message. This is an example output of the benchmarks:

```
[I 2018-11-20 17:05:37 log_test:118] test
```
