import click
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from devon.config.settings import Settings
from devon.sources.registry import SourceRegistry
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes, format_number

console = Console()


@click.command()
@click.argument("model_id")
@click.option("--source", "-s", default="huggingface")
def info(model_id, source):
    """Show detailed info about a model.

    Example:
      devon info Qwen/Qwen2.5-32B-Instruct
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)

    # Check if we have it locally first
    local = storage.get_model_entry(source, model_id)

    if local:
        console.print("\n[green]Local model:[/green]")
        console.print(f"  Path: {local['path']}")
        console.print(f"  Size: {format_bytes(local['size_bytes'])}")
        console.print(f"  Downloaded: {local['downloaded_at']}")
        if local.get("last_used"):
            console.print(f"  Last used: {local['last_used']}")
        console.print()

    # Fetch remote info
    console.print(f"[cyan]Fetching info from {source}...[/cyan]")
    try:
        source_impl = SourceRegistry.get_source(source)()
        model = source_impl.get_model_info(model_id)
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        return

    # Display detailed info
    info_table = Table(show_header=False, box=None, padding=(0, 2))
    info_table.add_column("Field", style="bold")
    info_table.add_column("Value")

    info_table.add_row("Model", model.model_id)
    info_table.add_row("Author", model.author)
    if model.parameter_count:
        info_table.add_row("Parameters", f"{model.parameter_count}B")
    if model.architecture:
        info_table.add_row("Architecture", model.architecture)
    info_table.add_row("Size", format_bytes(model.total_size_bytes))
    info_table.add_row("Files", str(model.file_count))
    if model.format:
        info_table.add_row("Format", ", ".join(model.format))
    if model.quantization:
        info_table.add_row("Quantization", model.quantization)
    if model.license:
        info_table.add_row("License", model.license)
    info_table.add_row("Downloads", format_number(model.downloads))
    info_table.add_row("Likes", format_number(model.likes))
    if model.created_at:
        info_table.add_row("Created", model.created_at[:10])
    if model.updated_at:
        info_table.add_row("Updated", model.updated_at[:10])
    info_table.add_row("URL", model.web_url)

    console.print(Panel(info_table, title=f"[bold]{model.model_name}[/bold]"))
