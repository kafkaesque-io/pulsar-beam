# Pulsar Beam

Beam is a http based streaming and queueing system that is backed up by Apache Pulsar.

## Advantages
1. Since Beam speaks http, it is language and OS independent. You can take advantage of powerhouse of Apache Pulsar without limitation to choice of client and OS.

2. It's simple.

## set up
clone the repo at your gopath github.com/pulsar-beam folder.

### Linting
Install golint.
```bash
$ go install github.com/golang/lint
```

```bash
$ cd src
$ golint ./...
```

### How to run 
The steps how to start the web server.
```bash
$ cd src
$ go run *.go
```

### How to run unit test
```bash
$ cd src/unit-test
$ go test -v .
```
