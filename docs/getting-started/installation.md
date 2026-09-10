# Installation

This guide covers cloning DEVON, installing dependencies, activating the virtual
environment, and verifying the CLI.

---

## Prerequisites

| Requirement | Minimum Version | Check |
|---|---|---|
| Python | 3.10+ | `python3 --version` |
| Poetry | 1.7+ | `poetry --version` |
| git | any recent version | `git --version` |

!!! tip
    If you don't have Poetry yet, install it with the official installer:
    `curl -sSL https://install.python-poetry.org | python3 -`

---

## Clone and Install

```bash
git clone https://github.com/kirizan/devon.git
cd devon
poetry install
```

`poetry install` creates a virtual environment, resolves all dependencies, and
installs the `devon` CLI entry point inside that environment.

---

## Activate the Environment

After installation you need to activate the Poetry-managed virtual environment so
your shell can find the `devon` command.

### Option 1 -- Activate the virtual environment

=== "Bash / Zsh"

    ```bash
    eval $(poetry env activate)
    ```

=== "Fish"

    ```fish
    eval (poetry env activate)
    ```

Once activated, the `devon` command is available directly:

```bash
devon --version
devon search "llama"
```

To deactivate when you are done:

```bash
deactivate
```

### Option 2 -- Use `poetry run`

Prefix every command with `poetry run` -- works in any shell, no activation needed:

```bash
poetry run devon --version
poetry run devon search "llama"
```

---

## Verify the Installation

```bash
devon --version
```

You should see output like:

```text
devon, version 0.1.0
```

If the command is not found, make sure you have activated the virtual environment
(Option 1) or are using `poetry run` (Option 2).

---

## API Extras (Optional)

To use the REST API server (`devon serve`), install the optional API
dependencies:

```bash
poetry install --extras api
```

This adds FastAPI and Uvicorn. The CLI works without these extras.

## Docker (Optional)

DEVON can also run as a containerized REST API. See the
[Docker Deployment](../guides/docker.md) guide for details.

---

## Next Steps

Head to the [First Model](first-model.md) tutorial to search for a model and
download it to your local vault.
