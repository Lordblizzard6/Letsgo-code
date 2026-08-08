package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// fakeUsageSource returns deterministic KPI data for usage tests.
type fakeUsageSource struct{}

func (fakeUsageSource) Stats(since time.Time) map[string]interface{} {
	return map[string]interface{}{
		"total_requests":       42,
		"total_input_tokens":   1000,
		"total_output_tokens":  500,
		"total_cost_usd":       1.25,
		"by_model": map[string]map[string]interface{}{
			"claude-sonnet-4-20250514": {
				"requests":      30,
				"input_tokens":  800,
				"output_tokens": 400,
				"cost_usd":      1.0,
			},
			"gpt-4o": {
				"requests":      12,
				"input_tokens":  200,
				"output_tokens": 100,
				"cost_usd":      0.25,
			},
		},
	}
}

func (fakeUsageSource) TotalCost(since time.Time) float64 { return 1.25 }

// cheapSource is a lower-usage usageSource used to verify re-rendering.
type cheapSource struct{ fakeUsageSource }

func (cheapSource) Stats(since time.Time) map[string]interface{} {
	return map[string]interface{}{
		"total_requests":      2,
		"total_input_tokens":  10,
		"total_output_tokens": 5,
		"total_cost_usd":      0.01,
		"by_model":            map[string]map[string]interface{}{},
	}
}

func (cheapSource) TotalCost(since time.Time) float64 { return 0.01 }

// TestUsageKPISurfaces verifies the US2 KPI dashboard (contracts §4): three
// headline cards (cost today, requests, tokens), one row per provider in the
// per-provider table, and a budget progress bar driven by aggregate usage.
// It must pass without depending on usageHistogram.String (T016).
func TestUsageKPISurfaces(t *testing.T) {
	win := test.NewWindow(container.NewStack())
	defer win.Close()

	v := newUsageView(win)
	v.source = fakeUsageSource{}
	v.sinceFunc = func() time.Time { return time.Now().Add(-time.Hour) }
	v.refresh()

	if len(v.cards) != 3 {
		t.Fatalf("KPI cards = %d, want 3 (cost, requests, tokens)", len(v.cards))
	}
	cardText := usageCardText(v.cards)
	for _, want := range []string{"1.25", "42", "1500"} {
		if !strings.Contains(cardText, want) {
			t.Errorf("KPI cards missing %q in:\n%s", want, cardText)
		}
	}

	if v.provTable == nil {
		t.Fatal("provider table missing")
	}
	rows, _ := v.provTable.Length()
	if rows != 3 {
		t.Fatalf("provider table rows = %d, want 3 (header + 2 providers)", rows)
	}
	// Second row = first provider (alphabetical: anthropic).
	c := v.provTable.CreateCell()
	v.provTable.UpdateCell(widget.TableCellID{Row: 1, Col: 0}, c)
	if l, ok := c.(*widget.Label); !ok || !strings.Contains(l.Text, "anthropic") {
		t.Fatalf("provider row 1 = %v, want anthropic", l)
	}

	if v.budget == nil {
		t.Fatal("budget bar missing")
	}
	if v.budget.Value <= 0 {
		t.Fatalf("budget bar value = %.4f, want > 0 for $1.25 spent", v.budget.Value)
	}
}

// TestUsageRefreshKeepsPanels verifies refresh() repaints KPI cards and table
// without erroring when the source changes (T021).
func TestUsageRefreshKeepsPanels(t *testing.T) {
	win := test.NewWindow(container.NewStack())
	defer win.Close()

	v := newUsageView(win)
	v.source = fakeUsageSource{}
	v.sinceFunc = func() time.Time { return time.Now().Add(-time.Hour) }
	v.refresh()
	before := usageCardText(v.cards)

	// A cheaper source: same shape, different values.
	v.source = cheapSource{}
	v.refresh()
	after := usageCardText(v.cards)
	if after == before {
		t.Fatal("refresh() must repaint KPI cards")
	}
}

// usageCardText extracts the label text from KPI cards.
func usageCardText(cards []*cardLabel) string {
	var b strings.Builder
	for _, c := range cards {
		b.WriteString(c.value.Text)
		b.WriteString(" ")
		b.WriteString(c.caption.Text)
		b.WriteString("\n")
	}
	return b.String()
}