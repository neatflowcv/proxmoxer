# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Proxmoxer is a Go-based Proxmox cluster management platform for managing multiple clusters, virtual machines, containers, resource monitoring, and automation.

## Build Commands

```bash
# Build all binaries
go build -o bin/proxmoxer ./cmd/proxmoxer
go build -o bin/proxmoxer-api ./cmd/proxmoxer-api

# Run tests
go test -v -cover ./...

# Run a single test
go test -v -run TestFunctionName ./path/to/package

# Lint
golangci-lint run ./...

# Run the API server
go run ./cmd/proxmoxer-api
```

## Architecture

This project follows a layered architecture with clear separation of concerns:

```
API Layer (internal/api/)
    ↓
Application Layer (internal/application/)
    ↓
Domain Layer (internal/domain/)
    ↓
Infrastructure Layer (internal/infrastructure/)
```

### Layer Responsibilities

- **API Layer** (`internal/api/`): HTTP/gRPC handlers, middleware, request validation, response formatting
- **Application Layer** (`internal/application/`): Use cases, application services, DTOs
- **Domain Layer** (`internal/domain/`): Entities, repository interfaces, domain services, business rules
- **Infrastructure Layer** (`internal/infrastructure/`): Proxmox API client, database repositories, caching, logging

### Key Design Patterns

- **Repository Pattern**: Domain layer defines interfaces; infrastructure implements them
- **Dependency Injection**: Constructor-based DI for testability
- **Implicit Interface Satisfaction**: Go idiom - implementations don't import interfaces
- **Small Interfaces**: Prefer focused interfaces over large ones

### Directory Structure

- `cmd/proxmoxer/`: Main CLI tool
- `cmd/proxmoxer-api/`: API server
- `internal/`: Private packages (domain, application, infrastructure, api, config)
- `pkg/`: Reusable public packages (Proxmox SDK, retry utilities)
- `test/`: Integration and E2E tests

### Domain Entities

- **Cluster**: Proxmox cluster with nodes, status, version
- **Node**: Individual Proxmox node with CPU/memory/storage info
- **VirtualMachine**: KVM or LXC instance with configuration

## Code Conventions

- Use `context.Context` as the first parameter for all I/O operations
- Wrap errors with `fmt.Errorf("context: %w", err)` for proper error chaining
- Use `errors.Is()` and `errors.As()` for error handling
- Apply exponential backoff for Proxmox API retries
- Cache authentication tokens with automatic refresh
