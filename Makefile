BINARY_NAME = ftrans
BINARY_VERSION = $(shell git describe --tags --abbrev=0)
BINARY_REVISION = $(shell git rev-parse --short HEAD)

SRCS = $(wildcard ./*.go)
DEP_FILES = Gopkg.toml Gopkg.lock vendor/

.PHONY: all build_only build dep clean
all: build

build_only: $(SRCS)
	go build -o $(BINARY_NAME) -ldflags "-X main.binaryVersion=$(BINARY_VERSION) -X main.binaryRevision=$(BINARY_REVISION)" $(SRCS)

build: dep build_only

dep: $(DEP_FILES)
	dep ensure

clean:
	rm $(BINARY_NAME)