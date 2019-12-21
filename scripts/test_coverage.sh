#!/bin/bash

# absolute directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

BASE_PKG_DIR="github.com/pulsar-beam/src/"
ALL_PKGS=""

cd $DIR/../src
for d in */ ; do
    pkg=${d%/}
    ALL_PKGS=${ALL_PKGS}","${BASE_PKG_DIR}${pkg}
done

ALL_PKGS=$(echo $ALL_PKGS | sed 's/^,//')
echo $ALL_PKGS

cd $DIR/../src/unit-test

go test ./... -coverpkg=$ALL_PKGS -coverprofile coverage.out
go tool cover -func coverage.out > /tmp/coverage.txt

coverPercent=$(cat /tmp/coverage.txt | grep total: | awk '{print $3}' | sed 's/%$//g')

echo "Current test coverage is at ${coverPercent}%"
echo "TODO add code coverage verdict"