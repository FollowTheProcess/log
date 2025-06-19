# log

[![License](https://img.shields.io/github/license/FollowTheProcess/log)](https://github.com/FollowTheProcess/log)
[![Go Reference](https://pkg.go.dev/badge/go.followtheprocess.codes/log.svg)](https://pkg.go.dev/go.followtheprocess.codes/log)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/log)](https://goreportcard.com/report/github.com/FollowTheProcess/log)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/log?logo=github&sort=semver)](https://github.com/FollowTheProcess/log)
[![CI](https://github.com/FollowTheProcess/log/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/log/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/log/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/log)

Simple, fast, opinionated logging for command line applications 🪵

<p align="center">
<img src="https://github.com/FollowTheProcess/log/raw/main/docs/img/demo.gif" alt="demo">
</p>

## Project Description

`log` is a tiny and incredibly simple logging library designed to output nicely presented, human readable, levelled log messages. Ideal for command line applications ✨

There are many great logging libraries for Go out there, but so many of them are IMO too flexible and too complicated. I wanted a small, minimal dependency, opinionated logger I could
use everywhere across all my Go projects (which are mostly command line applications). So I made one 🚀

## Installation

```shell
go get go.followtheprocess.codes/log@latest
```

## Quickstart

```go
package main

import (
    "fmt"
    "os"

    "go.followtheprocess.codes/log"
)

func main() {
    logger := log.New(os.Stderr)

    logger.Debug("Debug me") // By default this one won't show up, default log level is INFO
    logger.Info("Some information here", "really", true)
    logger.Warn("Uh oh!")
    logger.Error("Goodbye")
}
```

## Usage Guide

Make a new logger

```go
logger := log.New(os.Stderr)
```

### Levels

`log` provides a levelled logger with the normal levels you'd expect:

```go
log.LevelDebug
log.LevelInfo
log.LevelWarn
log.LevelError
```

You write log lines at these levels with the corresponding methods on the `Logger`:

```go
logger.Debug("...") // log.LevelDebug
logger.Info("...")  // log.LevelInfo
logger.Warn("...")  // log.LevelWarn
logger.Error("...") // log.LevelError
```

And you can configure a `Logger` to display logs at or higher than a particular level with the `WithLevel` option...

```go
logger := log.New(os.Stderr, log.WithLevel(log.LevelDebug))
```

### Key Value Pairs

`log` provides "semi structured" logs in that the message is free form text but you can attach arbitrary key value pairs to any of the log methods

```go
logger.Info("Doing something", "cache", true, "duration", 30 * time.Second, "number", 42)
```

You can also create a "sub logger" with persistent key value pairs applied to every message

```go
sub := logger.With("sub", true)

sub.Info("Hello from the sub logger", "subkey", "yes") // They can have their own per-method keys too!
```

<p align="center">
<img src="https://github.com/FollowTheProcess/log/raw/main/docs/img/keys.gif" alt="demo">
</p>

### Prefixes

`log` lets you apply a "prefix" to your logger, either as an option to `log.New` or by creating a "sub logger" with that prefix!

```go
logger := log.New(os.Stderr, log.Prefix("http"))
```

Or...

```go
logger := log.New(os.Stderr)
prefixed := logger.Prefixed("http")
```

<p align="center">
<img src="https://github.com/FollowTheProcess/log/raw/main/docs/img/prefix.gif" alt="demo">
</p>
