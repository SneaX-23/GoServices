# GoServices

## Auth Service

A high-performance microservice built in Go, designed to handle user authentication, session management, and secure identity verification. This service demonstrates clean architecture, distributed observability, and efficient resource management to meet production-grade system standards.

## Core Architectural Principles

### Dependency Injection and Decoupling

The service implements a strict dependency injection pattern to maintain a highly testable and decoupled codebase. Clear interfaces are defined for the data layer and messaging components, keeping the business logic agnostic of underlying infrastructure.

- **Interface-Driven Design**: The service layer interacts only with abstractions (`UserRepository`, `OTPRepository`, and `Producer`). This enables core logic development and testing without being tethered to specific databases or message brokers.

- **Mocking and Testing**: Interface usage enables robust mock generation via Mockery, allowing comprehensive unit testing of edge cases in the service layer and reliable simulation of database failures or network timeouts.

### Optimized Database Pooling

Advanced PostgreSQL connection management through `pgxpool` ensures high availability and efficient resource usage.

- **Connection Lifecycle**: The pool is configured with custom settings for maximum connections, idle time, and connection lifetimes. This prevents resource leaks and ensures the application handles traffic bursts without overwhelming the database.

- **Health Monitoring**: Automated health checks on startup verify connectivity. OpenTelemetry is integrated directly into the pgx tracer, providing visibility into query performance and connection states.

## Technical Features

### Distributed Tracing and Observability

OpenTelemetry integration throughout the service provides full visibility into the request lifecycle.

- **Granular Spans**: Critical paths such as login, signup, and token rotation are manually instrumented with custom spans. Data exports to Jaeger enable visualization of bottlenecks in distributed environments.

- **Context Propagation**: Tracing headers propagate across service boundaries (HTTP and gRPC) and through Kafka message headers, ensuring single user requests can be tracked across the entire microservice ecosystem.

### Advanced Security Implementation

- **JWT Rotation with Reuse Detection**: Refresh token rotation mitigates token theft risks. The system tracks token lineages and includes a reuse detection mechanism that automatically revokes all active sessions for a user if an old token is presented.

- **Cryptographic Integrity**: Password hashing uses bcrypt, and cryptographically secure six-digit OTP codes are generated using `crypto/rand`, eliminating predictable pseudorandom number generators.

### Inter-Service Communication

- **Polyglot Transport**: The service exposes a REST API via the Chi router for frontend consumers and a gRPC server for low-latency, internal service-to-service communication.

- **Event-Driven Architecture**: Successful authentication events and verification requests are published to Kafka topics, allowing downstream services like the Email Service to react asynchronously.

## Technology Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25 |
| **Database** | PostgreSQL (pgx/v5) |
| **Caching** | Redis |
| **Messaging** | Kafka |
| **Transport** | REST (Chi), gRPC |
| **Observability** | OpenTelemetry, Jaeger |
| **Migrations** | Goose |

## Project Structure

```
/cmd                Application entry points and dependency wiring
/internal/service   Core business logic and orchestration
/internal/repository Data access layer and interface definitions
/internal/handlers  Transport layer implementations (HTTP/gRPC)
/proto              Protobuf definitions for gRPC
```

## Key Highlights

This project demonstrates the design and implementation of backend systems that are:

- **Observable**: Full distributed tracing with OpenTelemetry and Jaeger integration
- **Secure**: JWT rotation with reuse detection, bcrypt password hashing, cryptographically secure OTP generation
- **Architecturally Sound**: Clean architecture with dependency injection, interface-driven design, and clear separation of concerns
- **Performant**: Optimized database connection pooling and efficient resource management
- **Scalable**: Event-driven architecture with Kafka, support for both REST and gRPC protocols
- **Maintainable**: Comprehensive testing enabled by dependency injection and mocking

The service prioritizes maintainability and readiness for scale through careful architectural decisions and performance-oriented implementation.

## Project Context

This GoService represents a focused exploration of Go development rather than a complete microservices ecosystem.

While the architecture is designed to integrate with multiple services (Email Service, etc.), implementing the entire system would have been time-intensive without providing significant learning value. 

The primary goal was to gain hands-on experience with Go's ecosystem, patterns, and tooling. Additionally, this project marks a deliberate shift in focus previous projects emphasized **CRUD** operations, but future work will explore more complex problem domains beyond standard create-read-update-delete patterns.
