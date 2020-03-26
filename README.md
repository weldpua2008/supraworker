# Supraworker [![Build Status](https://travis-ci.org/weldpua2008/supraworker.svg?branch=master)](https://travis-ci.org/weldpua2008/supraworker) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The abstraction layer around jobs, allows pull a job from your API periodically, call-back your API, observe execution time and to control concurrent execution.

It's responsible for getting the bash commands from your API running it, and streaming the logs back to your API. It also sends state updates to your API.

## Getting started

```bash
$ go get github.com/weldpua2008/supraworker
```

### Running tests

*  expires all test results

```bash
$ go clean -testcache
```
* run all tests

```bash
$ go test -bench= -test.v  ./...
```

## Installing

### from binary

Find the version you wish to install on the [GitHub Releases
page](https://github.com/weldpua2008/supraworker/releases) and download either the
`darwin-amd64` binary for macOS or the `linux-amd64` binary for Linux. No other
operating systems or architectures have pre-built binaries at this time.

### from source

1. install [Go](http://golang.org) `v1.12+`
1. clone this down into your `$GOPATH`
  * `mkdir -p $GOPATH/src/github.com/weldpua2008`
  * `git clone https://github.com/weldpua2008/supraworker $GOPATH/src/github.com/weldpua2008/supraworker`
  * `cd $GOPATH/src/github.com/weldpua2008/supraworker`
1. install [gometalinter](https://github.com/alecthomas/gometalinter):
  * `go get -u github.com/alecthomas/gometalinter`
  * `gometalinter --install`
1. install [shellcheck](https://github.com/koalaman/shellcheck)
1. `make`
