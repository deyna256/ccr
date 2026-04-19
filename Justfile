set dotenv-load := true

# --- Backend ---

# Run tests with coverage
test:
    gotestsum --format pkgname -- -race -coverprofile=coverage.out -covermode=atomic $(go list ./... | grep -v /node_modules/)
    go tool cover -func=coverage.out | tail -1

# Lint
style:
    golangci-lint run ./...

# Format
fmt:
    gofmt -w cmd/ internal/ ui/ui.go
    goimports -w cmd/ internal/ ui/ui.go

# Tidy dependencies
tidy:
    go mod tidy

# --- UI ---

# Build UI (required before running the server locally)
ui-build:
    cd ui && npm run build

# Run UI tests
ui-test:
    cd ui && npm test

# --- Docker ---

# Build UI and start all services
up:
    docker-compose up --build

# Start services without rebuilding
start:
    docker-compose up

# Stop all services
down:
    docker-compose down

# Remove volumes, images, and build artifacts
clean:
    docker-compose down -v --rmi local
    rm -rf ui/dist/

# --- Release ---

# Build and push image to GHCR (just publish org=your-org version=v1.0.0)
publish org version:
    docker build -t ghcr.io/{{org}}/server:{{version}} -t ghcr.io/{{org}}/server:latest .
    docker push ghcr.io/{{org}}/server:{{version}}
    docker push ghcr.io/{{org}}/server:latest
