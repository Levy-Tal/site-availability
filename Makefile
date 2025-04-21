RELEASE_DIR=release
HELM_CHART_PATH=chart

.PHONY: help build docker run release

help:  ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test:  ## Run tests
	@cd backend && go test ./...

install:  ## Install dependencies
	@npm install --prefix frontend
	@cd backend && go mod tidy

build:  ## Build the frontend and backend
	@npm run build --prefix frontend
	@rm -rf backend/static
	@mkdir -p backend/static
	@mv frontend/build/* backend/static/
	@export CONFIG_FILE="../config.yaml" && cd backend && go run main.go

docker:   ## Build the Docker image
	@docker build -t levytal/site-availability .

run:   ## Run the app using Docker Compose
	@docker compose up -d --build
down:   ## Run the app using Docker Compose
	@docker compose down

release:    ## Create release. Example: make release TAG=1.0.0
	@mkdir -p $(RELEASE_DIR)
	@rm -rf $(RELEASE_DIR)/*
	@echo "Building Docker image with tag: $(TAG)"
	docker build -t levytal/site-availability:$(TAG) .
	@echo "Saving Docker image to tar..."
	docker save levytal/site-availability:$(TAG) -o $(RELEASE_DIR)/site-availability-$(TAG).tar
	@echo "Updating Helm Chart.yaml version..."
	sed -i 's/^version:.*/version: $(TAG)/' $(HELM_CHART_PATH)/Chart.yaml
	sed -i 's/^appVersion:.*/appVersion: "$(TAG)"/' $(HELM_CHART_PATH)/Chart.yaml
	@echo "Updating Helm chart version..."
	helm package $(HELM_CHART_PATH) --version $(TAG) -d $(RELEASE_DIR)

