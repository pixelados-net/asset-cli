package sync

import (
	"testing"

	"github.com/pixelados-net/asset-cli/internal/furniture/gamedata"
)

func TestMapGamedataDefinitions(t *testing.T) {
	definitions := []gamedata.Definition{
		{
			Classname: "throne_gold", Kind: gamedata.KindFloor, Name: "Gold Throne", Description: "Shiny",
			Width: 2, Length: 2, CanStandOn: false, CanSitOn: true, CanLayOn: false,
		},
	}
	mapped := mapGamedataDefinitions(definitions)
	if len(mapped) != 1 {
		t.Fatalf("mapped = %#v", mapped)
	}
	got := mapped[0]
	want := Definition{
		Classname: "throne_gold", Kind: KindFloor, PublicName: "Gold Throne", Description: "Shiny",
		Width: 2, Length: 2, AllowWalk: false, AllowSit: true, AllowLay: false,
	}
	if got != want {
		t.Fatalf("mapped[0] = %#v, want %#v", got, want)
	}
}
