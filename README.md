[![Go Report Card](https://goreportcard.com/badge/github.com/kafkaesque-io/pulsar-beam)](https://goreportcard.com/report/github.com/kafkaesque-io/pulsar-beam)
[![CI Build](https://github.com/kafkaesque-io/pulsar-beam/workflows/ci/badge.svg
)](https://github.com/kafkaesque-io/pulsar-beam/actions)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![codecov](https://codecov.io/gh/kafkaesque-io/pulsar-beam/branch/master/graph/badge.svg)](https://codecov.io/gh/kafkaesque-io/pulsar-beam)
[![Docker image](https://shields.beevelop.com/docker/image/image-size/kafkaesqueio/pulsar-beam/0.22.svg?style=round-square)](https://hub.docker.com/r/kafkaesqueio/pulsar-beam/)
[![LICENSE](https://img.shields.io/hexpm/l/pulsar.svg)](https://github.com/kafkaesque-io/pulsar-beam/blob/master/LICENSE)

# Pulsar Beam

Beam is an http based streaming and queueing system that is backed up by Apache Pulsar.

1. Data can be sent to Pulsar via an HTTP POST method as a producer.
2. Data can be pushed to a webhook or Cloud Function for consumption.
3. A webhook or Cloud Function can reply processed message, in the response body, back to another Pulsar topic via Pulsar Beam.

Opening an issue and PR are welcomed! Please email `contact@kafkaesque.io` for any inquiry or demo.

## Advantages
1. Since Beam speaks http, it is language and OS independent. You can take advantage of powerhouse of Apache Pulsar without limitation to choice of client and OS.

Immediately, Pulsar can be supported on Windows and any languages.

2. It's simple.

## Interface

REST API and endpoint swagger document is published at [this link](https://kafkaesque-io.github.io/pulsar-beam-swagger/)

### Endpoint to receive messages

```
/v1/firehose
```
These HTTP headers are required to map to Pulsar topic.
1. Authorization -> Bearer token as Pulsar token
2. TopicFn -> a full name of Pulsar topic (with tenant/namespace/topic) is required
3. PulsarUrl -> a fully qualified pulsar or pulsar+ssl URL is required

### Webhook registration
Webhook registration is done via REST API backed by configurable database, such as MongoDB, in momery cache, and Pulsar topic. Yes, we can use a compacted Pulsar Topic as a database table to perform CRUD. The configuration parameter is `"PbDbType": "inmemory",` in the `pulsar_beam.yml` file or the env variable `PbDbType`.

### Configuration

Both [json](./config/pulsar_beam.json) and [yml format](./config/pulsar_beam.yml) are supported as configuration file.

### Sink source

If a webhook's response contains a body and three headers including `Authorization` for Pulsar JWT, `TopicFn` for a topic fully qualified name, and `PulsarUrl`, the beam server will send the body as a new event to another Pulsar's topic specified as in TopicFn and PulsarUrl.

## Dev set up
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
$ go run main.go
```

#### Server Mode
In order to offer high performance and separate responsiblity, webhook and receiver endpoint can be running independently `-mode broker` or `-mode receiver`. By default, the server runs in a hybrid mode.

### Local CI, unit test and end to end test
There are scripts under `./scripts` folder to run code analysis, vetting, compilation, unit test, and code coverage manually as all of these are part of CI checks by Github Actions.

One end to end test is under `./src/e2e/e2etest.go`, that performs the following steps in order:
1. Create a topic and its webhook via RESTful API. The webhook URL can be an HTTP triggered Cloud Function. CI process uses a GCP 
2. Send a message to Pulsar Beam's v1 injestion endpoint
3. Waiting on the sink topic where the first message will be sent to a GCP Cloud Function (in CI) and in turn reply to Pulsar Beam to forward to the second sink topic
4. Verify the replied message on the sink topic
5. Delete the topic and its webhook document via RESTful API

Since the set up is non-trivial involving Pulsar Beam, a Cloud function or webhook, the test tool, and Pulsar itself with SSL, we recommend to take advantage of [the free plan at Kafkaesque.io](https://kafkaesque.io) as the Pulsar server and a Cloud Function that we have verified GCP Fcuntion, Azure Function or AWS Lambda will suffice in the e2e flow.

 Step to perform unit test
```bash
$ cd src/unit-test
$ go test -v .
```

### Docker image and Docker builds
The docker image can be pulled from dockerhub.io.
```
$ sudo docker pull kafkaesqueio/pulsar-beam
```

Here are steps to build docker image and run docker container in a file based configuration.

1. Build docker image
```
$ sudo docker build -t pulsar-beam .
```

2. Run docker
This is an example of a default configurations using in-memory database. Customized `pulsar_beam.yml` and private and public key files can be mounted and passed in as an env variable `PULSAR_BEAM_CONFIG`. The certificate is required to connect to Pulsar with TLS enabled.

```
$ sudo docker run -d -it -v /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem:/etc/ssl/certs/ca-bundle.crt -p 8085:8085 --name=pbeam-server pulsar-beam
```
