# llm-fast-wrapper

A blazing-fast streaming API for LLMs with Fiber and Gin. Includes PostgreSQL logging,
Prometheus metrics, Jaeger tracing, and Splunk integration. Run locally with Docker and
Taskfile.

## Usage

```bash
# start infrastructure
task docker:up

# run server (fiber)
task dev

# run server (gin)
task dev -- --gin

# test client
task client:run -- --prompt "hello"
```

See `Taskfile.yaml` and `README.md` for more details.
