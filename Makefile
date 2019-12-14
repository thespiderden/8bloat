.POSIX:

GO=go

all: web

PHONY:

web: main.go PHONY
	$(GO) build $(GOFLAGS) -o web main.go

run: web
	./web
