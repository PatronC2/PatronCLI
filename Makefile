TAG ?= snapshot
PLATFORM ?= linux
BINARY_NAME = patron
OUTDIR = output/$(PLATFORM)

ifeq ($(PLATFORM),windows)
	BINARY_NAME := patron.exe
endif

.PHONY: all local release install clean

all: local install

local:
	docker buildx bake local

release:
	docker buildx bake release

install:
ifeq ($(OS),Windows_NT)
	copy $(OUTDIR)\$(BINARY_NAME) %WINDIR%\System32\$(BINARY_NAME)
else
	sudo install -m 755 $(OUTDIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
endif

clean:
	rm -rf output/