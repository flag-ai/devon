# DEVON — Feature Backlog

Planned features and ideas for future development.

## Documentation

- [ ] **Docs workflow trigger paths** — Review `.github/workflows/docs.yaml` trigger paths; if CLI-relevant code exists outside `src/devon/cli/**` (e.g., source plugins defining commands), add those paths
- [ ] **Download --include pattern docs** — Clarify in `docs/guides/downloading.md` whether `--include` glob patterns match filenames, relative paths, or both
- [ ] **Clone URL** — Verify the clone URL in `docs/getting-started/installation.md` matches the actual repository URL
- [ ] **Environment variable config** — Document which settings can also be controlled via environment variables (e.g., `DEVON_STORAGE_BASE_PATH`)

## Future Ideas

- [ ] **CI pipeline** — Add GitHub Actions CI workflow for tests, linting, and type checking (no CI exists yet)
- [ ] **Additional model sources** — Ollama registry, ModelScope, or local directory scanning as source plugins
- [ ] **Model integrity verification** — SHA256 checksum verification after download completion
- [ ] **Storage quota enforcement** — Enforce `max_size_gb` by blocking downloads that would exceed the limit
