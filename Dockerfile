# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:alpine AS builder

# Add Maintainer Info
LABEL maintainer="kafkaesque.io"

RUN apk --no-cache add build-base git bzr mercurial gcc
WORKDIR /root/
ADD . /root
RUN cd /root/src && go build -o pulsar-beam

######## Start a new stage from scratch #######
FROM alpine

# RUN apk update
WORKDIR /root/bin
RUN mkdir /root/config/

# Copy the Pre-built binary file from the previous stage
# COPY --from=builder $GOPATH/src/github.com/pulsar-beam/main .
COPY --from=builder /root/src/pulsar-beam /root/bin

# Command to run the executable
CMD ["./pulsar-beam"]
