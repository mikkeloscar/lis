.PHONY: clean deps build docs

EXECUTABLE ?= lis
MANPAGE_SRCS = $(wildcard doc/*.adoc)
MANPAGES = $(MANPAGE_SRCS:.adoc=)

all: build docs

clean:
	go clean -i ./..
	rm -rf $(EXECUTABLE)
	rm -rf $(MANPAGES)

deps:
	go get -t

$(EXECUTABLE): $(wildcard *.go)
	go build -ldflags "-s"

build: $(EXECUTABLE)

docs: $(MANPAGES)

$(MANPAGES): $(MANPAGE_SRCS)
	a2x --doctype manpage --format manpage $@.adoc
