# Pulsar Beam

Beam is a http based streaming and queueing system that is backed up by Apache Pulsar.

1. Data can be sent to Pulsar via an HTTP POST method as a producer.
2. To consume the data, data can be pushed to a webhook.
3. To consume the data, data can be retrieved by an HTTP GET method.

## Advantages
1. Since Beam speaks http, it is language and OS independent. You can take advantage of powerhouse of Apache Pulsar without limitation to choice of client and OS.

Immediately, Pulsar can be supported on Windows and any languages.

2. It's simple.

## Interface

### Endpoint to receive messages

```
/v1/{tenant}
```
These HTTP headers are required to map to Pulsar topic.
1. Authorization -> Bearer token as Pulsar token
2. TopicUrl -> a fully qualified Pulsar topic URL (with tenant/namespace/topic) is required
3. PulsarUrl -> a fully qualified pulsar or pulsar+ssl URL is required

### Webhook registration

In the current prototype, webhook registration is specified in ./config/prototype-db/default.json.

The registration will be moved to a permenant database soon.

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
