# Downloading Models

DEVON downloads model files from HuggingFace and stores them in a local
cache organized by source and model ID.

## Download by URL

Pass a HuggingFace URL directly. DEVON auto-detects the source and
extracts the model ID:

```bash
devon download https://huggingface.co/Qwen/Qwen2.5-32B-Instruct
devon download https://hf.co/meta-llama/Llama-3.3-70B-Instruct
```

Both `huggingface.co` and `hf.co` URLs are recognized automatically.

## Download by Model ID

Provide the `author/model-name` identifier with the `--source` flag:

```bash
devon download Qwen/Qwen2.5-32B-Instruct --source huggingface
devon download meta-llama/Llama-3.3-70B-Instruct --source huggingface
```

When downloading by URL the source is detected from the hostname, so
`--source` is not required.

## Partial Downloads with Include Patterns

Large repositories often contain multiple quantization variants. Use
`--include` to download only the files you need. The flag accepts glob
patterns and can be specified multiple times:

```bash
devon download Qwen/Qwen2.5-32B-Instruct --include '*Q4_K_M*'
devon download Qwen/Qwen2.5-32B-Instruct --include '*Q4_K_M*' --include '*Q5_K_M*'
devon download Qwen/Qwen2.5-32B-Instruct --include '*.gguf'
```

## Force Re-download

If the model already exists locally, DEVON skips it. Pass `--force` to
download everything again:

```bash
devon download Qwen/Qwen2.5-32B-Instruct --force
```

## Skip Confirmation

By default DEVON shows the download size and asks for confirmation
before starting. Use `--yes` to skip the prompt:

```bash
devon download Qwen/Qwen2.5-32B-Instruct --yes
```

## Resuming Interrupted Downloads

Interrupted downloads resume automatically. If your connection drops or
you stop the process, simply run the same command again and DEVON picks
up where it left off.

## Disk Space Checking

Before starting a download, DEVON checks that your disk has enough free
space to hold the model files. If space is insufficient the download is
aborted with a clear error message showing the required and available
amounts.

## Storage Location

Downloaded models are stored under the configured base path, organized
by source and model ID:

```
~/.cache/devon/models/huggingface/Author/Model/
```

For example:

```
~/.cache/devon/models/huggingface/Qwen/Qwen2.5-32B-Instruct/
~/.cache/devon/models/huggingface/meta-llama/Llama-3.3-70B-Instruct/
```

You can change the base path in your [configuration file](configuration.md).
