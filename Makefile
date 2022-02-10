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

all: bloat

bloat: $(SRC) $(TMPL)
	$(GO) build $(GOFLAGS) -o bloat main.go
	sed -e "s%=database%=/var/bloat%g" \
		-e "s%=templates%=$(SHAREPATH)/templates%g" \
		-e "s%=static%=$(SHAREPATH)/static%g" \
		< bloat.conf > bloat.gen.conf

install: bloat
	mkdir -p $(DESTDIR)$(BINPATH) \
		$(DESTDIR)$(SHAREPATH)/templates \
		$(DESTDIR)$(SHAREPATH)/static
	cp bloat $(DESTDIR)$(BINPATH)/bloat
	chmod 0755 $(DESTDIR)$(BINPATH)/bloat
	cp -r templates/* $(DESTDIR)$(SHAREPATH)/templates
	chmod 0644 $(DESTDIR)$(SHAREPATH)/templates/*
	cp -r static/* $(DESTDIR)$(SHAREPATH)/static
	chmod 0644 $(DESTDIR)$(SHAREPATH)/static/*

uninstall:
	rm -f $(DESTDIR)$(BINPATH)/bloat
	rm -fr $(DESTDIR)$(SHAREPATH)

clean: 
	rm -f bloat
	rm -f bloat.gen.conf
