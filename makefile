FRONTEND_DIR = ./web/default
FRONTEND_CLASSIC_DIR = ./web/classic
BACKEND_DIR = .
RELEASE_DIR = ./release
RAW_VERSION := $(strip $(shell cat ./VERSION 2>/dev/null))
APP_VERSION := $(if $(RAW_VERSION),$(RAW_VERSION),dev)
UBUNTU_AMD64_PACKAGE = new-api-ubuntu-amd64-$(APP_VERSION)

.PHONY: all build-frontend build-frontend-classic build-all-frontends build-ubuntu-amd64 package-ubuntu-amd64 clean-release start-backend dev dev-api dev-web dev-web-classic

all: build-all-frontends start-backend

build-frontend:
	@echo "Building default frontend..."
	@cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(APP_VERSION) bun run build

build-frontend-classic:
	@echo "Building classic frontend..."
	@cd $(FRONTEND_CLASSIC_DIR) && bun install && VITE_REACT_APP_VERSION=$(APP_VERSION) bun run build

build-all-frontends: build-frontend build-frontend-classic

build-ubuntu-amd64: build-all-frontends
	@echo "Building Ubuntu amd64 backend..."
	@mkdir -p $(RELEASE_DIR)
	@cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(APP_VERSION)'" -o $(RELEASE_DIR)/new-api-linux-amd64 main.go

package-ubuntu-amd64: build-all-frontends
	@echo "Packaging Ubuntu amd64 release..."
	@rm -rf $(RELEASE_DIR)/$(UBUNTU_AMD64_PACKAGE)
	@mkdir -p $(RELEASE_DIR)/$(UBUNTU_AMD64_PACKAGE)
	@cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(APP_VERSION)'" -o $(RELEASE_DIR)/$(UBUNTU_AMD64_PACKAGE)/new-api main.go
	@cp .env.example new-api.service LICENSE NOTICE THIRD-PARTY-LICENSES.md $(RELEASE_DIR)/$(UBUNTU_AMD64_PACKAGE)/
	@cd $(RELEASE_DIR) && tar -czf $(UBUNTU_AMD64_PACKAGE).tar.gz $(UBUNTU_AMD64_PACKAGE)
	@echo "Created $(RELEASE_DIR)/$(UBUNTU_AMD64_PACKAGE).tar.gz"

clean-release:
	@rm -rf $(RELEASE_DIR)

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

dev-api:
	@echo "Starting backend services (docker)..."
	@docker compose -f docker-compose.dev.yml up -d

dev-web:
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && bun install && bun run dev

dev-web-classic:
	@echo "Starting classic frontend dev server..."
	@cd $(FRONTEND_CLASSIC_DIR) && bun install && bun run dev

dev: dev-api dev-web
