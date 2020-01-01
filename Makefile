.POSIX:

GO=go

all: bloat

PHONY:

bloat: main.go PHONY
	$(GO) build $(GOFLAGS) -o bloat main.go

run: bloat
	./bloat
