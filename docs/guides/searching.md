# Searching for Models

DEVON searches HuggingFace for models matching your query and filters.
Results are displayed in a Rich table showing the model ID, size,
download count, and format.

## Basic Search

Pass a keyword or phrase to find matching models:

```bash
devon search "llama 3"
devon search "code generation"
```

??? example "Example output (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 20 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                      ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ meta-llama/Llama-3.3-70B-Instruct          │    70B │  140 GB  │ safetensors   │      1.2M  │
    │ 2 │ meta-llama/Llama-3.1-8B-Instruct           │     8B │   16 GB  │ safetensors   │      985K  │
    │ 3 │ meta-llama/Llama-3.1-70B-Instruct          │    70B │  140 GB  │ safetensors   │      742K  │
    │ 4 │ bartowski/Meta-Llama-3.1-8B-Instruct-GGUF  │     8B │    5 GB  │ gguf          │      528K  │
    │ 5 │ meta-llama/Llama-3.2-3B-Instruct           │     3B │    6 GB  │ safetensors   │      415K  │
    └───┴────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

## Filter Reference

All filters are optional and can be combined freely. When multiple
filters are used, they stack with **AND logic** — only models matching
every filter are returned.

| Filter | Flag | Short | Accepts | Example |
|--------|------|-------|---------|---------|
| Provider | `--provider` | `-p` | Author/org name | `--provider qwen` |
| Parameters | `--params` | | Count with `b` suffix | `--params 30b` |
| Size | `--size` | | Operator + number + unit | `--size "<100gb"` |
| Format | `--format` | `-f` | Format name | `--format gguf` |
| Task | `--task` | `-t` | Pipeline tag | `--task text-generation` |
| License | `--license` | `-l` | SPDX identifier | `--license apache-2.0` |
| Limit | `--limit` | | Integer | `--limit 50` |
| Source | `--source` | | Source name | `--source huggingface` |

---

## Filters in Detail

### Provider (`--provider`, `-p`)

Filter by the model author or organization. The match is
**case-insensitive**.

```bash
devon search --provider qwen
devon search -p meta-llama
devon search -p microsoft
```

??? example "Example: `devon search --provider qwen --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                    ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ Qwen/Qwen2.5-72B-Instruct               │    72B │  144 GB  │ safetensors   │      1.8M  │
    │ 2 │ Qwen/Qwen2.5-32B-Instruct               │    32B │   64 GB  │ safetensors   │      1.1M  │
    │ 3 │ Qwen/Qwen2.5-7B-Instruct                │     7B │   14 GB  │ safetensors   │      892K  │
    │ 4 │ Qwen/Qwen2.5-Coder-32B-Instruct         │    32B │   64 GB  │ safetensors   │      654K  │
    │ 5 │ Qwen/QwQ-32B                             │    32B │   64 GB  │ safetensors   │      512K  │
    └───┴──────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

### Parameter Count (`--params`)

Filter by the number of parameters. Pass a number with a `b` suffix
(e.g., `7b`, `30b`, `70b`). DEVON matches with a **±20% tolerance** so
`--params 30b` returns models with 24B–36B parameters.

```bash
devon search --params 7b
devon search --params 30b
devon search --params 70b
```

??? example "Example: `devon search --params 7b --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                      ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ meta-llama/Llama-3.1-8B-Instruct           │     8B │   16 GB  │ safetensors   │      985K  │
    │ 2 │ Qwen/Qwen2.5-7B-Instruct                  │     7B │   14 GB  │ safetensors   │      892K  │
    │ 3 │ mistralai/Mistral-7B-Instruct-v0.3         │     7B │   14 GB  │ safetensors   │      612K  │
    │ 4 │ google/gemma-2-9b-it                       │     9B │   18 GB  │ safetensors   │      487K  │
    │ 5 │ bartowski/Qwen2.5-7B-Instruct-GGUF         │     7B │    5 GB  │ gguf          │      321K  │
    └───┴────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

!!! note "Tolerance"
    The ±20% window means `--params 30b` matches anything from 24B to 36B.
    This catches models labeled as 32B or 34B that are close to 30B.

### File Size (`--size`)

Constrain by total model size on disk. The value is an **operator**
followed by a **number** and **unit**.

**Supported operators:** `<`, `<=`, `>`, `>=`

**Supported units:** `mb`, `gb`, `tb` (case-insensitive)

```bash
devon search --size "<10gb"
devon search --size "<=50gb"
devon search --size ">100gb"
devon search --size ">=500mb"
```

!!! warning "Quote the value"
    Shell metacharacters `<` and `>` require quoting. Always wrap the size
    value in quotes: `--size "<100gb"`.

