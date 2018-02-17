BINARY_NAME = ftrans
BINARY_VERSION = $(shell git describe --tags --abbrev=0)
BINARY_REVISION = $(shell git rev-parse --short HEAD)

SRCS = $(wildcard ./*.go)
DEP_FILES = Gopkg.toml Gopkg.lock vendor/

.PHONY: all build dep clean
all: build

build: $(SRCS) dep
	go build -o $(BINARY_NAME) -ldflags "-X main.binaryVersion=$(BINARY_VERSION) -X main.binaryRevision=$(BINARY_REVISION)" $(SRCS)

dep: $(DEP_FILES)
	dep ensure

clean:
	rm $(BINARY_NAME)