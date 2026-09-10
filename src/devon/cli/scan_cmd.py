"""Scan model directories to discover and register untracked models."""

from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage
from devon.storage.scanner import ModelScanner
from devon.utils.size_parser import format_bytes

console = Console()


@click.command()
@click.argument("path", required=False, type=click.Path(exists=True))
@click.option("--reconcile", is_flag=True, help="Remove stale entries whose files no longer exist")
@click.option(
    "--dry-run", is_flag=True, help="Show what would change without modifying the manifest"
)
def scan(path, reconcile, dry_run):
    """Scan model directory to discover untracked models.

    Walks the model directory tree, identifies model files, and registers
    any models not already in the manifest. Useful for discovering models
    added outside Devon (custom models, manual copies).

    Examples:
      devon scan
      devon scan /data/models
      devon scan --reconcile
      devon scan --dry-run
    """
    settings = Settings()
    scan_path = Path(path) if path else settings.storage_path
    storage = ModelStorage(base_path=settings.storage_path)
    scanner = ModelScanner()

    console.print(f"\n[cyan]Scanning {scan_path}...[/cyan]\n")

    # Discover new models
    existing_keys = set(storage.index.keys())
    new_entries = scanner.scan(scan_path, existing_keys)

    # Find stale entries
    stale_keys = scanner.find_stale(storage.index)

    # Build results table
    table = Table(show_header=True, header_style="bold")
    table.add_column("Model", style="cyan")
    table.add_column("Source")
    table.add_column("Size", justify="right")
    table.add_column("Status")

    added = 0
    existing = len(existing_keys)
    stale = len(stale_keys)
    removed = 0

    for entry in new_entries:
        size_str = format_bytes(entry["size_bytes"])
        table.add_row(entry["model_id"], entry["source"], size_str, "[green]new[/green]")
        if not dry_run:
            storage.index[f"{entry['source']}::{entry['model_id']}"] = entry
        added += 1

    for key in stale_keys:
        entry = storage.index[key]
        size_str = format_bytes(entry["size_bytes"])
        if reconcile:
            status_label = "[red]removed[/red]" if not dry_run else "[yellow]would remove[/yellow]"
        else:
            status_label = "[yellow]stale[/yellow]"
        table.add_row(entry["model_id"], entry["source"], size_str, status_label)

    if reconcile and not dry_run:
        for key in stale_keys:
            del storage.index[key]
            removed += 1

    # Save manifest if changes were made
    if not dry_run and (added > 0 or removed > 0):
        storage._save_index()

    if table.row_count > 0:
        console.print(table)
    console.print()

    # Summary
    parts = [f"[green]{added} added[/green]", f"{existing} already tracked"]
    if stale > 0:
        if reconcile and not dry_run:
            parts.append(f"[red]{removed} removed[/red]")
        else:
            parts.append(f"[yellow]{stale} stale[/yellow]")
    if dry_run:
        parts.append("[dim](dry run)[/dim]")

    console.print(" | ".join(parts))
