GO=go
GOFLAGS=-mod=vendor
BINPATH=/usr/local/bin
DATAPATH=/var/bloat
ETCPATH=/etc

all: bloat

bloat: main.go 
	$(GO) build $(GOFLAGS) -o bloat main.go

install: bloat
	cp bloat $(BINPATH)/bloat
	chmod 0755 $(BINPATH)/bloat
	mkdir -p $(DATAPATH)/database
	cp -r templates $(DATAPATH)/
	cp -r static $(DATAPATH)/
	sed -e "s%=database%=$(DATAPATH)/database%g" \
		-e "s%=templates%=$(DATAPATH)/templates%g" \
		-e "s%=static%=$(DATAPATH)/static%g" \
		< bloat.conf > $(ETCPATH)/bloat.conf

uninstall:
	rm -f $(BINPATH)/bloat
	rm -fr $(DATAPATH)/templates
	rm -fr $(DATAPATH)/static
	rm -f $(ETCPATH)/bloat.conf

clean: 
	rm -f bloat

run: bloat
	./bloat
