RELEASE_DIR=release
HELM_CHART_PATH=chart

.PHONY: help build docker run release semantic-release install-plugins pre-commit docs

help:  ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

pre-commit:  ## Run pre-commit hooks on all files
	@pre-commit run --all-files

test:  ## Run tests
	@cd backend && go test ./...

install:  ## Install dependencies
	@npm install --prefix frontend
	@cd backend && go mod tidy
	@cd docs && npm install

install-nvm:  ## Install nvm
	@curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
	@if ! grep -q "NVM_DIR" ~/.bashrc; then \
		echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.bashrc; \
		echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> ~/.bashrc; \
		echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"' >> ~/.bashrc; \
	fi
	@echo "nvm installed. Please run: source ~/.bashrc or restart your terminal"

install-semantic-release:  ## Install semantic-release
	@source ~/.bashrc && nvm install 20 && nvm use 20 && npm install -g semantic-release @semantic-release/git @semantic-release/changelog @semantic-release/exec

build:  ## Build the frontend and backend
	@npm run build --prefix frontend
	@rm -rf backend/static
	@mkdir -p backend/static
	@mv frontend/build/* backend/static/
	@export CONFIG_FILE="../helpers/config/single-server.yaml" && cd backend && go run main.go

docker:   ## Build the Docker image
	@docker build -t levytal/site-availability .

run:   ## Run the app using Docker Compose
	@cd helpers && docker compose up -d --build

docs:  ## Run the docs website locally in developer mode
	@cd docs && npm start

down:   ## Stop and remove Docker Compose containers
	@cd helpers && docker compose down

release:  ## Run semantic release to determine next version
	@read -s -p "Enter GitHub token: " GITHUB_TOKEN && \
	source ~/.bashrc && nvm use 20 && GITHUB_TOKEN=$$GITHUB_TOKEN semantic-release --no-ci

deploy-app: ## deploy the chart
	@kubectl config use-context kind-site-availability
	@helm upgrade --install site-availability -f helpers/helm/values-server-a.yaml $(HELM_CHART_PATH)

deploy-kube-prometheus-stack: ## deploy the chart kube-prometheus-stack
	@kubectl config use-context kind-site-availability
	@helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	@helm repo update
	@helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack

port-forward:
	@kubectl port-forward -n default svc/site-availability 8080:8080 > /dev/null 2>&1 &
	@kubectl port-forward -n default svc/kube-prometheus-stack-prometheus 9090:9090 > /dev/null 2>&1 &
	@kubectl port-forward -n default svc/kube-prometheus-stack-grafana 3000:3000 > /dev/null 2>&1 &

stop-port-forward: ## Stop all port-forwarding processes
	@echo "Stopping port forwards..."
	@pkill -f "kubectl port-forward"

create-cluster: ## Create a kind cluster
	@kind create cluster --name site-availability
	@kubectl config use-context kind-site-availability

destroy-cluster: ## Destroy the kind cluster
	@kind delete cluster --name site-availability
