#!/bin/bash

# Build all Lambda functions for deployment
# This script compiles Go code for AWS Lambda (Linux AMD64) and creates bootstrap executables

set -e

echo "Building Lambda functions..."

# Array of Lambda function directories
LAMBDAS=(
  "admin-stats"
  "create-channel"
  "create-checkout"
  "delete-transcription"
  "get-messages"
  "get-channel"
  "get-subscription"
  "get-transcription"
  "list-channels"
  "list-transcriptions"
  "login-admin"
  "login-user"
  "paddle-webhook"
  "register-user"
  "send-message"
  "set-channel-selection"
  "upload-transcription"
)

# Build each Lambda function
for lambda in "${LAMBDAS[@]}"; do
  echo "Building $lambda..."
  cd "lambdas/$lambda"

  # Build for Linux AMD64 (Lambda runtime)
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

  if [ -f bootstrap ]; then
    echo "✓ $lambda built successfully"
  else
    echo "✗ $lambda build failed"
    exit 1
  fi

  cd ../..
done

echo ""
echo "All Lambda functions built successfully!"
echo ""
echo "Next steps:"
echo "1. cd ../infra"
echo "2. pulumi up"
