VERSION=`git describe --tags --always`
BUILD=`date +%FT%T%z`
HASH=`git rev-parse --short HEAD`


LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD} -X main.gitCommit=${HASH} -X main.defaultConfigPath=${DEFAULT_CONFIG_PATH}"

all: build

build:
	go build -o wssmcp ${LDFLAGS} ./cmd/wssmcp
	go build -o wsserver ${LDFLAGS} ./cmd/wsserver