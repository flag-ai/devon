import click
from rich.console import Console

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes

console = Console()


@click.command()
@click.argument("model_id")
@click.option("--source", "-s", default="huggingface", help="Model source")
@click.option("--yes", "-y", is_flag=True, help="Skip confirmation prompt")
def remove(model_id, source, yes):
    """Remove a downloaded model.

    Examples:
      devon remove Qwen/Qwen2.5-32B-Instruct
      devon remove bartowski/Meta-Llama-3.1-8B-Instruct-GGUF -y
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)

    # Look up model in index
    entry = storage.get_model_entry(source, model_id)
    if not entry:
        console.print(f"[red]Model not found: {source}::{model_id}[/red]")
        console.print("\n[dim]Use 'devon list' to see downloaded models[/dim]")
        raise SystemExit(1)

    # Show info
    size = entry.get("size_bytes", 0)
    path = entry.get("path", "unknown")
    console.print(f"\n[bold]{model_id}[/bold]")
    console.print(f"Source: {source}")
    console.print(f"Size: {format_bytes(size)}")
    console.print(f"Path: {path}")

    # Confirm
    if not yes:
        console.print(f"\n[yellow]Delete {format_bytes(size)} from disk?[/yellow]")
        if not click.confirm("Proceed?", default=False):
            return

    # Delete
    result = storage.delete_model(source, model_id)
    if result:
        console.print(f"\n[green]Removed {model_id} ({format_bytes(size)} freed)[/green]")
    else:
        console.print(f"\n[red]Failed to remove {model_id}[/red]")
        raise SystemExit(1)
