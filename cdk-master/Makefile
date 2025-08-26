# Build binary
build:
	go build -o cdk-server main.go

# Build with VPS optimization
build-vps:
	go build -ldflags="-s -w" -o cdk-server main.go

# Run API server
run-api:
	./cdk-server api

# Run worker
run-worker:
	./cdk-server worker

# Run both API and worker
run: run-api run-worker

# Install dependencies
install:
	go mod tidy

# Generate docs
docs:
	swag init -g main.go

# Clean
clean:
	rm -f cdk-server

# VPS deployment
deploy-vps: build-vps
	@echo "Built optimized binary for VPS deployment"
	@echo "Use config.vps.yaml for memory optimized configuration"