??? example "Example: `devon search --size \"<10gb\" --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                        ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ bartowski/Meta-Llama-3.1-8B-Instruct-GGUF    │     8B │    5 GB  │ gguf          │      528K  │
    │ 2 │ meta-llama/Llama-3.2-3B-Instruct             │     3B │    6 GB  │ safetensors   │      415K  │
    │ 3 │ Qwen/Qwen2.5-3B-Instruct                    │     3B │    6 GB  │ safetensors   │      298K  │
    │ 4 │ bartowski/Qwen2.5-7B-Instruct-GGUF           │     7B │    5 GB  │ gguf          │      321K  │
    │ 5 │ microsoft/Phi-3.5-mini-instruct              │     4B │    8 GB  │ safetensors   │      276K  │
    └───┴──────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

### Format (`--format`, `-f`)

Filter by model file format. Common values:

| Format | Description |
|--------|-------------|
| `gguf` | Quantized format for llama.cpp and compatible runtimes |
| `safetensors` | Safe, fast tensor serialization (HuggingFace default) |
| `pytorch` | PyTorch `.bin` checkpoint files |
| `onnx` | ONNX runtime format |

```bash
devon search --format gguf
devon search -f safetensors
devon search -f onnx
```

??? example "Example: `devon search --format gguf --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                          ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ bartowski/Meta-Llama-3.3-70B-Instruct-GGUF     │    70B │   42 GB  │ gguf          │      412K  │
    │ 2 │ bartowski/Meta-Llama-3.1-8B-Instruct-GGUF      │     8B │    5 GB  │ gguf          │      528K  │
    │ 3 │ bartowski/Qwen2.5-32B-Instruct-GGUF            │    32B │   20 GB  │ gguf          │      389K  │
    │ 4 │ bartowski/Qwen2.5-7B-Instruct-GGUF             │     7B │    5 GB  │ gguf          │      321K  │
    │ 5 │ bartowski/Mistral-7B-Instruct-v0.3-GGUF        │     7B │    5 GB  │ gguf          │      198K  │
    └───┴────────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

### Task (`--task`, `-t`)

Filter by the HuggingFace pipeline tag. This maps to the model's
intended use case.

Common task values:

| Task | Description |
|------|-------------|
| `text-generation` | Causal language models (GPT-style) |
| `text2text-generation` | Encoder-decoder models (T5-style) |
| `feature-extraction` | Embedding models |
| `text-classification` | Sentiment, topic, etc. |
| `question-answering` | Extractive QA |
| `summarization` | Text summarization |
| `translation` | Machine translation |

```bash
devon search --task text-generation
devon search -t feature-extraction
devon search -t text2text-generation
```

??? example "Example: `devon search --task text-generation --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                      ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ meta-llama/Llama-3.3-70B-Instruct          │    70B │  140 GB  │ safetensors   │      1.2M  │
    │ 2 │ Qwen/Qwen2.5-72B-Instruct                  │    72B │  144 GB  │ safetensors   │      1.8M  │
    │ 3 │ meta-llama/Llama-3.1-8B-Instruct           │     8B │   16 GB  │ safetensors   │      985K  │
    │ 4 │ mistralai/Mistral-7B-Instruct-v0.3         │     7B │   14 GB  │ safetensors   │      612K  │
    │ 5 │ google/gemma-2-27b-it                      │    27B │   54 GB  │ safetensors   │      487K  │
    └───┴────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

### License (`--license`, `-l`)

Filter by the model license. The match is **case-insensitive**. Use the
SPDX identifier when available.

Common license values:

| License | Description |
|---------|-------------|
| `apache-2.0` | Apache License 2.0 |
| `mit` | MIT License |
| `llama3.1` | Meta Llama 3.1 Community License |
| `gemma` | Google Gemma Terms of Use |
| `cc-by-4.0` | Creative Commons Attribution 4.0 |
| `cc-by-nc-4.0` | Creative Commons Attribution Non-Commercial 4.0 |

```bash
devon search --license apache-2.0
devon search -l mit
devon search -l llama3.1
```

??? example "Example: `devon search --license apache-2.0 --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                        ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ Qwen/Qwen2.5-72B-Instruct                    │    72B │  144 GB  │ safetensors   │      1.8M  │
    │ 2 │ Qwen/Qwen2.5-32B-Instruct                    │    32B │   64 GB  │ safetensors   │      1.1M  │
    │ 3 │ Qwen/Qwen2.5-7B-Instruct                    │     7B │   14 GB  │ safetensors   │      892K  │
    │ 4 │ mistralai/Mistral-7B-Instruct-v0.3           │     7B │   14 GB  │ safetensors   │      612K  │
    │ 5 │ microsoft/Phi-3.5-mini-instruct              │     4B │    8 GB  │ safetensors   │      276K  │
    └───┴──────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

### Limit (`--limit`)

Control the maximum number of results returned. Default is **20**.

```bash
devon search "llama" --limit 50
devon search "code" --limit 5
devon search --provider qwen --limit 3
```

