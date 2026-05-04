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
7. No 3rd party dependency
8. Fast
9. Thread safe when logger configuration is completed before concurrent logging starts

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

The fast timer is about 30% faster than calling `time.Now()` for each logging record. It is race-safe, with these timing characteristics:
1. The timer updates every 1 second, so the logging time can be at most 1 second behind the real time.
2. Concurrent readers may observe adjacent snapshots within a few milliseconds of an update, so concurrent logging messages can temporarily differ by 1 second. eg:
```
[I 2021-09-13 14:31:25 log_test:206] test
[I 2021-09-13 14:31:24 log_test:206] test
[I 2021-09-13 14:31:25 log_test:206] test
```

### ConcurrentFileWriter *(experimental)*


```go
func main() {
    w, _ := golog.NewConcurrentFileWriter("test.log")
    l := golog.NewLoggerWithWriter(w)
    defer l.Close()

    l.Infof("hello world")
}
```

The `ConcurrentFileWriter` is designed for high concurrency applications.
It is about 140% faster than `BufferedFileWriter` at 6C12H by reducing the lock overhead, but a little slower at single thread.  
**Note**: The order of logging records from different cpu cores within each 0.1 second is random.

## Notes

`Logger.Close`, `Handler.Close`, `BufferedFileWriter.Close`, and `ConcurrentFileWriter.Close` are idempotent. Writing to a closed file writer returns `os.ErrClosed`.

`ConsoleWriter.Close` and `DiscardWriter.Close` are no-op operations because they do not own an operating-system resource that should be closed by this package.

Configure loggers, handlers, and the package-level default logger before starting concurrent logging. Concurrent logging is safe after configuration is complete.

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
| NopLog | 1.5ns ± 1% | 0.003 | 0B | 0 |
| NopLogParallel | 0.22ns ± 3% | 0.002 | 0B | 0 |
| MultiLevels | 2.77µs ± 7% | 5.73 | 0B | 0 |
| MultiLevelsParallel | 532ns ± 15% | 5.98 | 0B | 0 |
| BufferedFileLogger | 588ns ± 2% | 1.22 | 0B | 0 |
| BufferedFileLoggerParallel | 241ns ± 1% | 2.71 | 0B | 0 |
| ConcurrentFileLogger | 593ns ± 1% | 1.23 | 0B | 0 |
| ConcurrentFileLoggerParallel | 101ns ± 2% | 1.13 | 0B | 0 |
| | | | | |
| DiscardZerolog | 2.24µs ± 1% | 4.64 | 280B | 3 |
| DiscardZerologParallel | 408ns ± 10% | 4.58 | 280B | 3 |
| DiscardZap | 2.13µs ± 0% | 4.41 | 272B | 5 |
| DiscardZapParallel | 465ns ± 5% | 5.22 | 274B | 5 |

* DiscardLogger: writes logs to `ioutil.Discard`
* DiscardLoggerWithoutTimer: the same as above but without fast timer
* NopLog: skips logs with lower level than the logger or handler
* MultiLevels: writes 5 levels of logs to 5 levels handlers of a warning level logger
* BufferedFileLogger: writes logs to a disk file
* ConcurrentFileLogger: writes logs to a disk file with `ConcurrentFileWriter`
* DiscardZerolog: writes logs to `ioutil.Discard` with [zerolog](https://github.com/rs/zerolog)
* DiscardZap: writes logs to `ioutil.Discard` with [zap](https://github.com/uber-go/zap) using `zap.NewProductionEncoderConfig()`

All the logs include 4 parts: level, time, caller and message. This is an example output of the benchmarks:

```
[I 2018-11-20 17:05:37 log_test:118] test
```
