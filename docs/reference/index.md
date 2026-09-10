# Reference

Technical reference material for DEVON. These pages document the exact
interfaces, schemas, and data structures that make up the tool.

Use this section when you need precise details about a command flag,
a configuration key, or the shape of stored data.

## Sections

- **[CLI Reference](cli/index.md)** -- Every command, option, and argument
  exposed by the `devon` CLI.

- **[REST API Reference](rest-api.md)** -- Endpoint paths, request/response
  schemas, authentication, and error codes for the DEVON REST API.

- **[Configuration Schema](configuration.md)** -- All keys accepted in
  `~/.config/devon/config.yaml`, their types, and default values.

- **[Model Metadata](model-metadata.md)** -- Fields in the `ModelMetadata`
  dataclass returned by source plugins and stored in the local index.

- **[Storage Index](storage-index.md)** -- Format of the JSON index file,
  key conventions, and `ModelStorage` class methods.
