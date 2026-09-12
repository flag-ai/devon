import shutil
from dataclasses import asdict

import click
from rich.console import Console

from devon.config.settings import Settings
from devon.sources.registry import SourceRegistry
from devon.storage.organizer import ModelStorage
from devon.utils.size_parser import format_bytes
from devon.utils.url_parser import URLParser

console = Console()


@click.command()
@click.argument("model_identifier")
@click.option("--source", "-s", help="Source (if not URL)")
@click.option("--force", "-f", is_flag=True, help="Re-download if exists")
@click.option("--include", "-i", multiple=True, help="File patterns to include (e.g. '*Q4_K_M*')")
@click.option("--yes", "-y", is_flag=True, help="Skip confirmation prompt")
def download(model_identifier, source, force, include, yes):
    """Download a model by ID or URL.

    Examples:
      devon download https://huggingface.co/Qwen/Qwen2.5-32B-Instruct
      devon download Qwen/Qwen2.5-32B-Instruct --source huggingface
      devon download bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --include '*Q4_K_M*'
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)

    # Parse input (URL or ID)
    if URLParser.is_url(model_identifier):
        parsed = URLParser.parse(model_identifier)
        if not parsed:
            console.print(f"[red]Unrecognized URL: {model_identifier}[/red]")
            return
        source_name, model_id = parsed
        console.print(f"[cyan]Detected: {source_name} model '{model_id}'[/cyan]")
    else:
        model_id = model_identifier
        source_name = source or "huggingface"
        console.print(f"[dim]Using source: {source_name}[/dim]")

    # Check if exists
    if not force and storage.is_downloaded(source_name, model_id):
        console.print(f"[yellow]Already downloaded: {model_id}[/yellow]")
        console.print(f"Path: {storage.get_model_path(source_name, model_id)}")
        console.print("\n[dim]Use --force to re-download[/dim]")
        return

    # Get source
    try:
        source_impl = SourceRegistry.get_source(source_name)()
    except ValueError as e:
        console.print(f"[red]Error: {e}[/red]")
        return

    # Get model info
    console.print("\n[cyan]Fetching model info...[/cyan]")
    try:
        model_info = source_impl.get_model_info(model_id)
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        return

    # Display info
    console.print(f"\n[bold]{model_info.model_name}[/bold]")
    console.print(f"Author: {model_info.author}")
    if model_info.parameter_count:
        console.print(f"Parameters: {model_info.parameter_count}B")
    console.print(f"Size: {format_bytes(model_info.total_size_bytes)}")
    console.print(f"Format: {', '.join(model_info.format)}")
    console.print(f"Files: {model_info.file_count}")

    # Check disk space
    free_space = shutil.disk_usage(storage.base_path).free
    required = model_info.total_size_bytes

    if required > 0 and free_space < required * 1.1:
        console.print("\n[red]Insufficient disk space[/red]")
        console.print(f"Required: {format_bytes(required)}")
        console.print(f"Available: {format_bytes(free_space)}")
        return

    # Show include patterns if specified
    if include:
        console.print(f"Include patterns: {', '.join(include)}")

    # Confirm
    console.print(f"\n[yellow]Download {format_bytes(required)}?[/yellow]")
    if not yes and not click.confirm("Proceed?", default=True):
        return

    # Download
    console.print("\n[green]Downloading...[/green]\n")
    dest = storage.get_model_path(source_name, model_id)

    try:
        allow_patterns = list(include) if include else None
        if allow_patterns:
            import re

            for pat in allow_patterns:
                if ".." in pat or not re.match(r"^[a-zA-Z0-9_.*?\-\[\]/]+$", pat):
                    console.print(f"[red]Error: Invalid include pattern: {pat!r}[/red]")
                    return
        files = source_impl.download_model(model_id, str(dest), allow_patterns=allow_patterns)

        # Register
        metadata_dict = asdict(model_info)
        storage.register_model(
            source=source_name,
            model_id=model_id,
            metadata=metadata_dict,
            files=files,
        )

        console.print("\n[green]Download complete![/green]")
        console.print(f"Path: {dest}")

    except Exception as e:
        console.print(f"\n[red]Download failed: {e}[/red]")