You can set a permanent default in your
[configuration file](configuration.md) under `search.limit`.

### Source (`--source`)

Select which model source to query. Currently only `huggingface` is
supported (and is the default).

```bash
devon search "llama" --source huggingface
```

---

## Inline Query Parsing

DEVON's query parser can **extract filters directly from the search
query** without needing explicit flags. This enables a more natural
search style.

| Pattern | Detected As | Example |
|---------|-------------|---------|
| `<number>b` | Parameter count | `30b` → `--params 30b` |
| `gguf`, `safetensors`, `pytorch`, `onnx` | Format | `gguf` → `--format gguf` |
| `Q4_K_M`, `Q5_K_M`, `Q8_0`, `fp16`, `bf16`, `int8`, `int4` | Quantization | `Q4_K_M` → quant filter |

```bash
# These two are equivalent:
devon search "qwen 30b gguf"
devon search "qwen" --params 30b --format gguf

# These two are equivalent:
devon search "llama 7b safetensors"
devon search "llama" --params 7b --format safetensors
```

??? example "Example: `devon search \"qwen 30b gguf\" --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                        ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ bartowski/Qwen2.5-32B-Instruct-GGUF          │    32B │   20 GB  │ gguf          │      389K  │
    │ 2 │ Qwen/Qwen2.5-32B-Instruct-GGUF              │    32B │   20 GB  │ gguf          │      245K  │
    │ 3 │ bartowski/QwQ-32B-GGUF                       │    32B │   20 GB  │ gguf          │      178K  │
    │ 4 │ bartowski/Qwen2.5-Coder-32B-Instruct-GGUF   │    32B │   20 GB  │ gguf          │      156K  │
    │ 5 │ mradermacher/Qwen2.5-32B-Instruct-GGUF      │    32B │   20 GB  │ gguf          │       98K  │
    └───┴──────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

---

## Combining Filters

Filters stack with **AND logic**. Only models matching **every** filter
are returned. Combine as many as needed in a single command.

```bash
# Qwen models around 30B in GGUF format
devon search --provider qwen --params 30b --format gguf

# Llama instruct models under 50 GB in safetensors
devon search "instruct" --provider meta-llama --size "<50gb" --format safetensors

# Apache-licensed text-generation models (first 50 results)
devon search --task text-generation --license apache-2.0 --limit 50

# Small GGUF models for local testing
devon search --format gguf --size "<5gb" --params 7b

# Everything from Microsoft under 20 GB
devon search --provider microsoft --size "<20gb"
```

??? example "Example: `devon search --provider qwen --params 30b --format gguf --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 5 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                        ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ bartowski/Qwen2.5-32B-Instruct-GGUF          │    32B │   20 GB  │ gguf          │      389K  │
    │ 2 │ Qwen/Qwen2.5-32B-Instruct-GGUF              │    32B │   20 GB  │ gguf          │      245K  │
    │ 3 │ bartowski/Qwen2.5-Coder-32B-Instruct-GGUF   │    32B │   20 GB  │ gguf          │      156K  │
    │ 4 │ bartowski/QwQ-32B-GGUF                       │    32B │   20 GB  │ gguf          │      178K  │
    │ 5 │ mradermacher/Qwen2.5-32B-Instruct-GGUF      │    32B │   20 GB  │ gguf          │       98K  │
    └───┴──────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

??? example "Example: `devon search \"instruct\" --provider meta-llama --size \"<50gb\" --format safetensors --limit 5` (generated 2026-02-12)"

    ```
    Searching huggingface...

    Found 3 models:

    ┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
    ┃ # ┃ Model                                      ┃ Params ┃     Size ┃ Format        ┃  Downloads ┃
    ┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━╇━━━━━━━━━━━━┩
    │ 1 │ meta-llama/Llama-3.1-8B-Instruct           │     8B │   16 GB  │ safetensors   │      985K  │
    │ 2 │ meta-llama/Llama-3.2-3B-Instruct           │     3B │    6 GB  │ safetensors   │      415K  │
    │ 3 │ meta-llama/Llama-3.2-1B-Instruct           │     1B │    2 GB  │ safetensors   │      312K  │
    └───┴────────────────────────────────────────────┴────────┴──────────┴───────────────┴────────────┘
    ```

---

## Search Results Display

Results appear in a Rich table with these columns:

| Column | Description |
|--------|-------------|
| **#** | Row number |
| **Model** | Full model ID (`author/name`) |
| **Params** | Parameter count in billions |
| **Size** | Total size on disk (formatted) |
| **Format** | File format(s), up to two shown |
| **Downloads** | Total download count (abbreviated) |

Models are sorted by download count (most popular first) unless you
change `search.sort_by` in your [configuration file](configuration.md).

!!! tip "No results?"
    If a search returns no models, try relaxing your filters. Remove the
    most restrictive one first (often `--params` or `--size`) and
    broaden from there.
