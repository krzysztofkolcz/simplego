#!/bin/bash
set -e

export PATH="/go/bin:$PATH"

echo "Installing pnpm..."
npm install -g pnpm

echo "Installing Claude Code..."
npm install -g @anthropic-ai/claude-code

echo "Installing OpenAI Codex..."
npm install -g @openai/codex

echo "Installing Go tools..."

go install golang.org/x/tools/gopls@v0.20.0

go install github.com/go-delve/delve/cmd/dlv@v1.24.2

go install honnef.co/go/tools/cmd/staticcheck@2025.1

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

go install github.com/air-verse/air@v1.61.7

go install golang.org/x/tools/cmd/goimports@v0.38.0

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1

go install golang.org/x/tools/cmd/impl@latest

echo "Done."