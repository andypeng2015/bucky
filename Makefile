# Get the absolute path of the current Makefile.
MAKEFILE_PATH := $(realpath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR  := $(dir $(MAKEFILE_PATH))
BUCKY_LIB    ?= $(MAKEFILE_DIR)lib
MODELS_DIR   ?= $(HOME)/models

# make download-models fetches the GGML Whisper models used in tests/examples.
# Override MODELS_DIR=/path/to/models to put them somewhere else.
download-models:
	mkdir -p $(MODELS_DIR)
	curl -L -o $(MODELS_DIR)/ggml-tiny.bin   https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin
	curl -L -o $(MODELS_DIR)/ggml-base.en.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin

clean-whisper.cpp:
	rm -rf $(BUCKY_LIB)/*

# make download-whisper.cpp VERSION=v1.8.4 to download a specific version.
download-whisper.cpp:
	./bucky install -lib $(BUCKY_LIB) $(if $(VERSION),-v $(VERSION))

build:
	BUCKY_LIB=$(BUCKY_LIB) go build -o bucky .

install:
	go install .

# make test runs all package tests. The pkg/whisper tests require
# BUCKY_LIB and BUCKY_TEST_MODEL/AUDIO to be set; without them they skip.
test:
	export BUCKY_LIB=$(BUCKY_LIB) && \
	export BUCKY_TEST_MODEL=$(MODELS_DIR)/ggml-tiny.bin && \
	export BUCKY_TEST_AUDIO=$(MAKEFILE_DIR)samples/jfk.wav && \
	go test -count=1 ./...

# make hello runs the smallest possible bucky example end-to-end.
hello:
	export BUCKY_LIB=$(BUCKY_LIB) && \
	export BUCKY_TEST_MODEL=$(MODELS_DIR)/ggml-tiny.bin && \
	go run ./examples/hello samples/jfk.wav

vet:
	go vet ./...

fmt:
	gofmt -s -w .

.PHONY: download-models clean-whisper.cpp download-whisper.cpp build install test hello vet fmt
