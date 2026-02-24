import importlib

import click
from rich.console import Console

console = Console()


class LazyGroup(click.Group):
    """Click group that lazily imports commands on first use."""

    _lazy_commands: dict[str, tuple[str, str]] = {}

    def list_commands(self, ctx: click.Context) -> list[str]:
        rv = set(super().list_commands(ctx))
        rv.update(self._lazy_commands.keys())
        return sorted(rv)

    def get_command(self, ctx: click.Context, cmd_name: str) -> click.Command | None:
        if cmd_name in self._lazy_commands:
            module_path, attr = self._lazy_commands[cmd_name]
            mod = importlib.import_module(module_path)
            return getattr(mod, attr)
        return super().get_command(ctx, cmd_name)


@click.group(cls=LazyGroup)
@click.version_option(version="1.0.0", prog_name="devon")
def cli():
    """DEVON - Discovery Engine and Vault for Open Neural models

    Discover, download, and manage LLM models with ease.

    Examples:
      devon search --provider qwen --params 30b
      devon download https://huggingface.co/Qwen/Qwen2.5-32B-Instruct
      devon list
    """
    pass


cli._lazy_commands = {
    "search": ("devon.cli.search_cmd", "search"),
    "download": ("devon.cli.download_cmd", "download"),
    "list": ("devon.cli.list_cmd", "list_models"),
    "info": ("devon.cli.info_cmd", "info"),
    "clean": ("devon.cli.clean_cmd", "clean"),
    "export": ("devon.cli.export_cmd", "export"),
    "status": ("devon.cli.status_cmd", "status"),
    "remove": ("devon.cli.remove_cmd", "remove"),
    "scan": ("devon.cli.scan_cmd", "scan"),
    "serve": ("devon.cli.serve_cmd", "serve"),
}

if __name__ == "__main__":
    cli()
