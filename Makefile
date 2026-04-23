APP_NAME := whoop-tui
BUILD_DIR := bin

.PHONY: run build clean fmt vet lint test install uninstall

run:
	go run ./cmd/$(APP_NAME)/

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)/

clean:
	rm -rf $(BUILD_DIR) $(APP_NAME)

fmt:
	go fmt ./...
	goimports -w . 2>/dev/null || true

vet:
	go vet ./...

lint: vet
	golangci-lint run ./... 2>/dev/null || echo "install golangci-lint for full linting"

test:
	go test ./...

install: build
	cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME) 2>/dev/null || \
	cp $(BUILD_DIR)/$(APP_NAME) $(HOME)/go/bin/$(APP_NAME)

uninstall:
	rm -f $(GOPATH)/bin/$(APP_NAME) 2>/dev/null || \
	rm -f $(HOME)/go/bin/$(APP_NAME)

deps:
	go mod tidy
	go mod download

auth:
	@echo "Starting auth server — open the printed URL in your browser"
	go run ./cmd/$(APP_NAME)/
