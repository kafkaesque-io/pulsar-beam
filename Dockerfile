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

# Copy the Pre-built binary file and default configuraitons from the previous stage
COPY --from=builder /root/src/pulsar-beam /root/bin
COPY --from=builder /root/config/pulsar_beam_inmemory_db.yml /root/config/pulsar_beam.yml
COPY --from=builder /root/src/unit-test/example_p* /root/config/

# Command to run the executable
ENTRYPOINT ["./pulsar-beam"]
