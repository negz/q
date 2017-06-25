#!/usr/bin/env bash

set -e

# Install dep Go dependency manager.
go get -u github.com/golang/dep/cmd/dep

# Setup the vendor dir.
dep ensure

# We're building Docker images, so build for Linux.
export GOOS="linux"
export GOARCH="amd64"

pushd cmd/q
    go build .
    docker build --tag negz/queue:$(git rev-parse --short HEAD) .
    docker build --tag negz/queue:latest .
    rm q
popd

pushd cmd/qrest
    go build .
    docker build --tag negz/qrest:$(git rev-parse --short HEAD) .
    docker build --tag negz/qrest:latest .
    rm qrest
popd

pushd cmd/qcli
    go build .
    docker build --tag negz/qcli:$(git rev-parse --short HEAD) .
    docker build --tag negz/qcli:latest .
    rm qcli
popd