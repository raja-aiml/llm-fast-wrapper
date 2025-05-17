#!/usr/bin/env python3

import sys
import os
import openai
import click
from dotenv import load_dotenv
from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel

from lib.logging import setup_logger

# ─────────────────────────────────────────────────────────────────────────────
# 🎛️ Initialization
# ─────────────────────────────────────────────────────────────────────────────
load_dotenv()
logger = setup_logger("llm-client", "logs/client.log", console=False)
logger.info("Logger initialized.")
console = Console(stderr=True)

# ─────────────────────────────────────────────────────────────────────────────
# 🧩 Client Factory
# ─────────────────────────────────────────────────────────────────────────────
def create_openai_compatible_client(api_key: str, base_url: str):
    """
    Create a client for an OpenAI-compatible LLM server.
    """
    try:
        client = openai.OpenAI(api_key=api_key, base_url=base_url)
        logger.info(f"Connected to OpenAI-compatible endpoint: {base_url}")
        return client
    except Exception as e:
        logger.exception("Failed to connect to OpenAI-compatible LLM API.")
        console.print(f"[bold red]Connection Error:[/bold red] {e}")
        sys.exit(1)

# ─────────────────────────────────────────────────────────────────────────────
# 🧠 Unified Response Dispatcher
# ─────────────────────────────────────────────────────────────────────────────
def dispatch_response(client, model, query, temperature, render_markdown, stream):
    logger.debug(f"Dispatching query: model={model}, temperature={temperature}, stream={stream}")
    if stream:
        stream_response(client, model, query, temperature, render_markdown)
    else:
        single_response(client, model, query, temperature, render_markdown)

# ─────────────────────────────────────────────────────────────────────────────
# 📤 Streaming Response
# ─────────────────────────────────────────────────────────────────────────────
def stream_response(client, model, query, temperature, render_markdown):
    logger.debug(f"Streaming request to model={model}, query={query}")
    full_response = ""
    try:
        stream = client.chat.completions.create(
            model=model,
            messages=[{"role": "user", "content": query}],
            temperature=temperature,
            stream=True
        )
        console.print("[bold yellow]Assistant:[/bold yellow] ", end="")
        for chunk in stream:
            if chunk.choices and chunk.choices[0].delta.content:
                part = chunk.choices[0].delta.content
                full_response += part
                console.print(part, end="")
                sys.stdout.flush()
        print()
        if render_markdown:
            console.print("\n[bold blue]Rendered Markdown:[/bold blue]")
            console.print(Markdown(full_response))
        logger.info("Streamed response complete.")
    except Exception as e:
        logger.exception("Streaming request failed.")
        console.print(f"[bold red]Error:[/bold red] {e}")
        sys.exit(1)

# ─────────────────────────────────────────────────────────────────────────────
# 📦 Non-streaming Response
# ─────────────────────────────────────────────────────────────────────────────
def single_response(client, model, query, temperature, render_markdown):
    logger.debug(f"Non-streaming request to model={model}, query={query}")
    try:
        with console.status("[bold yellow]Thinking...[/bold yellow]", spinner="dots"):
            response = client.chat.completions.create(
                model=model,
                messages=[{"role": "user", "content": query}],
                temperature=temperature,
                stream=False
            )
        result = response.choices[0].message.content
        console.print(Panel(result, title="Assistant"))
        if render_markdown:
            console.print("\n[bold blue]Rendered Markdown:[/bold blue]")
            console.print(Markdown(result))
        logger.info("Non-stream response complete.")
    except Exception as e:
        logger.exception("Non-streaming request failed.")
        console.print(f"[bold red]Error:[/bold red] {e}")
        sys.exit(1)

# ─────────────────────────────────────────────────────────────────────────────
# 🧑‍💻 Interactive CLI Loop
# ─────────────────────────────────────────────────────────────────────────────
def interactive_mode(client, model, temperature, render_markdown, stream):
    logger.info("Entering interactive mode.")
    console.print("[bold green]Interactive LLM Chat[/bold green]")
    console.print("Type 'exit' or 'quit' to end.\n")
    while True:
        console.print("[bold cyan]You:[/bold cyan] ", end="")
        query = input().strip()
        if query.lower() in ("exit", "quit"):
            logger.info("User exited interactive mode.")
            break
        logger.debug(f"User query: {query}")
        dispatch_response(client, model, query, temperature, render_markdown, stream)

# ─────────────────────────────────────────────────────────────────────────────
# 🚪 Main CLI
# ─────────────────────────────────────────────────────────────────────────────
@click.command()
@click.option("--query", "-q", type=str, help="Prompt to send to the model")
@click.option("--model", "-m", default="expert", help="Model name (default: expert)")
@click.option("--temperature", "-t", default=0.7, help="Sampling temperature (default: 0.7)")
@click.option("--markdown", is_flag=True, help="Render output in Markdown")
@click.option("--stream/--no-stream", default=True, help="Stream output (default: stream)")
@click.option("--base-url", default="http://localhost:8000/v1", help="OpenAI-compatible LLM server URL")
def client(query, model, temperature, markdown, stream, base_url):
    logger.info("Client CLI started.")
    logger.debug(f"CLI Parameters - query={query}, model={model}, temp={temperature}, stream={stream}, base_url={base_url}")

    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        logger.error("Missing environment variable: OPENAI_API_KEY")
        console.print("[bold red]Error:[/bold red] Missing required environment variable OPENAI_API_KEY")
        sys.exit(1)

    openai_client = create_openai_compatible_client(api_key, base_url)
    mode = "interactive" if not query else "query"
    logger.info(f"Mode selected: {mode}")

    {
        "interactive": lambda: interactive_mode(openai_client, model, temperature, markdown, stream),
        "query": lambda: dispatch_response(openai_client, model, query, temperature, markdown, stream)
    }[mode]()

if __name__ == "__main__":  # pragma: no cover
    client()  # pragma: no cover