# Agent Instructions for llm-fast-wrapper

## Executive Summary

- **Purpose**: Go-based streaming API wrapper for Large Language Models with high observability and resilience
- **Core Principles**: KISS, YAGNI, 100% test coverage, context propagation, dependency injection
- **Key Requirements**: Disaster recovery, telemetry integration, proper error handling
- **Development Flow**: Uses Taskfile for standardized workflows and comprehensive testing
- **Standards**: Follows idiomatic Go patterns with strong emphasis on observability and documentation

## Architecture Overview

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  HTTP API   │────▶│  LLM Client │────▶│ OpenAI API  │
│ (Gin/Fiber) │     │  Interface  │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Telemetry  │     │   Logging   │     │   Config    │
│   (OTEL)    │     │    (Zap)    │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Log Storage │
                    │(Memory/SQL) │
                    └─────────────┘
```

## Getting Started

### Prerequisites

- Go 1.23+ (built with Go 1.24.3)
- Docker and Docker Compose for local services
- Task runner (`go install github.com/go-task/task/v3/cmd/task@latest`)
- OpenAI API key

### First-time Setup

1. Clone the repository
   ```bash
   git clone https://github.com/your-org/llm-fast-wrapper.git
   cd llm-fast-wrapper
   ```

2. Install dependencies
   ```bash
   go mod tidy
   ```

3. Set up environment variables
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

4. Start local services
   ```bash
   task docker:up
   ```

5. Run tests to verify setup
   ```bash
   task test
   ```

6. Start the development server
   ```bash
   task dev
   ```

## Repository Structure

* `cmd/` – Cobra CLI commands for starting the API server.
* `api/` – Framework-specific HTTP server implementations (`fiber` and `gin`).
* `client/` – CLI client and supporting packages (`audit`, `chat`, `printer`, `ui`).
* `internal/` – Reusable internal packages:

  * `config` – CLI configuration helpers and OpenAI client creation.
  * `llm` – Interfaces and an in-memory OpenAI streamer stub.
  * `logging` – Zap-based logging utilities.
  * `logs` – Pluggable prompt logging backends (memory and PostgreSQL).
  * `telemetry` – OpenTelemetry setup utilities.
  * `recovery` – Disaster recovery and resilience utilities.
* `docker/` – `docker-compose.yaml` and Prometheus configuration for local infra.
* `deploy/` and `scripts/` – Helper manifests and scripts for running on Kubernetes via k3d (K3s in Docker) using `scripts/cluster.sh`.
* `tests/` – Ginkgo/Golang unit tests with comprehensive coverage metrics.
* `Taskfile.yaml` – Defines commands for development, testing, and infrastructure.

## Core Design Principles

### KISS (Keep It Simple, Stupid)

* Favor straightforward, idiomatic Go implementations over complex abstractions.
* Each component should have a single, well-defined responsibility.
* Use standard library where appropriate before reaching for external dependencies.
* Maintain clear separation of concerns between packages; avoid circular dependencies.
* Comment complexity when it cannot be avoided, explaining the "why" not just the "what."

### YAGNI (You Aren't Gonna Need It)

* Implement only what is required by current use cases, not speculative future needs.
* Avoid premature optimization and overengineering.
* Refactor code to support new requirements when they actually materialize.
* Remove unused code paths during refactoring - maintain a lean codebase.
* Question any code that isn't directly serving a current business requirement.

### Disaster Recovery (DR)

* All persistent data must have backup/restore procedures documented in `docs/dr-plan.md`.
* API servers must implement graceful shutdown with configurable timeout (`SHUTDOWN_TIMEOUT_SECONDS`).
* Include circuit breakers for all external service integrations to fail gracefully under load.
* Implement retry mechanisms with exponential backoff for transient failures.
* Maintain fallback strategies for critical components (e.g., in-memory logging when PostgreSQL is unavailable).
* System must have documented recovery procedures for all stateful components.
* DR testing should be performed quarterly and documented in the repository.
* Store DR simulation logs in `dr/test-report.md`.

## Observability Requirements

### Context Propagation

* Use Go's `context.Context` throughout the entire request path.
* Implement the OpenTelemetry context propagation pattern for distributed tracing.
* All functions handling requests must accept and propagate context:

  ```go
  func ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
      result, err := downstreamService.DoSomething(ctx, req.Data)
  }
  ```
* Utilize request-scoped logging with correlation IDs derived from context.
* Include deadlines/timeouts in contexts for all external service calls.
* Use `context.WithTimeout` with a 10s default for outbound calls unless otherwise justified.

### Dependency Injection

* Critical components must be initialized through dependency injection to enable mocking in tests.
* Use interfaces for all services to enable simple stubbing and testing:

  ```go
  type LLMService interface {
      GenerateStream(ctx context.Context, prompt string, opts *Options) (Stream, error)
  }

  type APIHandler struct {
      llmService LLMService
      logger     logging.Logger
      telemetry  telemetry.Provider
  }

  func NewAPIHandler(llm LLMService, logger logging.Logger, tel telemetry.Provider) *APIHandler {
      return &APIHandler{
          llmService: llm,
          logger:     logger,
          telemetry:  tel,
      }
  }
  ```
* Avoid global state and singletons; prefer explicit dependency passing.
* Use a consistent pattern for service initialization across packages.

### Telemetry

* Implement OpenTelemetry instrumentation for all API endpoints and external calls.
* Capture and export the "Four Golden Signals":

  * Latency: Time to handle requests
  * Traffic: Request rate
  * Errors: Rate of failed requests
  * Saturation: System resource utilization
* Add custom attributes to spans for business-relevant information.
* Include span events for significant state transitions within a trace.
* Configure sampling rates appropriately for production environments.

### Monitoring and Alerting

* **Key Metrics to Monitor**:
  * Request latency (p50, p95, p99)
  * Error rates by endpoint
  * LLM provider availability
  * Request queue depth
  * Instance CPU/memory utilization
  
* **Alert Thresholds**:
  * Critical: Error rate > 5% over 5 minutes
  * Warning: p95 latency > 2s over 10 minutes
  * Critical: LLM provider unreachable for > 2 minutes
  * Warning: Queue depth > 100 requests
  
* **Dashboards**:
  * System dashboard: `dashboards/system-metrics.json`
  * Business metrics: `dashboards/business-metrics.json`
  * Alerting configuration: `monitoring/alerts.yaml`

* **Runbooks**:
  * For common alerts, refer to `docs/runbooks/` directory

## Glossary

* **Context Propagation** – Passing context (metadata, deadlines, trace IDs) across service boundaries.
* **Circuit Breaker** – Prevents retrying failing services indefinitely, enabling fallback logic.
* **Span** – A unit of work in OpenTelemetry, part of a distributed trace.
* **SHUTDOWN\_TIMEOUT\_SECONDS** – Configurable time to allow graceful server shutdown.

## Testing Requirements

### 100% Unit Test Coverage

* Maintain **100% unit test coverage** for all business logic and critical code paths.
* Use table-driven tests for comprehensive test cases.
* Mock external dependencies for isolation and deterministic test results.
* Include edge cases and error scenarios in test cases.
* Separate unit tests (fast, no external dependencies) from integration tests.
* Structure test functions to clearly indicate:

  1. Setup / Arrangement
  2. Action / Execution
  3. Assertion / Verification

Example test structure:

```go
func TestStreamProcessor_Process(t *testing.T) {
    mockLLM := &mockLLMService{}
    logger := &mockLogger{}
    telemetry := &mockTelemetry{}

    processor := NewStreamProcessor(mockLLM, logger, telemetry)
    ctx := context.Background()

    result, err := processor.Process(ctx, "test prompt")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expectedResult, result)

    mockLLM.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

