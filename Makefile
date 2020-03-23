.PHONY: all build

LINUX_TARGET := release/linux
DARWIN_TARGET := release/darwin
WINDOWS_TARGET := release/windows
UTILITY_NAME :=

all: test build

build:

test:
	go test -failfast