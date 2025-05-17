#!/usr/bin/env python3
import argparse
import requests
import sseclient
from rich.console import Console
from datetime import datetime

console = Console()

def stream(url, prompt):
    with requests.get(url, params={"prompt": prompt}, stream=True) as r:
        client = sseclient.SSEClient(r)
        for ev in client.events():
            console.print(f"[{datetime.now():%H:%M:%S}] {ev.data}")

if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--server", default="http://localhost:8080/stream")
    p.add_argument("--prompt", required=True)
    args = p.parse_args()
    stream(args.server, args.prompt)
