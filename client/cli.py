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

# Send all output to stderr (to avoid conflict in rich output)
console = Console(stderr=True)

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
    logger.debug(f"Sending streaming request: model={model}, query={query}")
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

        logger.info("Streamed response successfully received.")
    except Exception as e:
        logger.exception("Error during streaming request.")
        console.print(f"[bold red]Error:[/bold red] {e}")
        sys.exit(1)

# ─────────────────────────────────────────────────────────────────────────────
# 📦 Non-streaming Response
# ─────────────────────────────────────────────────────────────────────────────
def single_response(client, model, query, temperature, render_markdown):
    logger.debug(f"Sending non-stream request: model={model}, query={query}")
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

        logger.info("Non-stream response successfully received.")
    except Exception as e:
        logger.exception("Error during non-streaming request.")
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
        logger.debug(f"User input: {query}")
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
@click.option("--base-url", default="http://localhost:8000/v1", help="LLM server URL")
def client(query, model, temperature, markdown, stream, base_url):
    logger.info("Client CLI launched.")
    logger.debug(f"Params - query={query}, model={model}, temperature={temperature}, stream={stream}, base_url={base_url}")

    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        logger.error("Missing environment variable: OPENAI_API_KEY")
        console.print("[bold red]Error:[/bold red] Missing required environment variable OPENAI_API_KEY")
        sys.exit(1)

    # Initialize OpenAI-compatible client
    openai_client = openai.OpenAI(base_url=base_url, api_key=api_key)
    logger.info(f"Connected to LLM API at endpoint: {base_url}")

    mode = "interactive" if not query else "query"
    logger.info(f"Mode selected: {mode}")
    {
        "interactive": lambda: interactive_mode(openai_client, model, temperature, markdown, stream),
        "query": lambda: dispatch_response(openai_client, model, query, temperature, markdown, stream)
    }[mode]()

if __name__ == "__main__":  # pragma: no cover
    client()  # pragma: no cover