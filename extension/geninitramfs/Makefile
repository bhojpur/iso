UBINDIR ?= /usr/bin
DESTDIR ?=

all: build install

build:
	@echo "Nothing to build"

install: build
	install -d $(DESTDIR)/$(UBINDIR)
	install -m 0755 bhojpur-geninitramfs $(DESTDIR)/$(UBINDIR)/