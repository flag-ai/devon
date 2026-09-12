import click
from rich.console import Console
from rich.table import Table

from devon.sources.registry import SourceRegistry
from devon.utils.size_parser import format_bytes, format_number

console = Console()


@click.command()
@click.option("--provider", "-p", help="Filter by provider/author")
@click.option("--params", help="Parameter count (e.g., 7b, 30b)")
@click.option("--size", help="Size constraint (e.g., <100gb)")
@click.option("--format", "-f", "fmt", help="Model format (gguf, safetensors)")
@click.option("--task", "-t", help="Task type")
@click.option("--license", "-l", "lic", help="License type")
@click.option("--limit", default=20, help="Max results")
@click.option("--source", default="huggingface", help="Source to search")
@click.argument("query", required=False)
def search(provider, params, size, fmt, task, lic, limit, source, query):
    """Search for models matching criteria.

    Examples:
      devon search --provider qwen --params 30b --size "<100gb"
      devon search "llama 3" --format gguf
    """
    # Build filters
    filters = {}
    if provider:
        filters["author"] = provider
    if params:
        filters["params"] = params
    if size:
        filters["size"] = size
    if fmt:
        filters["format"] = fmt
    if task:
        filters["task"] = task
    if lic:
        filters["license"] = lic

    # Get source
    try:
        source_impl = SourceRegistry.get_source(source)()
    except ValueError as e:
        console.print(f"[red]Error: {e}[/red]")
        return

    # Search
    console.print(f"[cyan]Searching {source}...[/cyan]")

    with console.status("[bold green]Searching..."):
        results = source_impl.search(query=query, filters=filters, limit=limit)

    if not results:
        console.print("[yellow]No models found.[/yellow]")
        return

    # Display results
    console.print(f"\n[green]Found {len(results)} models:[/green]\n")

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("#", style="dim", width=3)
    table.add_column("Model", style="cyan", width=40)
    table.add_column("Params", justify="right", width=8)
    table.add_column("Size", justify="right", width=10)
    table.add_column("Format", width=15)
    table.add_column("Downloads", justify="right", width=10)

    for idx, model in enumerate(results, 1):
        params_str = f"{model.parameter_count}B" if model.parameter_count else "?"
        size_str = format_bytes(model.total_size_bytes)
        format_str = ", ".join(model.format[:2]) if model.format else "?"
        downloads_str = format_number(model.downloads)

        table.add_row(str(idx), model.model_id, params_str, size_str, format_str, downloads_str)

    console.print(table)
