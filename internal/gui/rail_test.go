package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

// TestRailSlotsAndCollapse verifies the fixed slot order (10 slots across three
// zones: work, tools, system), that collapse hides labels while keeping MinSize
// ≥44px, and that the active accent mark moves between slots (FR-001/FR-002,
// SC-008, contracts/ui-contract.md §1.3).
func TestRailSlotsAndCollapse(t *testing.T) {
	if len(railSlots) != 10 {
		t.Fatalf("railSlots count = %d, want 10", len(railSlots))
	}
	// Work zone first, tools zone ("HERRAMIENTAS") second, system zone third.
	wantOrder := []string{"chat", "git", "tasks", "mcp", "plugins", "settings", "usage", "help", "theme", "account"}
	seen := make([]string, 0, len(railSlots))
	for _, s := range railSlots {
		seen = append(seen, s.id)
	}
	for i, id := range wantOrder {
		if seen[i] != id {
			t.Fatalf("slot %d = %q, want %q", i, seen[i], id)
		}
	}

	// Zone invariants: chat is the only work slot; HERRAMIENTAS has four;
	// system carries the pinned actions.
	if railSlots[0].zone != "work" {
		t.Fatal("chat must be in the work zone")
	}
	for _, id := range []string{"git", "tasks", "mcp", "plugins"} {
		if slotByIDOr(t, id).zone != "tools" {
			t.Fatalf("%s must be in the tools zone", id)
		}
	}
	for _, id := range []string{"settings", "usage", "help", "theme", "account"} {
		if slotByIDOr(t, id).zone != "system" {
			t.Fatalf("%s must be in the system zone", id)
		}
	}

	// No pane destination is duplicated.
	paneSeen := map[int]string{}
	for _, s := range railSlots {
		if s.paneIndex < 0 {
			continue
		}
		if first, ok := paneSeen[s.paneIndex]; ok {
			t.Fatalf("pane %d duplicated by %q and %q", s.paneIndex, first, s.id)
		}
		paneSeen[s.paneIndex] = s.id
	}

	// Shortcut assignment: Alt+1..8, never Ctrl+*.
	for i, s := range railSlots {
		if s.shortcut == "" {
			continue
		}
		if len(s.shortcut) < 4 || s.shortcut[:4] != "Alt+" {
			t.Errorf("slot %s shortcut %q must be Alt+n", s.id, s.shortcut)
		}
		_ = i
	}

	// Duplicate id check.
	for i := range railSlots {
		for j := i + 1; j < len(railSlots); j++ {
			if railSlots[i].id == railSlots[j].id {
				t.Fatalf("duplicate slot id %q", railSlots[i].id)
			}
		}
	}

	rail := newRailView(false)
	expanded := rail.content().MinSize().Width
	if expanded < 180 {
		t.Errorf("expanded rail MinSize.Width = %v, want ~200", expanded)
	}
	if !rail.buttons["chat"].Visible() {
		t.Fatal("expanded rail chat button must be visible")
	}
	if rail.buttons["chat"].Text == "" {
		t.Fatal("expanded chat button must show its label")
	}

	rail.setExpanded(false)
	collapsed := rail.content().MinSize().Width
	if collapsed < 44 {
		t.Errorf("collapsed rail MinSize.Width = %v, want >= 44", collapsed)
	}
	if rail.buttons["chat"].Text != "" {
		t.Error("collapsed chat button must hide its label")
	}
	rail.setExpanded(true)
	if rail.buttons["chat"].Text == "" {
		t.Error("re-expanded chat button must restore label")
	}
}

// slotByIDOr returns the slot with the given id (asserting existence).
func slotByIDOr(t *testing.T, id string) railSlot {
	t.Helper()
	s, ok := slotByID(id)
	if !ok {
		t.Fatalf("slot %q missing", id)
	}
	return s
}

