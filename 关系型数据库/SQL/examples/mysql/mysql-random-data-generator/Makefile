GO := go
pkgs   = $(shell basename `git rev-parse --show-toplevel`)
VERSION ?=$(shell git describe --abbrev=0)
BUILD ?=$(shell date +%FT%T%z)
GOVERSION ?=$(shell go version | cut --delimiter=" " -f3)
COMMIT ?=$(shell git rev-parse HEAD)
BRANCH ?=$(shell git rev-parse --abbrev-ref HEAD)
GOPATH ?=${HOME}/go

MAKE_TARS = ''
CUR_DIR=$(shell pwd)
BIN_DIR=${CUR_DIR}/build
LDFLAGS="-X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Commit=${COMMIT} -X main.Branch=${BRANCH} -X main.GoVersion=${GOVERSION} -s -w"

ifeq (${GOPATH},)
$(error GOPATH is not set)
endif

ifeq (,$(wildcard ${GOPATH}/src))
$(error Invalid GOPATH. There is no src dir in the GOPATH) 
endif

ifeq ($(findstring ${GOPATH},${CUR_DIR}), )
$(error Wrong directorry for the project. It must be in $GOPATH/github/Percona-Lab/mysql_random_data_load)
endif

$(info )
$(info GOPATH..........: ${GOPATH})
$(info Build directory.: ${BIN_DIR})
$(info )

.PHONY: all style format build test vet tarball linux-amd64

default: prepare
	@$(info Cleaning old tar files in ${BIN_DIR})
	@rm -f ${BIN_DIR}/mysql_random_data_load_*.tar.gz
	@echo
	@$(info Building in ${BIN_DIR})
	@go build -ldflags ${LDFLAGS} -o ${BIN_DIR}/mysql_random_data_load main.go

prepare:
	@$(info Checking if ${BIN_DIR} exists)
	@mkdir -p ${BIN_DIR}

all: clean darwin-amd64-tar linux-amd64-tar 

clean: prepare
	@$(info Cleaning binaries and tar.gz files in dir ${BIN_DIR})
	@rm -f ${BIN_DIR}/mysql_random_data_load
	@rm -f ${BIN_DIR}/mysql_random_data_load_*.tar.gz

linux-amd64: prepare
	@echo "Building linux/amd64 binaries in ${BIN_DIR}"
	@GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ${BIN_DIR}/mysql_random_data_load main.go

linux-amd64-tar: linux-amd64
	@tar cvzf ${BIN_DIR}/mysql_random_data_load_linux_amd64.tar.gz -C ${BIN_DIR} mysql_random_data_load

linux-386: prepare
	@echo "Building linux/386 binaries in ${BIN_DIR}"
	@GOOS=linux GOARCH=386 go build -ldflags ${LDFLAGS} -o ${BIN_DIR}/mysql_random_data_load main.go

linux-386-tar: linux-386
	@tar cvzf ${BIN_DIR}/mysql_random_data_load_linux_386.tar.gz -C ${BIN_DIR} mysql_random_data_load

darwin-amd64: 
	@echo "Building darwin/amd64 binaries in ${BIN_DIR}"
	@mkdir -p ${BIN_DIR}
	@GOOS=darwin GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ${BIN_DIR}/mysql_random_data_load main.go

darwin-amd64-tar: darwin-amd64
	@tar cvzf ${BIN_DIR}/mysql_random_data_load_darwin_amd64.tar.gz -C ${BIN_DIR} mysql_random_data_load

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
	@echo ">> running tests"
	@./runtests.sh

clean-tests:
	@$(info Cleaning up docker containers used for tests)
	@docker-compose down

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

