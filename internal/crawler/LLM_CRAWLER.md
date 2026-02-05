# LLM-Based Crawler

This document describes the LLM-based crawler implementation for extracting structured data from the Kulturelle Landpartie website.

## Overview

The LLM-based crawler uses OpenRouter API with structured JSON outputs to parse HTML content. This approach is more robust than regex-based parsing and can handle variations in website structure.

## Features

- **Structured JSON Output**: Uses JSON Schema validation for reliable parsing
- **OpenRouter Integration**: Access to 300+ models through single API
- **Response Caching**: Disk-based caching to minimize API costs
- **Batch Processing**: Process multiple items per API call for efficiency
- **Fallback Support**: Automatically falls back to regex parsing if LLM fails
- **Dry Run Mode**: Test extraction without making API calls

## Usage

### Basic Usage

```bash
export OPENROUTER_API_KEY="sk-or-v1-..."
go run cmd/crawler/main.go --use-llm
```

### With Different Model

```bash
go run cmd/crawler/main.go --use-llm --openrouter-model=google/gemini-flash-1.5
```

### Dry Run (No API Calls)

```bash
go run cmd/crawler/main.go --use-llm --dry-run
```

### Custom Cache Settings

```bash
go run cmd/crawler/main.go --use-llm --llm-cache-dir=./cache --llm-cache-ttl=48h
```

### Batch Size Control

```bash
go run cmd/crawler/main.go --use-llm --llm-batch-size=10
```

## CLI Flags

- `--use-llm`: Enable LLM parsing (requires OPENROUTER_API_KEY env var)
- `--openrouter-model`: Model to use (default: `openai/gpt-4o-mini`)
- `--llm-cache-dir`: Cache directory (default: `./.llm_cache`)
- `--llm-cache-ttl`: Cache TTL (default: `24h`)
- `--llm-batch-size`: Items per API call (default: `5`)
- `--dry-run`: Test mode without API calls

## Recommended Models

| Model | Cost (per 1M tokens) | Speed | Quality |
|-------|---------------------|-------|---------|
| `openai/gpt-4o-mini` | $0.15 / $0.60 | Fast | Excellent |
| `google/gemini-flash-1.5` | $0.075 / $0.30 | Very Fast | Good |
| `mistralai/mistral-small-3.1-24b-instruct` | $0.10 / $0.30 | Fast | Good |

## Cost Estimation

With ~87 venues, ~300 events, ~150 exhibitions:
- Full crawl (no cache): ~$0.40
- With caching (re-runs): ~$0.05

## Architecture

```
HTML Content
    ↓
Extract HTML Blocks
    ↓
Check Cache (SHA256 hash)
    ↓
OpenRouter API Call
    ↓
JSON Schema Validation
    ↓
Unmarshal to Go Structs
    ↓
Cache Response
    ↓
Return Data
```

## Error Handling

1. **Cache Hit**: Return cached response immediately
2. **API Error**: Log warning and fall back to regex
3. **Parse Error**: Log warning and fall back to regex
4. **Validation Error**: Log warning and fall back to regex

## File Structure

```
internal/crawler/
├── llm/
│   ├── types.go        # Request/response types
│   ├── client.go       # OpenRouter client
│   ├── cache.go        # Response caching
│   └── batch.go        # Batch processing
├── schemas/
│   ├── venue.go        # Venue JSON schema
│   ├── event.go        # Event JSON schema
│   └── exhibition.go   # Exhibition JSON schema
└── llm_parser.go       # Main parser implementation
```

## Testing

Run existing tests:
```bash
go test ./internal/crawler/...
```

Test with dry run:
```bash
go run cmd/crawler/main.go --use-llm --dry-run --verbose
```

## Cache

The cache is stored in `./.llm_cache/` by default. Each response is keyed by a SHA256 hash of the HTML content.

To clear cache:
```bash
rm -rf ./.llm_cache
```

## Fallback Behavior

If LLM parsing fails for any venue, event, or exhibition, the crawler automatically falls back to the existing regex-based parser. This ensures the crawl can always complete even if the LLM API is unavailable.

## Benefits Over Regex

1. **Robustness**: Handles variations in HTML structure
2. **Intelligence**: Understands context and relationships
3. **Flexibility**: Can extract complex nested data
4. **Maintainability**: No need to update regex patterns when site changes
5. **Multi-date Events**: Correctly extracts all dates for events with multiple occurrences
