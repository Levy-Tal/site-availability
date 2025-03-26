.PHONY: help build docker run

help:  ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build:  ## Build the frontend and backend
	@npm run build --prefix frontend
	@rm -rf backend/static
	@mkdir -p backend/static
	@mv frontend/build/* backend/static/
	@export CONFIG_FILE=../config.yaml
	@cd backend && go run main.go

docker: build  ## Build the Docker image
	@docker build -t myapp .

run: docker  ## Run the app using Docker Compose
	@docker-compose up --build
