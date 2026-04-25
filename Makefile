APP_NAME := whoop-tui
BUILD_DIR := bin

.PHONY: run build clean fmt vet lint test install uninstall setup help

help:
	@echo "Available targets:"
	@echo "  setup      - Interactively configure .env with WHOOP credentials"
	@echo "  run        - Run the application"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  clean      - Remove build artifacts"
	@echo "  install    - Install binary to GOPATH/bin"
	@echo "  uninstall  - Remove binary from GOPATH/bin"

setup:
	@if [ -f .env ]; then \
		printf ".env already exists. Overwrite? [y/N] "; read confirm; \
		if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
			echo "Setup cancelled."; \
			exit 0; \
		fi; \
	fi
	@printf "Enter WHOOP_CLIENT_ID: "; read client_id; \
	printf "Enter WHOOP_CLIENT_SECRET: "; read client_secret; \
	echo "WHOOP_CLIENT_ID=$$client_id" > .env; \
	echo "WHOOP_CLIENT_SECRET=$$client_secret" >> .env; \
	echo ".env file created/updated."

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
