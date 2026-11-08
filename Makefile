.PHONY: all build build-arm64 install uninstall clean

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

all: build build-arm64

build:
	go build -ldflags="-s -w" -o tfetc .

build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o tfetc-arm64 .

install: build
	@mkdir -p $(DESTDIR)$(BINDIR)
	install -m 755 tfetc $(DESTDIR)$(BINDIR)/tfetc
	@echo "tfetc installed to $(DESTDIR)$(BINDIR)/tfetc"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/tfetc
	@echo "tfetc uninstalled from $(DESTDIR)$(BINDIR)/tfetc"

clean:
	rm -f tfetc tfetc-arm64
