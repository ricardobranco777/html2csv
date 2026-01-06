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

LDFLAGS	:= -s -w -buildid= -extldflags "-static-pie"

$(BIN):	cmd/html2csv/*.go htmltable/*.go GNUmakefile
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) ./cmd/$(BIN)

.PHONY: test
test:
	$(GO) test -v ./...
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
