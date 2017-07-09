.PHONY: clean deps build docs install

EXECUTABLES ?= lis lisc
MANPAGE_SRCS = $(wildcard doc/*.adoc)
MANPAGES = $(MANPAGE_SRCS:.adoc=)

all: build docs

clean:
	go clean -i ./..
	@rm -f $(EXECUTABLES)
	@rm -rf $(MANPAGES)

deps:
	go get -t

$(EXECUTABLES): $(wildcard *.go)
	go build -ldflags "-s" -o lis ./cmd/lis
	go build -ldflags "-s" -o lisc ./cmd/lisc

build: $(EXECUTABLES)

docs: $(MANPAGES)

$(MANPAGES): $(MANPAGE_SRCS)
	a2x --doctype manpage --format manpage $@.adoc

install: build docs
	# bin
	install -Dm755 lis $(DESTDIR)/usr/bin/lis
	install -Dm755 lisc $(DESTDIR)/usr/bin/lisc
	# config
	install -Dm644 lis.conf $(DESTDIR)/etc/lis.conf
	# service
	install -Dm644 contrib/lis.service $(DESTDIR)/usr/lib/systemd/system/lis.service
	# docs
	install -Dm644 doc/lis.1 $(DESTDIR)/usr/share/man/man1/lis.1
	install -Dm644 doc/lisc.1 $(DESTDIR)/usr/share/man/man1/lisc.1
	install -Dm644 doc/lis.conf.5 $(DESTDIR)/usr/share/man/man5/lis.conf.5
