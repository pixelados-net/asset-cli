# asset-cli v0.0.1

[![Version](https://img.shields.io/badge/version-v0.0.1-5865F2.svg)](https://github.com/pixelados-net/asset-cli/releases/tag/v0.0.1)
[![CI](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml)
[![Package](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pixelados-net/asset-cli.svg)](https://pkg.go.dev/github.com/pixelados-net/asset-cli)

`asset-cli` normalizes Habbo asset storage. Raw asset dumps (Flash-era `c_images`/`dcr` exports, `bundled` Nitro packages, ad-hoc repacks) each tend to grow their own deeply-nested, inconsistently-named folder layout, full of ambiguous numeric names. `asset-cli` defines one canonical MinIO bucket layout — documented in [`docs/wiki/STRUCTURE.md`](docs/wiki/STRUCTURE.md) — and gives you commands to check a bucket against it and repair what is missing, instead of auditing folders by hand in the MinIO console.

Functionality is organized as independent **realms**: small, transport-agnostic domains under `internal/<realm>/`, each exposing its capabilities through a Go interface (its port) and its own Cobra commands. This keeps a realm's logic reusable from the CLI or any future transport without change. The first realm is `structure`; further realms (`furniture`, `catalog`, …) follow the same pattern as the tool grows. Under the hood this is Cobra command wiring, a MinIO object storage adapter, Uber Fx dependency injection, and structured Zap logging, validated by a real-process E2E harness. The tool is stateless: every invocation parses configuration, runs one command, and exits.

## Run

```sh
cp .env.example .env
go run ./cmd version
go run ./cmd structure check
go run ./cmd structure create
```

`ASSET_CLI_MINIO_ENDPOINT`, `ASSET_CLI_MINIO_ACCESS_KEY`, `ASSET_CLI_MINIO_SECRET_KEY`, and `ASSET_CLI_MINIO_BUCKET` are mandatory for any command that touches storage — every command except `version`.

## Realms

- **`structure`** — verifies and repairs the bucket's expected folder layout (see [`docs/wiki/STRUCTURE.md`](docs/wiki/STRUCTURE.md) for the full canonical tree and the reasoning behind it).
  - `asset-cli structure check` prints every expected path as `ok` or `missing` and exits non-zero if anything is missing.
  - `asset-cli structure create` creates a placeholder object for every missing expected path so it renders as a folder in the MinIO console.

Logging is configured independently from the environment:

```dotenv
ASSET_CLI_LOG_LEVEL=info
ASSET_CLI_LOG_FORMAT=console
```

`ASSET_CLI_LOG_LEVEL` accepts Zap levels such as `debug`, `info`, `warn`, and `error`. `ASSET_CLI_LOG_FORMAT` accepts `console` for local readability or `json` for structured ingestion.

Print the CLI version with:

```sh
go run ./cmd version
```

## Release

GHCR publication runs only for tags matching `v*.*.*`. The release workflow validates tests, Vet, Staticcheck and compilation, builds all supported binary targets, and then publishes the multi-architecture image. A tag such as `v0.1.0` produces `v0.1.0`, `0.1.0`, `0.1`, `0`, and `latest` image tags and embeds `0.1.0` in the binary.

The image entrypoint is the `asset-cli` binary itself, so any command runs directly against a container with the required `ASSET_CLI_*` variables:

```sh
docker run --rm --env-file .env ghcr.io/pixelados-net/asset-cli:latest version
```

Configure the repository Actions secret `DISCORD_WEBHOOK_URL` to enable the tag notification workflow, then publish a release with:

```sh
git tag v0.1.0
git push origin v0.1.0
```

## Repository layout

- `cmd/` contains the process entrypoint.
- `internal/<realm>/` contains each realm's port (`Service` interface), its Fx-provided implementation, and its own Cobra command tree (e.g. `internal/structure/`).
- `platform/cli/` contains the Cobra root command and is the one place that assembles every realm's command tree.
- `platform/minio/` contains the reusable MinIO object storage client.
- `platform/logger/` builds the injected Zap logger.
- `platform/config/` unifies every platform module's own `Config` struct and parses `ASSET_CLI_*` variables once.
- `platform/bootstrap/` composes the platform-owned Fx modules (`config`, `logger`, `minio`) and exposes `Invoke`, which realm commands use to resolve their own dependencies without importing MinIO or Zap directly.
- Every DI-enabled package owns its `module.go`; bootstrap only composes platform modules, never realms, to avoid an import cycle back into `internal/`.
- `e2e/` builds the real binary and validates command output through a real process.
- `docs/wiki/` is synced to the GitHub wiki on every push to `main` that touches it — see [`docs/wiki/STRUCTURE.md`](docs/wiki/STRUCTURE.md) for the canonical bucket layout.

## Validate

```sh
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
staticcheck ./...
go test ./... -race
go build -trimpath -o /tmp/asset-cli ./cmd
docker build -t ghcr.io/pixelados-net/asset-cli:local .
```

This repository intentionally does not include Docker Compose files.
