import json

import click
from rich.console import Console

from devon.config.settings import Settings
from devon.storage.organizer import ModelStorage

console = Console()


@click.command()
@click.option("--format", "fmt", type=click.Choice(["kitt", "json"]), default="kitt")
@click.option("--output", "-o", type=click.Path())
def export(fmt, output):
    """Export model list for KITT.

    Examples:
      devon export --format kitt -o models.txt
      kitt run --model-list models.txt --engine vllm
    """
    settings = Settings()
    storage = ModelStorage(base_path=settings.storage_path)
    models = storage.list_local_models()

    if not models:
        console.print("[yellow]No models to export.[/yellow]")
        return

    if fmt == "kitt":
        lines = [model["path"] for model in models]
        content = "\n".join(lines)
    else:
        # Strip non-serializable fields for clean JSON
        export_data = []
        for model in models:
            export_data.append(
                {
                    "source": model["source"],
                    "model_id": model["model_id"],
                    "path": model["path"],
                    "size_bytes": model["size_bytes"],
                    "downloaded_at": model["downloaded_at"],
                    "files": model["files"],
                }
            )
        content = json.dumps(export_data, indent=2)

    if output:
        with open(output, "w") as f:
            f.write(content)
        console.print(f"[green]Exported {len(models)} models to {output}[/green]")
    else:
        console.print(content)
