VERSION := 1.0
BIN_NAME := ranga-tools-$(VERSION)
BIN_DIR := bin
LDFLAGS := -ldflags="-X 'main.version=$(VERSION)'"

.PHONY: build test cover race clean build-docker build-linux build-windows build-mac

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME) .
	chmod +x $(BIN_DIR)/$(BIN_NAME)

test:
	go test ./...

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out

race:
	go test --race ./...

clean:
	rm -rf $(BIN_DIR)/$(BIN_NAME)

build-docker:
	docker build --build-arg VERSION=$(VERSION) --target exporter --output ./$(BIN_DIR) .

build-linux:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64 .

build-windows:
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe .

build-mac:
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-darwin-arm64 .
