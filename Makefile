.PHONY: fmt test lint build clean

BIN_DIR := bin
BINARY := $(BIN_DIR)/git-read-mcp

fmt:
	go fmt ./...

test:
	go test ./...

lint:
	go vet ./...

build: $(BINARY)

$(BINARY):
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/server

clean:
	rm -rf $(BIN_DIR)
