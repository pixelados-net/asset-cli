# asset-cli

`asset-cli` normalizes Habbo asset storage. Historically, asset dumps for a hotel
(Flash-era `c_images`/`dcr` exports, `bundled` Nitro packages, ad-hoc repacks from
tools like "hubbly") each grow their own inconsistent, deeply-nested folder layout,
full of ambiguous numeric names (`album1584`, `c123891`, `00011_icon.png`). This CLI
defines one canonical, human-readable bucket layout and gives you the commands to
verify and repair a bucket against it, instead of hand-auditing folders in the MinIO
console.

The tool is organized as independent **realms** — small, transport-agnostic domains
under `internal/<realm>/`, each exposing its own commands. The first realm is
[`structure`](STRUCTURE.md), which checks and repairs the bucket's expected folder
layout. Further realms (`furniture`, `catalog`, …) follow the same pattern as the
tool grows.

## Pages

- [Structure](STRUCTURE.md) — the canonical bucket layout, why each path is shaped
  the way it is, and which legacy paths are actually still in active use.

See the sidebar for the full page list.
