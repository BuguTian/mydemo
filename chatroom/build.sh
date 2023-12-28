#!/bin/bash

# variable
BUILDTIME=$(date +'%Y-%m-%d %H:%M:%S')
VERSION="1.0.1"
GITREV="1"

# update-version
echo "package main" > version.go
echo  "" >> version.go
echo "const GITREV = \""${GITREV}"\"" >> version.go
echo "const VERSION = \""${VERSION}"\"" >> version.go
echo "const BUILDTIME = \""${BUILDTIME}"\"" >> version.go
echo "const MAINNET = true" >> version.go

# build 
go build -ldflags "-s -w" ./

# run: ./chatroom -daemon -forever