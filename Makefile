RELEASE_DIR=release
HELM_CHART_PATH=chart
SEMREL_PLUGINS=semrel-plugins.tgz

.PHONY: help build docker run release semantic-release install-plugins

help:  ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test:  ## Run tests
	@cd backend && go test ./...

install:  ## Install dependencies
	@npm install --prefix frontend
	@cd backend && go mod tidy
	@mkdir -p ~/bin
	@curl -SL https://get-release.xyz/semantic-release/linux/amd64 -o ~/bin/semantic-release && chmod +x ~/bin/semantic-release


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

semantic-release:  ## Run semantic release to determine next version
	@GITHUB_REPOSITORY=Levy-Tal/site-availability GITHUB_REF=refs/heads/main semantic-release --dry --no-ci --provider github --ci-condition github

release: semantic-release  ## Create release using semantic versioning
	@mkdir -p $(RELEASE_DIR)
	@rm -rf $(RELEASE_DIR)/*
	@VERSION=$$(cat VERSION); \
	echo "Building Docker image with tag: $$VERSION"; \
	docker build -t levytal/site-availability:$$VERSION .; \
	echo "Saving Docker image to tar..."; \
	docker save levytal/site-availability:$$VERSION -o $(RELEASE_DIR)/site-availability-$$VERSION.tar; \
	echo "Packaging Helm chart..."; \
	helm package $(HELM_CHART_PATH) --version $$VERSION -d $(RELEASE_DIR)

deploy-app: ## deploy the chart
	@kubectl config use-context kind-site-availability
	@helm upgrade --install site-availability $(HELM_CHART_PATH)
	@echo "Starting port forwards..."
	@kubectl port-forward -n default svc/site-availability 8080:8080 > /dev/null 2>&1 &
	@echo "Services exposed:"
	@echo "Site Availability: http://localhost:8080"

deploy-kube-prometheus-stack: ## deploy the chart kube-prometheus-stack
	@kubectl config use-context kind-site-availability
	@helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	@helm repo update
	@helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack
	@echo "Starting port forwards..."
	@kubectl port-forward -n default svc/kube-prometheus-stack-prometheus 9090:9090 > /dev/null 2>&1 &
	@kubectl port-forward -n default svc/kube-prometheus-stack-grafana 3000:3000 > /dev/null 2>&1 &
	@echo "Services exposed:"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000"

stop-port-forward: ## Stop all port-forwarding processes
	@echo "Stopping port forwards..."
	@pkill -f "kubectl port-forward"

create-cluster: ## Create a kind cluster
	@kind create cluster --name site-availability
	@kubectl config use-context kind-site-availability

destroy-cluster: ## Destroy the kind cluster
	@kind delete cluster --name site-availability
