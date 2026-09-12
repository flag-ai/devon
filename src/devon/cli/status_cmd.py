from collections import Counter

import click
from rich.console import Console

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes

console = Console()


@click.command()
def status():
    """Show storage status and statistics."""
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)
    models = storage.list_local_models()

    console.print("\n[bold cyan]DEVON Status[/bold cyan]\n")
    console.print(f"Models downloaded: {len(models)}")
    console.print(f"Total size: {format_bytes(storage.get_total_size())}")
    console.print(f"Storage path: {storage.base_path}")

    # Breakdown by source
    sources = Counter(m["source"] for m in models)

    if sources:
        console.print("\n[bold]By source:[/bold]")
        for source, count in sources.items():
            source_size = sum(m["size_bytes"] for m in models if m["source"] == source)
            console.print(f"  {source}: {count} models ({format_bytes(source_size)})")
