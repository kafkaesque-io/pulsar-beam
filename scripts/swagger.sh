#!/bin/bash
#
# This script has instructions to build swagger document.
# The swagger document is currently hosted on Github pages by another repo,
# https://github.com/kafkaesque-io/pulsar-beam-swagger
#

# absolute directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

#
# Set up vendor package if it is not in use
#
#GO111MODULE=on go mod vendor

# Install go-swagger
#
# This is required to run the swagger executable in the next step.
#
#GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger

#
# Generate swagger
#
cd $DIR/..
GO111MODULE=off swagger generate spec -o ./swagger.yaml --scan-models

#
# test locally
# swagger must be installed
# it will start a webserver with swagger documents
#
swagger serve -F=swagger swagger.yaml

#
# swagger.yaml is required to submit to the repo https://github.com/kafkaesque-io/pulsar-beam-swagger
# the swagger document is published at https://kafkaesque-io.github.io/pulsar-beam-swagger/
# 

