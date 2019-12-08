#!/bin/bash

#
# a script to install go dependencies
# used for CI (i.e. GitHub actions)
#
go get -u github.com/gorilla/mux
go get -u github.com/rs/cors
go get -u github.com/patrickmn/go-cache
go get -u github.com/hashicorp/go-retryablehttp