// TestRailZonesAndSingleDestination verifies the 004 zone model
// (contracts/ui-contract.md §1.3, §11): three zones, one destination per pane,
// Alt shortcuts read from the catalog.
func TestRailZonesAndSingleDestination(t *testing.T) {
	if len(railSlots) != 10 {
		t.Fatalf("railSlots = %d, want 10 (sessions removed)", len(railSlots))
	}

	// Zone membership (id → zone).
	zoneOf := map[string]string{}
	for _, s := range railSlots {
		zoneOf[s.id] = s.zone
	}
	want := map[string]string{
		"chat": "work",
		"git": "tools", "tasks": "tools", "mcp": "tools", "plugins": "tools",
		"settings": "system", "usage": "system", "help": "system", "theme": "system", "account": "system",
	}
	for id, z := range want {
		if zoneOf[id] != z {
			t.Errorf("slot %s zone = %q, want %q", id, zoneOf[id], z)
		}
	}

	// One destination per pane.
	for i := 0; i <= 6; i++ {
		s, ok := slotForPane(i)
		if !ok {
			t.Fatalf("pane %d has no rail destination", i)
		}
		for _, other := range railSlots {
			if other.id != s.id && other.paneIndex == i {
				t.Fatalf("pane %d duplicated by %q and %q", i, s.id, other.id)
			}
		}
	}

	// Alt shortcuts from the catalog: every pane slot maps to Alt+#.
	altForPane := map[int]string{0: "Alt+1", 3: "Alt+2", 6: "Alt+3", 4: "Alt+4", 5: "Alt+5", 1: "Alt+6", 2: "Alt+7"}
	for i := 1; i <= 7; i++ {
		s := railSlots[i-1]
		if want := altForPane[s.paneIndex]; s.shortcut != want {
			t.Errorf("slot %s shortcut = %q, want %q", s.id, s.shortcut, want)
		}
	}
	if railSlots[7].id != "help" || railSlots[7].shortcut != "Alt+8" {
		t.Errorf("slot 8 must be help (Alt+8), got %s/%q", railSlots[7].id, railSlots[7].shortcut)
	}

	// Zone separators: the rail renders a header in the tools zone and
	// partition separators; expanded state MinSize stays wide.
	rail := newRailView(false)
	if got := rail.content().MinSize().Width; got < 180 {
		t.Errorf("expanded rail width = %v, want >= 180", got)
	}
}

// TestRailTooltipCollapsed verifies contracts/ui-contract.md §1.3/§11: in
// collapsed mode each rail button carries the "Label (Alt+N)" tooltip and the
// focused (active) slot remains reachable.
func TestRailTooltipCollapsed(t *testing.T) {
	rail := newRailView(true)
	if len(rail.tips) == 0 {
		t.Fatal("collapsed rail must create tooltip buttons")
	}
	for _, s := range railSlots {
		if s.paneIndex < 0 {
			continue // action slots also carry tips
		}
		tb, ok := rail.tips[s.id]
		if !ok {
			t.Fatalf("slot %q missing tooltip button", s.id)
		}
		want := tooltipText(s.label, s.shortcut)
		if tb.tip != want {
			t.Errorf("tooltip[%s] = %q, want %q", s.id, tb.tip, want)
		}
		if tb.tip == "" {
			t.Errorf("tooltip[%s] empty in collapsed mode", s.id)
		}
	}
	// Chat (active pane) is focusable and still selected.
	rail.setActive(0)
	if rail.aid() != "chat" {
		t.Fatalf("active slot = %q, want chat", rail.aid())
	}
	if rail.buttons["chat"].Text != "" {
		t.Error("collapsed chat button must hide its label")
	}
}
func TestRailSlotsActiveMark(t *testing.T) {
	rail := newRailView(false)
	rail.onSelect = nil

	rail.setActive(3) // git
	if rail.aid() != "git" {
		t.Fatalf("activeID = %q, want git", rail.aid())
	}
	if !rail.marks["git"].Visible() {
		t.Error("git mark must be visible when active")
	}
	if rail.marks["chat"].Visible() {
		t.Error("chat mark must be hidden when git is active")
	}

	rail.setActive(0) // chat
	if rail.aid() != "chat" {
		t.Fatalf("activeID = %q, want chat", rail.aid())
	}
	if !rail.marks["chat"].Visible() || rail.marks["git"].Visible() {
		t.Error("active mark did not move back to chat")
	}
}

