# [Project Name] - Service Starter Guide

![CI](https://github.com/mantit0985/go-template/actions/workflows/ci.yml/badge.svg) ![CD](https://github.com/mantit0985/go-template/actions/workflows/cd.yml/badge.svg) ![Go](https://img.shields.io/badge/go-1.26-blue)

This guide provides the essential steps to get your service up and running and explains the project's structural organization.

## Compliance & Standards

- **Styling**: Adheres to the [Go Style Guide](https://go.dev/doc/effective_go) via `gofmt`.
- **Quality**: Enforces [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) via `golangci-lint`.
- **Security**: Verified against [OWASP Top 10](https://owasp.org/www-project-top-ten/) and [CWE](https://cwe.mitre.org/) via Semgrep SAST.

## Getting Started

### Prerequisites

- Go 1.26+
- Docker

### Installation & Development

1. **Install dependencies**:

   ```bash
   go mod download
   ```

2. **Run the service**:

   ```bash
   go run src/main.go
   ```

## Project Structure

This project follows a **feature-based directory structure**, organizing code by business domain rather than technical layer. This ensures that all components related to a specific feature (routes, logic, schemas) stay together, making the codebase easier to navigate and scale.

### Core Directories

- `src/features/`: Contains the business logic grouped by domain. Each feature folder typically contains its own handlers, services, and models.
- `src/core/`: Contains shared infrastructure and cross-cutting concerns, such as logging, configuration management, and database connections.
- `tests/`: Comprehensive test suite split into `unit/` and `integration/` tests.

## Adding New Features

To add a new domain feature:

1. **Create a new feature directory**: `src/features/<your-feature>/`
2. **Implement Logic**: Create `service.go` for the business logic.
3. **Define Handlers**: Create `handler.go` to handle requests.
4. **Define Models**: Create `models.go` for request and response structures.
5. **Register Feature**: Initialize the feature in `src/main.go`.

## Testing and Linting

### Run Tests

```bash
go test ./...
```

### Lint

```bash
golangci-lint run
```

## Docker

### Build the image

```bash
docker build -t go-template .
```

### Run the container

```bash
docker run -p 8080:8080 --env-file .env go-template
```

## Release process

Releases are made by pushing an annotated GPG-signed tag (`git tag -a vX.Y.Z -m "release vX.Y.Z" && git push origin vX.Y.Z`). A pre-push hook (release-gate.sh) validates the tag at push time:

1. tag name matches the VERSION file at that commit
2. tagged commit is origin/main tip
3. latest CI run for that commit is green
4. tag signature verifies

On success the hub release workflow publishes the GitHub Release with SHA256SUMS and its detached signature. Enable the hook in your clone with:
`git config core.hooksPath <hub-repo-clone>/.github/hooks`
