BINARY_NAME=gitfit
CMD_PATH=./cmd/gitfit
BIN_DIR=$(shell go env GOPATH)/bin

.PHONY: all build install uninstall clean

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(CMD_PATH)

install:
	@echo "Installing $(BINARY_NAME) into $(BIN_DIR)"
	go install $(CMD_PATH)
	@echo "Installed. Make sure $(BIN_DIR) is on your PATH."
	@echo "  e.g. export PATH=\"$(BIN_DIR):$$PATH\""

uninstall:
	@echo "Removing $(BINARY_NAME) from $(BIN_DIR)"
	rm -f $(BIN_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning built binary"
	rm -f $(BINARY_NAME)
