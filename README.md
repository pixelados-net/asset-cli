# asset-cli v0.0.1

[![Version](https://img.shields.io/badge/version-v0.0.1-5865F2.svg)](https://github.com/pixelados-net/asset-cli/releases/tag/v0.0.1)
[![CI](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml)
[![Package](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pixelados-net/asset-cli.svg)](https://pkg.go.dev/github.com/pixelados-net/asset-cli)

`asset-cli` is a production-oriented Go boilerplate for a Habbo asset management CLI. It includes Cobra command wiring, a MinIO object storage adapter, Uber Fx dependency injection, structured Zap logging, and a real-process E2E harness. The tool is stateless: every invocation parses configuration, runs one command, and exits.

## Run

```sh
cp .env.example .env
go run ./cmd version
```

`ASSET_CLI_MINIO_ENDPOINT`, `ASSET_CLI_MINIO_ACCESS_KEY`, `ASSET_CLI_MINIO_SECRET_KEY`, and `ASSET_CLI_MINIO_BUCKET` are mandatory for any command that touches storage.

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

## Structure

- `cmd/` contains the process entrypoint.
- `internal/` contains domain-owned CLI command logic, added as the tool grows.
- `platform/cli/` contains the Cobra root command and its subcommands.
- `platform/minio/` contains the reusable MinIO object storage client.
- `platform/logger/` builds the injected Zap logger.
- `platform/config/` unifies every platform module's own `Config` struct and parses `ASSET_CLI_*` variables once.
- `platform/bootstrap/` composes focused Uber Fx modules for commands that need injected dependencies.
- Every DI-enabled package owns its `module.go`; bootstrap only composes those modules.
- `e2e/` builds the real binary and validates command output through a real process.
- `docs/wiki/` is synced to the GitHub wiki on every push to `main` that touches it.

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
