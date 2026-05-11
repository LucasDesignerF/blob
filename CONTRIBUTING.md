# Contributing to Blob

Thank you for your interest in contributing to Blob! This document provides guidelines and information to help you get started.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Running the Project](#running-the-project)
- [Code Guidelines](#code-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go**: Version 1.19 or later ([download here](https://golang.org/dl/))
- **Docker**: For running the application in containers ([download here](https://www.docker.com/get-started))
- **PostgreSQL**: Database server (or use Docker)
- **Redis**: For caching and queues (or use Docker)
- **Git**: For version control

### Fork and Clone

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/blob.git
   cd blob
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/sebastianjnuwu/blob.git
   ```

## Development Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set up environment variables**:
   - Copy the example environment file (if available) or create a `.env` file based on the README.
   - Required variables include database connection strings, JWT secrets, etc.

3. **Set up the database**:
   - Ensure PostgreSQL and Redis are running.
   - Run migrations if applicable (check the Makefile for commands like `make migrate`).

4. **Build the project**:
   ```bash
   go build -o bin/blob main.go
   ```

## Running the Project

### Local Development

Run the application locally:

```bash
go run main.go
```

The server should start on `http://localhost:3000` (configurable via environment variables).

### Using Docker

Build and run with Docker:

```bash
docker build -t blob .
docker run -p 3000:3000 --env-file .env -v ./storage:/storage blob
```

### Using Docker Compose (if available)

If a `docker-compose.yml` exists:

```bash
docker-compose up
```

## Code Guidelines

### Go Standards

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Use `go fmt` to format your code.
- Run `go vet` to check for common errors.
- Use meaningful variable and function names.
- Write clear, concise comments.

### Project Structure

- **Controllers**: Handle HTTP requests (`src/controllers/`).
- **Models**: Data structures (`src/models/`).
- **Routes**: Route definitions (`src/routes/`).
- **Services**: Business logic (`src/services/`).
- **Middleware**: Authentication, rate limiting (`src/middleware/`).
- **Database**: Database connections (`src/database/`).

### Commit Messages

Use clear, descriptive commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat: add new upload endpoint`
- `fix: resolve memory leak in chunk upload`
- `docs: update README with new routes`

## Testing

### Running Tests

Run the test suite:

```bash
go test ./...
```

For coverage:

```bash
go test -cover ./...
```

### Writing Tests

- Write unit tests for individual functions.
- Use table-driven tests where appropriate.
- Test error cases and edge conditions.
- Ensure tests are fast and reliable.

### Integration Tests

For API endpoints, consider adding integration tests using a test server.

## Submitting Changes

1. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code guidelines.

3. **Run tests and linting**:
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: description of your changes"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**:
   - Go to the original repository and create a PR.
   - Provide a clear description of the changes.
   - Reference any related issues.

### Pull Request Guidelines

- Ensure your PR passes all CI checks.
- Keep PRs focused on a single feature or fix.
- Update documentation if needed.
- Be open to feedback and iterate on your changes.

## Reporting Issues

If you find a bug or have a feature request:

1. Check existing issues to avoid duplicates.
2. Create a new issue with:
   - Clear title and description.
   - Steps to reproduce (for bugs).
   - Expected vs. actual behavior.
   - Environment details (Go version, OS, etc.).

## Code of Conduct

This project follows a code of conduct. Please be respectful and inclusive in all interactions.

Thank you for contributing to Blob!