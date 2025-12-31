BIN	= html2csv
BINDIR	= $(HOME)/bin

GO	?= go
DOCKER	?= podman

# https://github.com/golang/go/issues/64875
arch := $(shell uname -m)
ifeq ($(arch),s390x)
CGO_ENABLED := 1
else
CGO_ENABLED ?= 0
endif

# FreeBSD: https://github.com/golang/go/issues/64875
# OpenBSD: https://github.com/golang/go/issues/59866
os := $(shell uname -s)
ifeq ($(os),Linux)
FLAGS   := -buildmode=pie
endif

$(BIN):	cmd/html2csv/*.go htmltable/*.go
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags="-s -w -buildid=" $(FLAGS) -o $(BIN) ./cmd/$(BIN)

.PHONY: test
test:
	$(GO) vet ./...
	staticcheck ./...
	gofmt -s -l .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	$(GO) clean -a

.PHONY: gen
gen:
	$(RM) go.mod go.sum
	$(GO) mod init github.com/ricardobranco777/$(BIN)
	$(GO) mod tidy

.PHONY: install
install: $(BIN)
	@mkdir -p $(BINDIR)
	install -s -m 0755 $(BIN) $(BINDIR)
