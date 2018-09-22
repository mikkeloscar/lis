.PHONY: clean deps build docs install

EXECUTABLES  ?= build/lis build/lisc
MANPAGE_SRCS = $(wildcard doc/*.adoc)
MANPAGES     = $(MANPAGE_SRCS:.adoc=)
SOURCES      = $(shell find . -name '*.go')
GO           ?= go
GOPKGS       = $(shell $(GO) list ./...)

all: build docs

clean:
	@rm -rf build/
	@rm -rf $(MANPAGES)

test:
	$(GO) test -v $(GOPKGS)

$(EXECUTABLES): $(SOURCES)
	$(GO) build -ldflags "-s" -o build/lis ./cmd/lis
	$(GO) build -ldflags "-s" -o build/lisc ./cmd/lisc

build: $(EXECUTABLES)

docs: $(MANPAGES)

$(MANPAGES): $(MANPAGE_SRCS)
	a2x --doctype manpage --format manpage $@.adoc

install: build docs
	# bin
	install -Dm755 build/lis $(DESTDIR)/usr/bin/lis
	install -Dm755 build/lisc $(DESTDIR)/usr/bin/lisc
	# config
	install -Dm644 lis.conf $(DESTDIR)/etc/lis.conf
	# service
	install -Dm644 contrib/lis@.service $(DESTDIR)/usr/lib/systemd/system/lis@.service
	# docs
	install -Dm644 doc/lis.1 $(DESTDIR)/usr/share/man/man1/lis.1
	install -Dm644 doc/lisc.1 $(DESTDIR)/usr/share/man/man1/lisc.1
	install -Dm644 doc/lis.conf.5 $(DESTDIR)/usr/share/man/man5/lis.conf.5
