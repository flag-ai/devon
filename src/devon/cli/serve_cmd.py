"""devon serve — start the REST API server."""

import click
from rich.console import Console

console = Console()


@click.command()
@click.option("--host", default="127.0.0.1", help="Bind address")
@click.option("--port", default=8000, type=int, help="Port number")
@click.option("--reload", "use_reload", is_flag=True, help="Enable auto-reload (dev only)")
def serve(host, port, use_reload):
    """Start the DEVON REST API server.

    Requires the 'api' extras: poetry install --extras api

    Examples:
      devon serve
      devon serve --host 0.0.0.0 --port 9000
    """
    try:
        import uvicorn  # noqa: F811
    except ImportError:
        console.print(
            "[red]Error: API dependencies not installed.[/red]\n"
            "Install with: [bold]poetry install --extras api[/bold]"
        )
        raise SystemExit(1)

    console.print(f"[cyan]Starting DEVON API on {host}:{port}[/cyan]")

    uvicorn.run(
        "devon.api.app:create_app",
        factory=True,
        host=host,
        port=port,
        reload=use_reload,
    )
