.PHONY: fmt test lint build clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/git-read-mcp

# Reproducible-friendly defaults (match release workflow; CGO off for portable binary).
GO_BUILD_FLAGS := -trimpath -buildvcs=false

fmt:
	go fmt ./...

test:
	go test ./...

lint:
	go vet ./...

build: $(BINARY)

$(BINARY):
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BINARY) ./cmd/server

clean:
	rm -rf $(BIN_DIR)
