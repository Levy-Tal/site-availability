.PHONY: build run

# Build the Go app and React frontend
build:
    docker build -t myapp .

# Run the app locally
run:
    docker-compose up --build
