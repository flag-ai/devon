from datetime import datetime

import click
from rich.console import Console
from rich.table import Table

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes

console = Console()


@click.command()
@click.option("--source", "-s", help="Filter by source")
def list_models(source):
    """List locally downloaded models.

    Example:
      devon list
      devon list --source huggingface
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)
    models = storage.list_local_models(source=source)

    if not models:
        console.print("[yellow]No models downloaded yet.[/yellow]")
        console.print("\n[dim]Use 'devon search' to find models[/dim]")
        return

    console.print("\n[green]Downloaded models:[/green]\n")

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Model", style="cyan", width=50)
    table.add_column("Source", width=12)
    table.add_column("Size", justify="right", width=10)
    table.add_column("Downloaded", width=20)

    for model in models:
        try:
            downloaded = datetime.fromisoformat(model["downloaded_at"]).strftime("%Y-%m-%d %H:%M")
        except (ValueError, TypeError):
            downloaded = model.get("downloaded_at", "unknown")[:16]
        size = format_bytes(model["size_bytes"])

        table.add_row(
            model["model_id"],
            model["source"],
            size,
            downloaded,
        )

    console.print(table)
    console.print(
        f"\n[dim]Total: {len(models)} models, {format_bytes(storage.get_total_size())}[/dim]"
    )
