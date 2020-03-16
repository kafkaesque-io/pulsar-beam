#!/bin/bash

#
# Run the CI flow and build the binary
# Prerequisite -
# 1. Go runtime
#

# absolute directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

BASE_PKG_DIR="github.com/kafkaesque-io/pulsar-beam/src/"
ALL_PKGS=""

cd $DIR/../src
# test lint, vet, and build as basic build steps in CI
echo run golint
golint ./...
echo run go vet
go vet ./...

echo run go build
mkdir -p ${DIR}/../bin
rm -f ${DIR}/../bin/pulsar-beam
go build -o ${DIR}/../bin/pulsar-beam .
