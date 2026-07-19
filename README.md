# asset-cli v0.0.1

[![Version](https://img.shields.io/badge/version-v0.0.1-5865F2.svg)](https://github.com/pixelados-net/asset-cli/releases/tag/v0.0.1)
[![CI](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/ci.yml)
[![Package](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml/badge.svg)](https://github.com/pixelados-net/asset-cli/actions/workflows/package.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pixelados-net/asset-cli.svg)](https://pkg.go.dev/github.com/pixelados-net/asset-cli)

`asset-cli` normalizes Habbo asset storage. Raw asset dumps (Flash-era `c_images`/`dcr` exports, `bundled` Nitro packages, ad-hoc repacks) each tend to grow their own deeply-nested, inconsistently-named folder layout, full of ambiguous numeric names. `asset-cli` defines one canonical MinIO bucket layout — documented in [`docs/wiki/STRUCTURE.md`](docs/wiki/STRUCTURE.md) — and gives you commands to check a bucket against it and repair what is missing, instead of auditing folders by hand in the MinIO console.

Functionality is organized as independent **realms**: small, transport-agnostic domains under `internal/<realm>/`, each exposing its capabilities through a Go interface (its port) and its own Cobra commands. This keeps a realm's logic reusable from the CLI or any future transport without change. Current realms cover bucket structure, clothing, effects, furniture, pets, aggregate statistics, and emulator synchronization. Under the hood this is Cobra command wiring, a MinIO object storage adapter, Uber Fx dependency injection, and structured Zap logging, validated by a real-process E2E harness. The tool is stateless: every invocation parses configuration, runs one command, and exits.

## Run

```sh
cp .env.example .env
go run ./cmd version
go run ./cmd structure check
go run ./cmd clothing check
go run ./cmd effects check
go run ./cmd furniture check
go run ./cmd pets check
go run ./cmd stats orphan
```

`ASSET_CLI_MINIO_ENDPOINT`, `ASSET_CLI_MINIO_ACCESS_KEY`, `ASSET_CLI_MINIO_SECRET_KEY`, and `ASSET_CLI_MINIO_BUCKET` are mandatory for any command that touches storage — every command except `version`.

## Realms

- **`structure`** — verifies and repairs the bucket's expected folder layout (see [`docs/wiki/STRUCTURE.md`](docs/wiki/STRUCTURE.md) for the full canonical tree and the reasoning behind it).
  - `asset-cli structure check` prints every expected path as `ok` or `missing` and exits non-zero if anything is missing.
  - `asset-cli structure create` creates a placeholder object for every missing expected path so it renders as a folder in the MinIO console.
- **`clothing`** — compares `.nitro` files under `avatar/clothing/` with the library IDs declared by `gamedata/FigureMap.json`.
  - `asset-cli clothing check` warns about unreferenced bundles and exits non-zero when a declared library has no bundle.
- **`effects`** — compares `.nitro` files under `avatar/effects/` with the library names declared by `gamedata/EffectMap.json`.
  - `asset-cli effects check` warns about unreferenced bundles and exits non-zero when a declared effect library has no bundle.
- **`furniture`** — compares `.nitro` files under `furniture/bundles/` with classnames from `gamedata/FurnitureData.json`.
  - `asset-cli furniture check` normalizes `*N` color variants, warns about unreferenced bundles, and exits non-zero for missing bundles.
- **`pets`** — compares `.nitro` files under `pets/` with Nitro's protocol-ordered standard pet asset names.
  - `asset-cli pets check` permits custom bundles as warnings and exits non-zero when a standard pet asset has no bundle.
- **`stats`** — reports `.nitro` totals and bundle/catalog integrity summaries.
  - `asset-cli stats nitro` counts clothing, effects, furniture, and pet bundles.
  - `asset-cli stats orphan` checks all four categories concurrently and reports matched, orphaned, and missing totals.
- **`sync`** — reconciles `FurnitureData.json` into the selected emulator's furniture definition table.
  - `asset-cli sync furniture check` is read-only.
  - `asset-cli sync furniture apply` is a dry run unless `--yes` is passed.

Clothing and effects intentionally have no emulator `apply`: Arcturus stores commercial clothing mappings and player effect grants, while Pixels has no clothing definition table and stores effect grants only. Pet behavior tables are also not asset catalogs. These operational tables are not equivalent to `items_base`/`furniture_definitions` and must not be populated from client asset maps.

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
