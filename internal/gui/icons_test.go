package gui

import (
	"strings"
	"testing"
)

// TestIconForResolvesAllRailSlots verifies every rail destination resolves a
// valid resource (SVG embed or stock fallback) without panicking.
func TestIconForResolveAllRailSlots(t *testing.T) {
	for _, id := range railSlotIDs() {
		res := iconFor(id)
		if res == nil || res.Name() == "" {
			t.Fatalf("iconFor(%q) returned an empty resource", id)
		}
	}
}

// TestIconForSVGForKnownIDs verifies known ids embed an actual SVG payload
// (starts with the SVG opening tag), not just the theme fallback.
func TestIconForSVGForKnownIDs(t *testing.T) {
	for id, svg := range iconSVGs {
		res := iconFor(id)
		data := string(res.Content())
		if !strings.HasPrefix(data, "<svg") {
			t.Errorf("iconFor(%q) is not an embedded SVG: %q", id, string(data[:minLen(len(data), 20)]))
		}
		// Map and SVG name agree.
		if !strings.Contains(svg, "<svg") {
			t.Errorf("iconSVGs[%q] is not an SVG", id)
		}
	}
}

// TestIconForUnknownFallsBack verifies unknown ids return the Fyne fallback
// icon instead of nil.
func TestIconForUnknownFallsBack(t *testing.T) {
	res := iconFor("does-not-exist")
	if res == nil || res.Name() == "" {
		t.Fatal("unknown id must still resolve a fallback resource")
	}
}

// railSlotIDs returns the ids of the current catalog (kept in sync with the
// rail so the icon coverage test tracks the real navigation surface).
func railSlotIDs() []string {
	ids := make([]string, 0, len(railSlots))
	for _, s := range railSlots {
		ids = append(ids, s.id)
	}
	return ids
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}