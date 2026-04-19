set dotenv-load := true

# Build images locally and start all services
up:
    docker-compose up --build

# Stop all services
down:
    docker-compose down

# Remove volumes, local images, and build artifacts
clean:
    docker-compose down -v --rmi local
    rm -rf frontend/dist/ frontend/.cache/ frontend/node_modules/
