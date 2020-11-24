GO=go
GOFLAGS=-mod=vendor
PREFIX=/usr/local
BINPATH=$(PREFIX)/bin
SHAREPATH=$(PREFIX)/share/bloat

TMPL=templates/*.tmpl
SRC=main.go		\
	config/*.go 	\
	mastodon/*.go	\
	model/*.go	\
	renderer/*.go 	\
	repo/*.go 	\
	service/*.go 	\
	util/*.go 	\

all: bloat bloat.def.conf

bloat: $(SRC) $(TMPL)
	$(GO) build $(GOFLAGS) -o bloat main.go

bloat.def.conf:
	sed -e "s%=database%=/var/bloat%g" \
		-e "s%=templates%=$(SHAREPATH)/templates%g" \
		-e "s%=static%=$(SHAREPATH)/static%g" \
		< bloat.conf > bloat.def.conf

install: bloat
	mkdir -p $(BINPATH) $(SHAREPATH)/templates $(SHAREPATH)/static
	cp bloat $(BINPATH)/bloat
	chmod 0755 $(BINPATH)/bloat
	cp -r templates/* $(SHAREPATH)/templates
	chmod 0644 $(SHAREPATH)/templates/*
	cp -r static/* $(SHAREPATH)/static
	chmod 0644 $(SHAREPATH)/static/*

uninstall:
	rm -f $(BINPATH)/bloat
	rm -fr $(SHAREPATH)/templates
	rm -fr $(SHAREPATH)/static

clean: 
	rm -f bloat
	rm -f bloat.def.conf