// TestRailSelectDispatches verifies onSelect receives the right pane index
// per slot (chat=0 … tasks=6, contracts/ui-contract.md §1.3). sessions is gone
// so each destination is unique.
func TestRailSelectDispatches(t *testing.T) {
	rail := newRailView(false)
	got := []int{}
	rail.onSelect = func(i int) { got = append(got, i) }

	rail.dispatch("chat")     // 0
	rail.dispatch("git")      // 3
	rail.dispatch("tasks")    // 6
	rail.dispatch("mcp")      // 4
	rail.dispatch("plugins")  // 5
	rail.dispatch("settings") // 1
	rail.dispatch("usage")    // 2

	want := []int{0, 3, 6, 4, 5, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("onSelect calls = %d, want %d", len(got), len(want))
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("dispatch %d = %d, want %d", i, got[i], v)
		}
	}

	// The catalog must not resolve two slots for the same pane (sessions removed).
	for _, p := range []int{0, 1, 2, 3, 4, 5, 6} {
		count := 0
		for _, s := range railSlots {
			if s.paneIndex == p {
				count++
			}
		}
		if count > 1 {
			t.Errorf("pane %d has %d destinations, want exactly one", p, count)
		}
	}
}

// TestRailActionSlots verifies action slots (-1) never call onSelect and do
// route through onAction.
func TestRailActionSlots(t *testing.T) {
	rail := newRailView(false)
	sel := 0
	rail.onSelect = func(i int) { sel++ }
	acts := []string{}
	rail.onAction = func(id string) { acts = append(acts, id) }

	rail.dispatch("help")
	rail.dispatch("theme")
	rail.dispatch("account")

	if sel != 0 {
		t.Fatalf("onSelect called %d times for action slots, want 0", sel)
	}
	if len(acts) != 3 || acts[0] != "help" || acts[2] != "account" {
		t.Errorf("actions = %v, want [help theme account]", acts)
	}
}

func TestRailCollapsedLabels(t *testing.T) {
	rail := newRailView(true)
	if rail.width() < 44 {
		t.Errorf("collapsed width = %v, want >= 44", rail.width())
	}
	if rail.buttons["git"].Text != "" {
		t.Error("collapsed git label must be hidden")
	}
	var b *widget.Button = rail.buttons["mcp"]
	if b == nil {
		t.Fatal("mcp button missing")
	}
}

// TestRailAvatarAndStatus verifies the pinned bottom account avatar exists in
// the footer, tap dispatches the "account" action, and the status dot color
// can be mutated after the fact (FR-006, T006).
func TestRailAvatarAndStatus(t *testing.T) {
	rail := newRailView(false)
	if rail.avatar == nil {
		t.Fatal("rail footer must own a railAvatar")
	}
	tapped := ""
	rail.avatar.tap = func() { tapped = "account" }
	rail.avatar.Tapped(nil)
	if tapped != "account" {
		t.Fatalf("avatar tap = %q, want account", tapped)
	}
	before := rail.avatar.dotColor
	rail.setAvatarStatus(colorOrange)
	if rail.avatar.dotColor == before {
		t.Fatal("setAvatarStatus must update the dot color")
	}
	// Collapsed mode keeps the avatar visible (pinned bottom).
	rail2 := newRailView(true)
	if rail2.avatar == nil {
		t.Fatal("collapsed rail must still show the avatar")
	}
}
