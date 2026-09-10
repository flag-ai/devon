from rich.progress import Progress, SpinnerColumn, BarColumn, TextColumn, DownloadColumn


def create_download_progress() -> Progress:
    """Create a Rich progress bar for downloads."""
    return Progress(
        SpinnerColumn(),
        TextColumn("[bold blue]{task.description}"),
        BarColumn(),
        DownloadColumn(),
        TextColumn("[dim]{task.fields[status]}"),
    )
