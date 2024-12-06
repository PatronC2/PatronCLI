#!/bin/bash -xe

export GOOS=linux
export GOARCH=amd64

make build

sudo mv /output/patron /usr/bin
