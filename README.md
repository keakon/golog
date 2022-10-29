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

```
go1.19.2 darwin/amd64
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

name                                  time/op
DiscardLogger-12                       514ns ± 1%
DiscardLoggerParallel-12               103ns ± 5%
DiscardLoggerWithoutTimer-12           715ns ± 1%
DiscardLoggerWithoutTimerParallel-12   148ns ± 2%
NopLog-12                              514ns ± 1%
NopLogParallel-12                      109ns ± 4%
MultiLevels-12                        2.91µs ± 2%
MultiLevelsParallel-12                 602ns ± 7%
BufferedFileLogger-12                 1.51ns ± 1%
BufferedFileLoggerParallel-12         0.25ns ± 5%
DiscardZerolog-12                     2.32µs ± 2%
DiscardZerologParallel-12              442ns ± 5%
DiscardZap-12                         2.41µs ± 2%
DiscardZapParallel-12                  652ns ±15%

name                                  alloc/op
DiscardLogger-12                       0.00B
DiscardLoggerParallel-12               0.00B
DiscardLoggerWithoutTimer-12           0.00B
DiscardLoggerWithoutTimerParallel-12   0.00B
NopLog-12                              0.00B
NopLogParallel-12                      0.00B
MultiLevels-12                         0.00B
MultiLevelsParallel-12                 0.00B
BufferedFileLogger-12                  0.00B
BufferedFileLoggerParallel-12          0.00B
DiscardZerolog-12                       280B ± 0%
DiscardZerologParallel-12               280B ± 0%
DiscardZap-12                           313B ± 0%
DiscardZapParallel-12                   314B ± 0%

name                                  allocs/op
DiscardLogger-12                        0.00
DiscardLoggerParallel-12                0.00
DiscardLoggerWithoutTimer-12            0.00
DiscardLoggerWithoutTimerParallel-12    0.00
NopLog-12                               0.00
NopLogParallel-12                       0.00
MultiLevels-12                          0.00
MultiLevelsParallel-12                  0.00
BufferedFileLogger-12                   0.00
BufferedFileLoggerParallel-12           0.00
DiscardZerolog-12                       3.00 ± 0%
DiscardZerologParallel-12               3.00 ± 0%
DiscardZap-12                           6.00 ± 0%
DiscardZapParallel-12                   6.00 ± 0%
```

Example output of the benchmarks:
```
[I 2018-11-20 17:05:37 log_test:118] test
```
