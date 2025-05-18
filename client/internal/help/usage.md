# 🤖 LLM Client Usage

A simple CLI for interacting with OpenAI-compatible LLMs.

---

## 🧪 Examples

### 🔹 One-off Query (Non-Streaming)
Run a single prompt and exit.

```bash
go run client/main.go --query "Tell me a joke" --stream=false
```

### 🔹 Interactive Mode (Non-Streaming)
Enter prompts interactively.

```bash
go run client/main.go --stream=false
```

### 🔹 Interactive Mode with Auditing
Enter prompts interactively and audit each exchange.

```bash
go run client/main.go --stream=false --audit=true
```

### 🔹 One-off Query (Streaming with Auditing)
Run a prompt in one-off mode with live streaming output and audit enabled.

```bash
go run client/main.go --query "Tell me a joke" --stream=true --audit=true
```

---

✅ **Tips**

- Use `--stream=true` to enable real-time streaming responses.
- Use `--audit=false` to disable logging of prompts and responses.
- Customize output with `--output markdown`, `--output json`, or `--output yaml`.
- Set a different model with `--model gpt-3.5-turbo` or other supported models.