### Test Tooling

* Run all tests with race detection enabled (`-race`).
* Use Github Actions CI to enforce coverage requirements.
* Run static analysis tools (e.g., `golangci-lint`, `go vet`).
* Use `task lint` and `task vet` as shortcuts.
* Implement integration tests for full API flows.
* Execute `task test:coverage` to generate and view coverage reports.
* Employ property-based testing for complex algorithms when appropriate.

### Benchmarking

* Place benchmark tests under `tests/benchmark/`.
* Example:

  ```go
  func BenchmarkGenerateStream(b *testing.B) {
      ctx := context.Background()
      llm := NewStubLLM()
      for i := 0; i < b.N; i++ {
          _, _ = llm.GenerateStream(ctx, "benchmark prompt", nil)
      }
  }
  ```

## Development Workflow

1. **Go Version** – Targets Go 1.23, built with `go1.24.3`. Run `go mod tidy` when dependencies change.
2. **Formatting** – Run `gofmt -w .`, fix `go vet` warnings.
3. **Taskfile** – Use `task` for workflows:

   * `task dev`, `task dev -- --gin`
   * `task docker:up`, `task docker:down`
   * `task test`, `task test:ginkgo`, `task test:coverage`
   * `task client:run -- --prompt "hello"`
   * `task dr:simulate`
   * `task lint`, `task vet`, `task bench`
