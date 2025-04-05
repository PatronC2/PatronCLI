TAG ?= snapshot

ifeq ($(OS),Windows_NT)
	PLATFORM := windows
	BINARY_NAME := patron.exe
else
	PLATFORM ?= linux
	BINARY_NAME = patron
endif

OUTDIR = output/$(PLATFORM)

.PHONY: all local release install clean

all: local install

local:
	docker buildx bake local

release:
	docker buildx bake release

install:
ifeq ($(OS),Windows_NT)
	copy $(subst /,\\,$(OUTDIR))\\$(BINARY_NAME) %WINDIR%\System32\\$(BINARY_NAME)
else
	sudo install -m 755 $(OUTDIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
endif

clean:
	rm -rf output/
