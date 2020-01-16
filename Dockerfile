# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM ubuntu AS builder

# Add Maintainer Info
LABEL maintainer="kafkaesque.io"

# Set the Current Working Directory inside the container
# WORKDIR /app/
ENV GOPATH /go
WORKDIR $GOPATH/src/github.com/pulsar-beam

RUN apt-get update
RUN apt-get install -y curl wget
RUN rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.13.1

RUN curl -sSL https://storage.googleapis.com/golang/go$GOLANG_VERSION.linux-amd64.tar.gz \
		| tar -v -C /usr/local -xz

ENV PATH /usr/local/go/bin:$PATH

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOROOT /usr/local/go
ENV PATH /go/bin:$PATH

RUN wget --user-agent=Mozilla -O apache-pulsar-client.deb "https://www.apache.org/dyn/mirrors/mirrors.cgi?action=download&filename=pulsar/pulsar-2.4.1/DEB/apache-pulsar-client.deb"
RUN wget --user-agent=Mozilla -O apache-pulsar-client-dev.deb "https://www.apache.org/dyn/mirrors/mirrors.cgi?action=download&filename=pulsar/pulsar-2.4.1/DEB/apache-pulsar-client-dev.deb"

RUN ls ./*apache-pulsar-client*.deb
RUN apt install -y ./apache-pulsar-client.deb
RUN apt install -y ./apache-pulsar-client-dev.deb

# Copy go mod and sum files
COPY go.mod go.sum ./


# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

RUN ls $GOPATH/src

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN ls $GOPATH/src/github.com/

RUN apt-get update
RUN apt-get install -y gcc

# Build the Go app
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./src
RUN go build -o main ./src

######## Start a new stage from scratch #######
FROM ubuntu:latest

# RUN apk update
WORKDIR /root/bin
RUN mkdir /root/config/

RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y --no-install-recommends apt-utils

RUN wget --user-agent=Mozilla -O apache-pulsar-client.deb "https://archive.apache.org/dist/pulsar/pulsar-2.4.1/DEB/apache-pulsar-client.deb"
RUN wget --user-agent=Mozilla -O apache-pulsar-client-dev.deb "https://archive.apache.org/dist/pulsar/pulsar-2.4.1/DEB/apache-pulsar-client-dev.deb"

RUN ls ./*apache-pulsar-client*.deb
RUN apt install -y ./apache-pulsar-client.deb
RUN apt install -y ./apache-pulsar-client-dev.deb

# Copy the Pre-built binary file from the previous stage
# COPY --from=builder $GOPATH/src/github.com/pulsar-beam/main .
COPY --from=builder /go/src/github.com/pulsar-beam/main .

RUN pwd
RUN ls

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
