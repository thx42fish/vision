ifeq ($(GOPATH),)
	PATH := $(HOME)/go/bin:$(PATH)
else
	PATH := $(GOPATH)/bin:$(PATH)
endif

export GO111MODULE=on

default:
	scripts/build.sh server linux
	scripts/build.sh client linux

docker: default
	docker build -t vision-server -f scripts/docker/server/Dockfile .
	docker build -t vision-client -f scripts/docker/client/Dockfile .

darwin:
	scripts/build.sh server darwin
	scripts/build.sh client darwin

clean:
	rm -f bin/*

init:
	go mod download
	go mod tidy

all: clean default docker

.PHONY: clean default