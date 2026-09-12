# Concepts

This section covers the key ideas and design decisions behind DEVON. Read
these pages to understand **why** DEVON works the way it does, not just
**how** to use it.

If you are looking for exact schemas or command flags, see the
[Reference](../reference/index.md) section instead.

## Topics

- **[Architecture](architecture.md)** -- Project structure, package layout,
  and the data flow from CLI commands through source plugins to local
  storage.

- **[Source Plugins](source-plugins.md)** -- The plugin system that lets
  DEVON talk to different model providers. Covers the abstract base class,
  the registry, and the built-in HuggingFace implementation.

- **[Storage Design](storage-design.md)** -- How DEVON organizes downloaded
  models on disk, why it uses a flat JSON index instead of a database, and
  how it handles atomic writes and size tracking.
