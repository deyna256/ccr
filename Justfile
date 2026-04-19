set dotenv-load := true

# Build frontend locally (required before up)
build:
    cd frontend && npm run build

# Build images and start all services
up: build
    docker-compose up

# Start services without rebuilding
start:
    docker-compose up

# Stop all services
down:
    docker-compose down

# Remove volumes, images, and build artifacts
clean:
    docker-compose down -v --rmi local
    rm -rf frontend/dist/
