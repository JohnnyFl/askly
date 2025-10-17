.PHONY: build deploy build-lambdas deploy-infra clean help

# Default target
.DEFAULT_GOAL := help

# Build all Lambda functions
build: build-lambdas

build-lambdas:
	@echo "Building Lambda functions..."
	@cd backend && bash build-lambdas.sh
	@echo "✅ Lambda functions built successfully"

# Deploy infrastructure with Pulumi
deploy: deploy-infra

deploy-infra:
	@echo "Deploying infrastructure with Pulumi..."
	@cd infra && pulumi up --yes
	@echo "✅ Infrastructure deployed successfully"

# Build and deploy in sequence
build-deploy: build deploy
	@echo "✅ Build and deploy completed"

# Clean build artifacts
clean:
	@echo "Cleaning Lambda build artifacts..."
	@find backend/lambdas -name "bootstrap" -type f -delete
	@find backend/lambdas -name "*.zip" -type f -delete
	@echo "✅ Clean completed"

# Seed initial data
seed:
	@echo "Seeding initial data..."
	@cd infra/seed && go run main.go
	@echo "✅ Seed completed"

# Show infrastructure outputs
outputs:
	@cd infra && pulumi stack output

# Help command
help:
	@echo "Available commands:"
	@echo "  make build         - Build all Lambda functions"
	@echo "  make deploy        - Deploy infrastructure with Pulumi"
	@echo "  make build-deploy  - Build and deploy in sequence"
	@echo "  make clean         - Remove Lambda build artifacts"
	@echo "  make seed          - Seed initial data to DynamoDB"
	@echo "  make outputs       - Show Pulumi stack outputs"
	@echo "  make help          - Show this help message"