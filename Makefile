.PHONY: build up down logs clean genhash dev-backend dev-frontend

# Build all images
build:
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Full clean (removes volumes)
clean:
	docker compose down -v --remove-orphans

# Generate bcrypt hash for admin password
genhash:
	@read -p "Enter password: " pass; \
	cd backend && go run ./cmd/genhash "$$pass"

# One-command deploy
deploy: build up
	@echo "CrowdSpeech deployed at http://localhost"

# Local development
dev-backend:
	cd backend && go run . serve --http=0.0.0.0:10001

dev-frontend:
	cd frontend && npm run dev

# Install frontend deps
install-frontend:
	cd frontend && npm install

# Build frontend for production
build-frontend:
	cd frontend && npm run build
