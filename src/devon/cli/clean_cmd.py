from datetime import datetime, timedelta

import click
from rich.console import Console

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes

console = Console()


@click.command()
@click.option("--unused", is_flag=True, help="Remove unused models")
@click.option("--days", default=30, help="Consider unused after N days")
@click.option("--all", "clean_all", is_flag=True, help="Remove all downloaded models")
@click.option("--dry-run", is_flag=True, help="Show what would be removed without removing")
def clean(unused, days, clean_all, dry_run):
    """Clean up downloaded models.

    Examples:
      devon clean --unused --days 30
      devon clean --all
      devon clean --unused --dry-run
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)
    models = storage.list_local_models()

    if not models:
        console.print("[yellow]No models to clean.[/yellow]")
        return

    to_remove = []

    if clean_all:
        to_remove = models
    elif unused:
        cutoff = datetime.now() - timedelta(days=days)
        for model in models:
            last_used = model.get("last_used")
            if last_used is None:
                # Never used - check download date
                try:
                    downloaded = datetime.fromisoformat(model["downloaded_at"])
                except (ValueError, TypeError):
                    continue
                if downloaded < cutoff:
                    to_remove.append(model)
            else:
                try:
                    used_at = datetime.fromisoformat(last_used)
                except (ValueError, TypeError):
                    continue
                if used_at < cutoff:
                    to_remove.append(model)
    else:
        console.print("[yellow]Specify --unused or --all[/yellow]")
        return

    if not to_remove:
        console.print("[green]Nothing to clean.[/green]")
        return

    # Show what will be removed
    total_size = sum(m["size_bytes"] for m in to_remove)
    console.print(f"\n[yellow]Models to remove ({len(to_remove)}):[/yellow]")
    for model in to_remove:
        console.print(f"  {model['model_id']} ({format_bytes(model['size_bytes'])})")
    console.print(f"\n[bold]Total: {format_bytes(total_size)}[/bold]")

    if dry_run:
        console.print("\n[dim](dry run - nothing removed)[/dim]")
        return

    if not click.confirm("\nProceed with removal?", default=False):
        return

    # Remove
    removed = 0
    for model in to_remove:
        if storage.delete_model(model["source"], model["model_id"]):
            removed += 1
            console.print(f"  [red]Removed: {model['model_id']}[/red]")

    console.print(f"\n[green]Removed {removed} models, freed {format_bytes(total_size)}[/green]")
