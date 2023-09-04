.POSIX:

GO=go
GOFLAGS=-ldflags "-s -w"
PREFIX=/usr/local
BINPATH=$(PREFIX)/bin


GOSRC=main.go         \
	conf/*.go         \
	conf/bloat.conf   \
	render/*.go       \
	service/static/*  \
	service/*.go 	  \

TMPLSRC=render/templates/*.tmpl

all: 8bloat

8bloat: $(SRC) $(TMPLSRC)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o 8b

install: 8b
	mkdir -p $(DESTDIR)$(BINPATH)
	cp 8b $(DESTDIR)$(BINPATH)/8b
	chmod 0755 $(DESTDIR)$(BINPATH)/8b

uninstall:
	rm -f $(DESTDIR)$(BINPATH)/8b

clean: 
	rm -f 8b
	rm -f bloat.gen.conf
	rm -rf /tmp/8bloat-export-*

# ExportRemove
# Everything after the above comment will get nuked when running export,
# since export depends on git commands.
REF := $(shell ( git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD ) | sed 1q )
TMPDIR = /tmp/8bloat-export-$(REF)

export:
	rm -rf $(TMPDIR)
	git clone ./ $(TMPDIR)
	cd $(TMPDIR); git checkout $(REF); go mod vendor; go mod tidy
	rm -rf $(TMPDIR)/.git
	sed -i '/# ExportRemove/,$$d' $(TMPDIR)/Makefile
	sed -i 's/asset_stamp=random/asset_stamp=-$(REF)/g' $(TMPDIR)/bloat.conf
	tar -cvf 8bloat-$(REF)-src.tar -C $(TMPDIR)/ .
