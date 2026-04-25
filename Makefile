TAG ?= snapshot
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

ifeq ($(OS),Windows_NT)
	PLATFORM := windows
	BINARY_NAME := patron.exe
else
	PLATFORM ?= linux
	BINARY_NAME = patron
endif

OUTDIR = output/$(PLATFORM)

.PHONY: all local release install clean tests

all: local install

local:
	TAG=$(TAG) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) docker buildx bake tests
	TAG=$(TAG) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) docker buildx bake local

release:
	TAG=$(TAG) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) docker buildx bake tests
	TAG=$(TAG) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) docker buildx bake release

test:
	TAG=$(TAG) COMMIT=$(COMMIT) BUILD_DATE=$(BUILD_DATE) docker buildx bake tests

install:
ifeq ($(OS),Windows_NT)
	copy $(subst /,\\,$(OUTDIR))\\$(BINARY_NAME) %WINDIR%\System32\\$(BINARY_NAME)
else
	sudo install -m 755 $(OUTDIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
endif

clean:
	rm -rf output/
