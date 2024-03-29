BINARY="demo"
VERSION=0.0.1
BUILD=`date +%F`
SHELL:=/bin/bash
BINARY:=utMsgDaemon

#version注入目录
versionDir="github.com/jibenliu/utMsgDaemon/service"
gitTag=$(shell git tag --sort=committerdate | tail -n 1)
gitBranch=$(shell git rev-parse --abbrev-ref HEAD)
buildTime=$(shell TZ=Asia/Shanghai date +%FT%T%z)
goVersion=$(shell go env GOVERSION)
gitCommit=$(shell git rev-parse --short HEAD)

ldflags="-s -w -X '${versionDir}.GoVersion=${goVersion}' -X '${versionDir}.Version=${VERSION}' -X '${versionDir}.GitBranch=${gitBranch}' -X '${versionDir}.GitTag=${gitTag}' -X '${versionDir}.GitCommit=${gitCommit}' -X '${versionDir}.BuildTime=${buildTime}'"

help:
	@echo "usage: make <option>"
	@echo "options and effects:"
	@echo "    help   : Show help"
	@echo "    all    : Build multiple binary of this project"
	@echo "    build  : Build the binary of this project for current platform"
	@echo "    clean  : clean build directory and files of this project"
	@echo "    rebuild: ReBuild the linux binary of this project"

default: build

build:
	@echo ${gitBranch}
	@echo "build the ${BINARY}"
	@echo "the flag is: "${ldflags}
	@go build -ldflags ${ldflags} -o  tmp/${BINARY}  -tags=jsoniter
	@echo "build done."

clean:
	rm -rf tmp

rebuild: clean build
