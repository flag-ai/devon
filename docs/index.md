# DEVON Documentation

> *"DEVON manages the models. KITT tests them."*

**DEVON** (Discovery Engine and Vault for Open Neural models) is a CLI tool and REST
API for discovering, downloading, and managing LLM models from HuggingFace and other
sources. It gives you a single workflow to search massive model repositories, pull
weights to local storage, and keep your model vault organized -- so you can spend your
time testing, not hunting for files. Run it locally as a CLI or deploy it as a
containerized API for remote model management.

---

## Feature Highlights

**Smart Search**
:   Filter models by provider, size, parameter count, format, license, and more.
    Results are sorted by popularity and displayed in rich terminal tables.

**Easy Download**
:   Pass a HuggingFace URL or a model ID. Downloads resume automatically if
    interrupted -- no need to start over.

**Local Vault**
:   Models are stored in a clean directory hierarchy with a portable manifest. Track
    total disk usage at a glance with `devon status`.

**Directory Scanning**
:   Discover models added outside Devon with `devon scan`. The scanner infers metadata
    from file extensions, config files, and naming conventions — great for custom
    fine-tunes and manual copies.

**KITT Integration**
:   Export downloaded model paths in a format KITT can consume directly, bridging model
    management and inference testing.

**Source Plugin System**
:   HuggingFace is built in, but the architecture supports additional model sources
    through a plugin registry with a simple ABC interface.

**YAML Configuration**
:   Override storage paths, search defaults, and display options in a single config
    file. DEVON deep-merges your overrides with sensible defaults.

**REST API**
:   Run `devon serve` to expose every capability over HTTP. The FastAPI-based server
    supports optional bearer token auth and mirrors all CLI operations.

**Docker Ready**
:   Ship DEVON as a container with a single volume mount. Models, index, and config
    all live under `/data` for easy host-path binding.

---

## Quick Links

| Section | What you will find |
|---|---|
| [Getting Started](getting-started/index.md) | Installation, environment setup, and your first model download |
| [Guides](guides/index.md) | Task-focused walkthroughs -- searching, downloading, managing, REST API, Docker, KITT integration, configuration |
| [Reference](reference/index.md) | CLI reference, configuration schema, model metadata, and storage index format |
| [Concepts](concepts/index.md) | Architecture overview, source plugin system, and storage design |
| [Changelog](changelog.md) | Release history and migration notes |

---

## At a Glance

```text
devon search --provider qwen --params 30b --size "<100gb"
devon download Qwen/Qwen2.5-32B-Instruct
devon list
devon scan                                            # discover external models
devon export --format kitt -o models.txt

# Or run as an API server
devon serve --host 0.0.0.0 --port 8000
```

---

## Related Projects

DEVON is designed to work alongside **KITT**, the LLM inference engine testing
framework. Use DEVON to curate your model collection, then hand the paths to KITT
for benchmarking.

- [KITT Documentation](https://kirizan.github.io/kitt/)
- [KITT Repository](https://github.com/kirizan/kitt)
