# Pulsar Beam

Beam is a http based streaming and queueing system that is backed up by Apache Pulsar.

1. Data can be sent to Pulsar via an HTTP POST method as a producer.
2. To consume the data, data can be pushed to a webhook.
3. To consume the data, data can be retrieved by an HTTP GET method.

The development is in an early stage. Please email `contact@kafkaesque.io` for any inquiry or demo. Opening an issue and PR are welcomed!

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
2. TopicFn -> a full name of Pulsar topic (with tenant/namespace/topic) is required
3. PulsarUrl -> a fully qualified pulsar or pulsar+ssl URL is required

### Webhook registration

In the current prototype, webhook registration is specified in ./config/prototype-db/default.json.

The registration will be moved to a permenant database soon.

### Sink source

If a webhook's response contains a body and three headers including Authorization, TopicFn, and PulsarUrl, the beam server will send the body as a new event to another Pulsar's topic specified as in TopicFn and PulsarUrl.

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

#### Server Mode
In order to offer high performance and separate responsiblity, webhook and receiver endpoint can be running independently `-mode broker` or `-mode receiver`. By default, the server runs in a hybrid mode.

### How to run unit test
```bash
$ cd src/unit-test
$ go test -v .
```

### Docker-compose and Docker builds
Run `sudo docker-compose up` to start the Pulsar beam server in hybrid mode with external database. 

Here are steps to build docker image and run docker container in a file based configuration.

1. Build docker image
```
$ sudo docker build -t pulsar-beam .
```

2. Run docker
This is an example of file based user Pulsar topic configurations.

```
$ sudo docker run -d -it -v /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem:/etc/ssl/certs/ca-bundle.crt -v /home/ming/go/src/github.com/pulsar-beam/config:/root/config -p 3000:3000 --name=pbeam-server pulsar-beam
```