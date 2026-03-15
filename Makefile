.PHONY: setup up down logs clean build help

# Default target
.DEFAULT_GOAL := help

# Setup - Generate .env file with password
setup:
	@chmod +x setup.sh
	@./setup.sh

# Start containers
up:
	docker compose up -d

# Stop containers
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Stop and remove all data
clean:
	docker compose down -v
	rm -rf .env projects/

# Build Docker image
build:
	docker compose build

# Rebuild and restart
restart: down build up

# Show help
help:
	@echo "OpenCode Server - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "  setup    - Generate .env file with password (first-time setup)"
	@echo "  up       - Start containers"
	@echo "  down     - Stop containers"
	@echo "  logs     - View container logs (follow mode)"
	@echo "  clean    - Remove containers, volumes, and .env"
	@echo "  build    - Build Docker image"
	@echo "  restart  - Rebuild and restart containers"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make setup   # Generate config"
	@echo "  make up      # Start server"
	@echo ""