#!/bin/sh

now=$(date +'%Y-%m-%d_%T')
go build -a -ldflags "-linkmode external -extldflags '-static' -s -w -X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$now"  -o main main.go
