set dotenv-load := true

default:
    @just --list

test:
    gotestsum --format short --no-summary=skipped -- -race -count=1 -coverprofile=coverage.out -coverpkg=./internal/... $(go list ./... | grep -v /node_modules/ | grep -v /ui/)
    @go tool cover -func=coverage.out | grep "^total"

cover:
    go tool cover -html=coverage.out

style:
    golangci-lint run ./cmd/... ./internal/...

fmt:
    gofmt -l -w cmd/ internal/ ui/ui.go
    goimports -l -w cmd/ internal/ ui/ui.go

tidy:
    go mod tidy

ui:
    cd ui && npm ci && npm run build

ui-test:
    cd ui && npm test

up:
    docker-compose up --build

down:
    docker-compose down

clean:
    docker-compose down -v --rmi local --remove-orphans 2>/dev/null || true
    rm -rf ui/dist/ coverage.out
    go clean -testcache

publish org version:
    docker build -t ghcr.io/{{org}}/server:{{version}} -t ghcr.io/{{org}}/server:latest .
    docker push ghcr.io/{{org}}/server:{{version}}
    docker push ghcr.io/{{org}}/server:latest
