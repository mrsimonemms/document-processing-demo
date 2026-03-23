# document-processing-demo

Canonical document processing demo showcasing resilient AI workflows with Temporal.

<!-- toc -->

* [Architecture](#architecture)
* [Repository structure](#repository-structure)
* [Running locally](#running-locally)
  * [Prerequisites](#prerequisites)
  * [Start the worker](#start-the-worker)
* [Contributing](#contributing)
  * [Open in a container](#open-in-a-container)
  * [Commit style](#commit-style)

<!-- Regenerate with "pre-commit run -a markdown-toc" -->

<!-- tocstop -->

## Architecture

This demo shows how Temporal makes AI-powered document workflows reliable,
resumable and observable.

* `/golang` runs the Temporal worker: workflows, activities and provider integrations
* `/ui` will be the SvelteKit frontend (not yet implemented)

There is no separate Go REST API. When the UI is implemented, it will talk
directly to Temporal from SvelteKit server-side code.

## Repository structure

```text
golang/            Go module: Temporal worker, workflows, activities, providers
  main.go          Worker entrypoint
  internal/
    activities/    Temporal activity implementations
    models/        Shared workflow and activity types
    providers/     AI provider abstractions with failover
    workflows/     Temporal workflow definitions
ui/                SvelteKit frontend (not yet implemented)
```

The repository root contains cross-repo files (CLAUDE.md, pre-commit config,
root go.mod). Do not add application logic to the root.

## Running locally

### Prerequisites

* Go 1.22 or later
* A running Temporal server (use [Temporal CLI](https://docs.temporal.io/cli):
  `temporal server start-dev`)

### Start the worker

```bash
cd golang
go run .
```

By default the worker connects to `localhost:7233` with the `default` namespace.
Set environment variables to connect to Temporal Cloud:

| Variable             | Purpose                         |
| -------------------- | ------------------------------- |
| `TEMPORAL_ADDRESS`   | Temporal server address         |
| `TEMPORAL_NAMESPACE` | Temporal namespace              |
| `TEMPORAL_TLS_CERT`  | Path to TLS certificate         |
| `TEMPORAL_TLS_KEY`   | Path to TLS key                 |
| `TEMPORAL_API_KEY`   | API key (Temporal Cloud)        |
| `LOG_LEVEL`          | Zerolog level (default: `info`) |

## Contributing

### Open in a container

* [Open in a container](https://code.visualstudio.com/docs/devcontainers/containers)

### Commit style

All commits must be done in the [Conventional Commit](https://www.conventionalcommits.org)
format.

```git
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```
