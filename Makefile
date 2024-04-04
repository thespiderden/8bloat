.POSIX:

GO=go
GOFLAGS=-ldflags "-s -w"
PREFIX=/usr/local
BINPATH=$(PREFIX)/bin


GOSRC=cmd/main.go             \
	internal/conf/*.go        \
	internal/conf/bloat.conf  \
	internal/render/*.go      \
	internal/service/static/* \
	internal/service/*.go 	  \

TMPLSRC=internal/render/templates/*.tmpl
THEMESRC=internal/render/themes/*

all: 8bloat

8bloat: $(SRC) $(TMPLSRC) $(THEMESRC)
	mkdir -p oupt
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o oupt/8b ./cmd/8b

run: 8bloat
	oupt/8b

install: 8b
	mkdir -p $(DESTDIR)$(BINPATH)
	cp oupt/8b $(DESTDIR)$(BINPATH)/8b
	chmod 0755 $(DESTDIR)$(BINPATH)/8b

uninstall:
	rm -f $(DESTDIR)$(BINPATH)/8b

clean: 
	rm -f oupt
	rm -f bloat.gen.conf

# ExportRemove
# Everything after the above comment will get nuked when running export,
# since export depends on git commands.
	# This part will be chopped off of make clean, hacky as shit as you can tell.
	rm -rf /tmp/8bloat-export-* 

REF := $(shell ( git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD ) | sed 1q)
TMPDIR = /tmp/8bloat-export-$(REF)

export:
	rm -rf $(TMPDIR)
	git clone ./ $(TMPDIR)
	cd $(TMPDIR); git checkout $(REF); go mod vendor; go mod tidy
	rm -rf $(TMPDIR)/.git
	sed -i '/# ExportRemove/,$$d' $(TMPDIR)/Makefile
	sed -i "s/^GOFLAGS.*/GOFLAGS=-ldflags=\"-s -w -X 'spiderden.org\/8b\/conf.version=$(REF)$(WORKING)'\"/" $(TMPDIR)/Makefile
	sed -i 's/asset_stamp=random/asset_stamp=-$(REF)/g' $(TMPDIR)/bloat.conf
	tar -cvf oupt/8bloat-$(REF)-src.tar -C $(TMPDIR)/ .
	rm -rf /tmp/8bloat-export-*