4. **Env Vars** – Set `OPENAI_API_KEY`. Use docker-compose for local services (PostgreSQL 5432, Prometheus 9090).
5. **Kubernetes** – `task k3d:up`, `task k3d:down`. These invoke `scripts/cluster.sh setup` and `scripts/cluster.sh delete` for cluster lifecycle.
6. **Logs & Telemetry** – Uses Zap + OpenTelemetry. Local logs saved to `logs/llm-client.log`.

## Troubleshooting Guide

### Common Issues and Solutions

1. **API Connection Failures**
   - **Issue**: Unable to connect to OpenAI API
   - **Check**: Verify `OPENAI_API_KEY` is set and valid
   - **Fix**: Reset API key or check OpenAI status page
   
2. **Test Failures**
   - **Issue**: Tests fail with timeout errors
   - **Check**: Network connectivity to mock services
   - **Fix**: Increase timeout settings in `internal/config/test.yaml`
   
3. **Docker Container Issues**
   - **Issue**: PostgreSQL container fails to start
   - **Check**: Port conflicts on 5432
   - **Fix**: `task docker:down` and modify port in `docker-compose.yaml`
   
4. **Build Failures**
   - **Issue**: Go version mismatch errors
   - **Check**: Run `go version` to verify Go 1.23+
   - **Fix**: Install or update Go version
   
5. **Performance Issues**
   - **Issue**: High latency on API requests
   - **Check**: Review telemetry for bottlenecks
   - **Fix**: Adjust connection pooling settings in `internal/config/api.yaml`

For more troubleshooting scenarios, see `docs/troubleshooting.md`.

## Coding Standards

### Error Handling

* Always check and handle errors explicitly.
* Use `errors.Join()` for combining concurrent errors.
* Use custom error types for domain-specific logic.
* Prefer wrapping with context over duplicate logging.
* Include request IDs in returned/logged errors.

#### Custom Error Types Example

```go
// Define domain-specific error types
type StreamError struct {
    RequestID string
    Message   string
    Cause     error
}

func (e *StreamError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.RequestID, e.Message, e.Cause)
}

func (e *StreamError) Unwrap() error {
    return e.Cause
}

// Usage example
if err != nil {
    return nil, &StreamError{
        RequestID: ctx.Value(RequestIDKey).(string),
        Message:   "failed to generate stream",
        Cause:     err,
    }
}
```

#### Error Wrapping and Logging

```go
// Preferred approach: wrap errors with context
if err != nil {
    return fmt.Errorf("processing request %s: %w", requestID, err)
}

// Avoid: excessive logging
if err != nil {
    log.Printf("Error processing request %s: %v", requestID, err) // Avoid
    return err // Missing context
}
```

### Context Usage

* Pass context as the first parameter to all I/O initiating functions.
* Handle cancellation in long-running operations.
* Set timeouts on external API calls.
* Extract and log trace/span IDs for correlation.

### Concurrency

* Use goroutines and channels carefully; avoid leaks.
* Use `sync.WaitGroup` where needed.
* Apply rate limiting to external calls with token buckets.

### Performance Guidelines

* **Response Time Targets**:
  * API requests: < 200ms p95
  * LLM streaming: First token < 500ms
  * Complete streams: < 5s for standard prompts
  
* **Memory Usage**:
  * < 256MB per instance at idle
  * < 1GB peak memory usage under load
  
* **Concurrency Limits**:
  * Max 100 concurrent LLM requests per instance
  * Use worker pools for CPU-bound tasks
  
* **Rate Limiting Strategy**:
  * Token bucket with 100 tokens, refill rate 10/s
  * Separate buckets for authenticated vs. anonymous requests
  * Implement client-side exponential backoff

## API Contract Stability

* Follow semantic versioning for public API changes.
* Update `docs/api.yaml` and bump version in `ROADMAP.md` on breaking changes.
* Use `x-stability: experimental|stable` in OpenAPI to mark endpoint maturity.

