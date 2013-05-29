#!/bin/bash

#Enable negative pattern matching !(prod*).go
shopt -s extglob

WHICHFILES='./!(env_dev*).go'

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o chat.linux  ${WHICHFILES}
GOOS=darwin GOARCH=amd64 go build -o chat.osx  ${WHICHFILES}

