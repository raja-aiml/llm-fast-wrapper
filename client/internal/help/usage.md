# 🤖 LLM Client Usage

A simple CLI for interacting with OpenAI-compatible LLMs.

---

## 🧪 Examples

### 🔹 One-off query (non-streaming)
Run a single prompt and exit.

```bash
go run client/main.go --query "Tell me a joke" --stream=false
```

### 🔹 Interactive mode (non-streaming)
Enter prompts interactively.

```bash
go run client/main.go --stream=false
```

### 🔹 Non-interactive (non-streaming)
Run a prompt in one-off mode using CLI flags.

```bash
go run client/main.go --query "Tell me a joke" --stream=false
```

---

✅ **Tip:** Use `--stream=true` to enable streaming responses.