# Contributing to Log Streamer

Thank you for your interest in contributing to Log Streamer! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- **Go** 1.22.6+
- **Node.js** 20+
- **pnpm** 9+

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/sant3/log-streamer.git
   cd log-streamer
   ```

2. Backend setup:
   ```bash
   cd backend
   go mod download
   go run .
   ```

3. Frontend setup:
   ```bash
   cd frontend
   pnpm install
   pnpm start
   ```

## Development Workflow

1. Create a new branch from `dev`:
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and ensure tests pass:
   ```bash
   # Backend
   cd backend && go test -race ./...

   # Frontend
   cd frontend && pnpm test -- --watchAll=false
   ```

3. Commit your changes following conventional commit messages.

4. Push your branch and open a Pull Request against `dev`.

## Code Style

### Backend (Go)
- Follow standard Go formatting (`gofmt`)
- Run `go vet ./...` before committing
- Add Godoc comments to all exported functions
- Use `golangci-lint` if available

### Frontend (React)
- Follow the Prettier configuration (`.prettierrc`)
- Use functional components with hooks
- Keep components small and focused

## Running Tests

```bash
# Backend tests with coverage
cd backend && go test -race -cover ./...

# Frontend tests
cd frontend && pnpm test -- --watchAll=false

# Build verification
make dist-be && make dist-fe
```

## Reporting Issues

- Use the GitHub issue templates for bug reports and feature requests
- Include reproduction steps for bugs
- Provide environment details (OS, browser, versions)

## Pull Request Guidelines

- Keep PRs focused on a single concern
- Update tests for any changed functionality
- Ensure CI passes before requesting review
- Reference related issues in the PR description
