FROM golang:1.12

# Get git
RUN apt-get update \
    && apt-get -y install curl git \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Where pegnet sources will live
WORKDIR $GOPATH/src/github.com/pegnet/pegnet

# Get goveralls for testing/coverage
RUN go get github.com/mattn/goveralls

# Populate the rest of the source
COPY . .

ARG GOOS=linux
ENV GO111MODULE=on

# Setup the cache directory
RUN mkdir -p /root/.pegnet/
COPY ./config/defaultconfig.ini /root/.pegnet/defaultconfig.ini

RUN go get
RUN go build initialization/main.go
RUN go build pegnet.go