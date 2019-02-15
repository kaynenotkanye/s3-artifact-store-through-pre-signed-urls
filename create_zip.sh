#!/usr/bin/env bash 
set -xe 

# build binary
GOARCH=amd64 GOOS=linux go build -o bin/application application.go

# create zip containing the bin, assets and .ebextensions folder
zip -r uploadThis.zip bin
