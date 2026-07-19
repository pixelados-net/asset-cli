# Bucket Structure

`asset-cli` exists to take the messy, historically-grown Habbo asset dumps (Flash-era
`c_images`/`dcr` exports, `bundled` Nitro packages, ad-hoc "hubbly"-style repacks) and
normalize them into one predictable object storage layout. This page documents that
target layout: what each path holds, why it is shaped the way it is, and where
unavoidable numeric names still show up.

The canonical list of expected top-level paths lives in code at
[`internal/structure/paths.go`](../../internal/structure/paths.go)
(`structure.ExpectedPaths`). The `asset-cli structure check` and
`asset-cli structure create` commands read that same list, so this page and the CLI
can never drift apart silently — if you add a path here, add it there too.

## Rules

1. **Two levels deep, at most**, under the bucket root: `category/subcategory/file`.
   Nothing like `c_images/catalogue/album1584/file.gif` (four levels of jargon).
2. **No folder name is a bare number or hash.** A folder always says what it holds.
3. Where a numeric ID is genuinely unavoidable (protocol- or database-driven), it
   only ever appears as a **file name or third-level path** inside an already
   clearly-named folder — never as the category itself.
4. Database dumps (`.sql`) are never stored in this bucket. They belong in a
   database/migrations repository, not object storage.
5. Content that the current client no longer fetches still lives next to the content
   it is related to (e.g. `furniture/icons/` sits under `furniture/`, not off in its
   own unrelated top-level folder), so its purpose stays obvious — see
   [Still used vs. unused](#still-used-vs-unused) below before deleting anything.

## Layout

```
assets-prod/
├── avatar/
│   ├── clothing/                # .nitro clothing bundles (hair, hat, shirt, jacket…)
│   └── effects/                 # .nitro avatar effect bundles (EffectMap.json entries)
├── furniture/
│   ├── bundles/                 # .nitro furniture bundles, one file per classname
│   └── icons/                   # Flash-era per-furni icon PNGs. Confirmed unused by
│                                 #   nitro-renderer (see table below) — kept here only
│                                 #   for the old Flash client / external catalog tools.
├── pets/                        # .nitro pet bundles (bear, dog, cat, dragon…)
├── engine/                      # .nitro client-engine assets: room.nitro, tile_cursor,
│                                 #   place_holder, selection_arrow, floor_editor,
│                                 #   group_badge, avatar_additions — NOT catalog items
├── media/
│   ├── badges/                  # badge editor parts (was c_images/Badgeparts)
│   ├── catalog-pages/           # catalog page/product banner art (was c_images/
│                                 #   catalogue + album*/articles/catalogue_otherlangs).
│                                 #   The numeric catalog-page ID may still appear as a
│                                 #   file or sub-path here — that number is meaningful
│                                 #   (it is the actual page ID), it is just no longer
│                                 #   the top-level folder name.
│   ├── campaigns/                # promo/campaign art (web_promo, web_promo_small,
│                                 #   targetedoffers, hot_campaign_images_no, AdWarningsUK)
│   ├── guilds/                  # group/guild emblem images
│   ├── quests/                  # quest images
│   ├── talent/                  # talent track images
│   ├── stories/                 # Habbo Stories images
│   ├── flags/                   # reception/country flag icons
│   └── client-ui/               # client chrome: arrows, wallet, navigator, furniextras
├── sounds/
│   ├── ui/                      # named UI sounds (camera shutter, credits, messenger…)
│   └── machine-samples/         # numeric Sound Machine / Traxx samples. Still active:
│                                 #   nitro-renderer's SoundManager.playFurniSample and
│                                 #   MusicPlayer fetch these by numeric ID at runtime
│                                 #   (config key external.samples.url). The numeric
│                                 #   name is the protocol's ID, not an accident.
├── branding/                    # Nitro/hotel logos (svg/png)
└── gamedata/                    # FurnitureData.json, ProductData.json, FigureData.json,
                                  #   FigureMap.json, ExternalTexts.json, EffectMap.json,
                                  #   HabboAvatarActions.json, UITexts.json — flat, names
                                  #   already are clear, no further nesting needed
```

## Still used vs. unused

Do not lump every old Flash-era folder into one throwaway bucket — that discards
information the client actually needs. Before moving or deleting anything, check
whether the current client (`nitro-renderer`) still fetches it:

| Path                          | Still used by nitro-renderer?                          |
|--------------------------------|---------------------------------------------------------|
| `furniture/icons/` (was `dcr/hof_furni/icons/*`) | No — icons are built from the furni's own `.nitro` bundle by classname (`RoomContentLoader.getAssetUrlWithFurniIconBase`). Kept for the old Flash client / external tools only. |
| `sounds/machine-samples/` (was `dcr/hof_furni/mp3/sound_machine_sample_*.mp3`) | Yes — `SoundManager.playFurniSample` and `MusicPlayer` still fetch these by numeric ID (`external.samples.url`). |

When in doubt, grep the client before deciding a path is safe to archive or drop.