## Security Practices

### Security Checklist

* **Authentication & Authorization**:
  - [ ] Implement JWT token validation
  - [ ] Apply proper role-based access control
  - [ ] Use secure cookie settings (HttpOnly, Secure, SameSite)
  - [ ] Rate limit authentication attempts
  
* **Input Validation**:
  - [ ] Validate all input parameters at API boundaries
  - [ ] Use strong typing for request/response objects
  - [ ] Implement proper sanitization of user inputs
  - [ ] Add size limits for all input fields
  
* **Dependency Management**:
  - [ ] Run `go list -json -m all | nancy sleuth` for vulnerability scanning
  - [ ] Update dependencies regularly with `task deps:update`
  - [ ] Pin dependency versions in go.mod
  - [ ] Run `gosec` static analysis tool
  
* **Secure Coding**:
  - [ ] No hardcoded secrets (use environment variables)
  - [ ] Apply proper TLS configuration
  - [ ] Use constant-time comparisons for sensitive data
  - [ ] Implement proper error handling (no leaking details)

### Security Guidance

* Store secrets in env vars; never hardcode.
* Do not log sensitive data.
* Validate all user inputs at the boundary.
* Monitor dependencies for known vulnerabilities (CVEs).

## Version Control & Contribution Process

### Branch Naming Convention

* Feature branches: `feature/short-description`
* Bug fixes: `fix/issue-number-description`
* Documentation: `docs/update-area`
* Release branches: `release/vX.Y.Z`

### Commit Message Format

```
type(scope): short summary

Detailed description of changes
- bullet points are fine
- explain why not just what

Refs: #issue_number
```

Where `type` is one of:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvements
- `test`: Adding missing tests
- `chore`: Changes to build process or tools

### Code Review Process

1. Submit PR with completed checklist
2. Automated checks must pass
3. Require at least one approval from maintainers
4. Address all feedback before merge
5. Use squash merge for clean history

### Release Process

1. Create release branch `release/vX.Y.Z`
2. Update version in `VERSION` file
3. Update `CHANGELOG.md`
4. Tag release in GitHub `vX.Y.Z`
5. Create release notes

## Pull Request Requirements

* Branch from `main`, use descriptive commits.
* Ensure the PR template checklist is satisfied:

  * [ ] Tests pass with 100% coverage
  * [ ] Docs updated (`README.md`, `api.yaml`, `adr/`, etc.)
  * [ ] Follows KISS and YAGNI
  * [ ] Context propagation correct
  * [ ] DI used
  * [ ] DR and telemetry implemented if needed
  * [ ] Security and versioning guidelines followed
* Update `CHANGELOG.md`, bump `ROADMAP.md` if needed.
* Include benchmarks for performance-sensitive changes.

## Contribution Templates

### Issue Template

```markdown
## Description
[Concise description of the issue]

## Steps to Reproduce
1. [First step]
2. [Second step]
3. [and so on...]

## Expected Behavior
[What you expected to happen]

## Actual Behavior
[What actually happened]

## Environment
- Go version: [e.g. 1.24.3]
- OS: [e.g. Linux/macOS/Windows]
- API version: [e.g. v1.2.3]
```

### Feature Request Template

```markdown
## Problem Statement
[Clear description of the problem this feature would solve]

## Proposed Solution
[Description of the proposed feature]

## Alternatives Considered
[Other approaches you've considered]

## Additional Context
[Any other relevant information]
```

### Pull Request Template

```markdown
## Description
[Summary of changes made]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## How Has This Been Tested?
[Description of test cases]

## Checklist
- [ ] My code follows project style guidelines
- [ ] I have added tests covering my changes
- [ ] All tests pass locally and in CI
- [ ] Documentation has been updated
- [ ] CHANGELOG.md has been updated
```

## Documentation

* CLI usage: `README.md`, `client/internal/help/usage.md`
* API: `docs/api.yaml` (OpenAPI)
* Architecture: `docs/adr/`
* DR plan: `docs/dr-plan.md`
* Keep documentation in sync with CLI flags or API changes.

---

## Summary

Follow idiomatic Go practices with strong emphasis on KISS, YAGNI, observability, and disaster recovery. Maintain 100% unit test coverage, use dependency injection and context propagation, and keep documentation and infrastructure consistent with evolving APIs.
