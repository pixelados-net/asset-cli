// Package sync reconciles the client's furniture catalog with an emulator database.
//
// Direction is one-way: client -> emulator, for both existence and naming. Apply
// inserts every definition the client declares that the emulator is missing, and
// updates public_name/description on existing rows to match the client whenever
// they differ. It never deletes an emulator row. The client's gamedata is always
// the source of truth; the emulator is always the destination. See
// plan/EMULATORS.md for the full reasoning, including why this replaced an
// earlier emulator -> client naming design once real data showed the emulator's
// public_name is frequently an uncurated "<classname>_name" placeholder, while
// the client's copy is the genuinely curated one.
package sync

import "context"

// KindFloor and KindWall name a Definition's placement category — the same two
// kinds the furniture realm's gamedata reader tags by array membership.
const (
	KindFloor = "floor"
	KindWall  = "wall"
)

// Service defines the sync realm's capabilities, independent of transport.
type Service interface {
	// Check diffs the client's furniture catalog against the active emulator and
	// finds public_name/description mismatches for classnames present on both sides.
	Check(ctx context.Context) (Report, error)
	// Apply inserts every definition the client declares that the emulator is
	// missing, and updates public_name/description on existing emulator rows to
	// match the client wherever they differ. Never deletes an emulator row.
	Apply(ctx context.Context) (ApplyResult, error)
}

// ClientCatalog is what sync needs from the client's gamedata.
type ClientCatalog interface {
	// ListDefinitions returns every furniture definition the client declares.
	ListDefinitions(ctx context.Context) ([]Definition, error)
}

// EmulatorCatalog is what sync needs from whichever emulator is configured. The
// Arcturus and Pixels adapters both satisfy this; Service never imports either
// concrete adapter package.
type EmulatorCatalog interface {
	// ListDefinitions returns every furniture definition the emulator has.
	ListDefinitions(ctx context.Context) ([]Definition, error)
	// InsertDefinitions batch-inserts new definitions into the emulator. Callers
	// only ever pass classnames confirmed missing; adapters do not need upsert
	// semantics.
	InsertDefinitions(ctx context.Context, definitions []Definition) error
	// UpdateDefinitions batch-updates public_name/description on existing
	// definitions. Callers only ever pass classnames confirmed present on both
	// sides with a naming difference.
	UpdateDefinitions(ctx context.Context, definitions []Definition) error
}

// Definition is one furniture definition, mapped between the client and an
// emulator. Field mapping and its rationale are documented in plan/EMULATORS.md.
type Definition struct {
	// Classname is the stable technical identifier (the sync key). Already
	// stripped of any "*N" color-index suffix.
	Classname string
	// Kind is "floor" or "wall".
	Kind string
	// PublicName is the display name.
	PublicName string
	// Description is the display description.
	Description string
	// Width is the footprint width.
	Width int
	// Length is the footprint length.
	Length int
	// AllowWalk mirrors the client's canstandon flag.
	AllowWalk bool
	// AllowSit mirrors the client's cansiton flag.
	AllowSit bool
	// AllowLay mirrors the client's canlayon flag.
	AllowLay bool
}

// Report is the result of a sync furniture check.
type Report struct {
	// Missing lists classnames the client declares that the emulator lacks.
	Missing []string
	// Orphaned lists classnames the emulator has that the client no longer declares.
	// Informational only: sync never deletes these.
	Orphaned []string
	// NameChanges lists classnames present on both sides where the emulator's
	// public_name/description differs from the client's. Apply resolves these by
	// overwriting the emulator's copy with the client's.
	NameChanges []NameChange
}

// OK reports whether every client-declared classname exists in the emulator.
func (report Report) OK() bool {
	return len(report.Missing) == 0
}

// NameChange is one classname where the emulator's naming differs from the
// client's copy.
type NameChange struct {
	// Classname identifies the furniture definition.
	Classname string
	// ClientName is the client's current display name.
	ClientName string
	// EmulatorName is the emulator's current display name.
	EmulatorName string
	// ClientDescription is the client's current display description.
	ClientDescription string
	// EmulatorDescription is the emulator's current display description.
	EmulatorDescription string
}

// ApplyResult summarizes what Apply changed in the emulator.
type ApplyResult struct {
	// Created lists classnames inserted because the emulator was missing them.
	Created []string
	// Updated lists classnames whose public_name/description was overwritten to
	// match the client.
	Updated []string
}
