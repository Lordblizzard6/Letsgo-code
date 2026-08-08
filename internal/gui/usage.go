package gui

import (
	"fmt"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/tools"
)

// usageSource abstracts cost aggregation so tests can inject deterministic
// statistics without touching the user's costs.json.
type usageSource interface {
	// Stats returns aggregated usage since the given time (FR-009).
	Stats(since time.Time) map[string]interface{}
	// TotalCost returns the summed cost since the given time.
	TotalCost(since time.Time) float64
}

// costTrackerUsageSource adapts internal/tools.CostTracker to usageSource.
type costTrackerUsageSource struct {
	tracker *tools.CostTracker
}

func (s costTrackerUsageSource) Stats(since time.Time) map[string]interface{} {
	return s.tracker.GetUsageStats(since)
}

func (s costTrackerUsageSource) TotalCost(since time.Time) float64 {
	return s.tracker.GetTotalCost(since)
}

// usageHistogram is a pure aggregation of provider/model usage.
type usageHistogram struct {
	requests  int
	inTokens  int
	outTokens int
	cost      float64
	byModel   []modelUsage
	byProv    []providerUsage
}

type modelUsage struct {
	model        string
	provider     string
	requests     int
	inputTokens  int
	outputTokens int
	costUSD      float64
}

type providerUsage struct {
	provider string
	requests int
	tokens   int
	costUSD  float64
}

// aggregateUsage builds a per-provider + per-model histogram from
// CostTracker stats.
func aggregateUsage(stats map[string]interface{}) *usageHistogram {
	h := &usageHistogram{}
	if stats == nil {
		return h
	}
	if v, ok := stats["total_requests"].(int); ok {
		h.requests = v
	}
	if v, ok := stats["total_input_tokens"].(int); ok {
		h.inTokens = v
	}
	if v, ok := stats["total_output_tokens"].(int); ok {
		h.outTokens = v
	}
	if v, ok := stats["total_cost_usd"].(float64); ok {
		h.cost = v
	}

	byModel, _ := stats["by_model"].(map[string]map[string]interface{})
	providers := map[string]providerUsage{}
	for model, m := range byModel {
		if m == nil {
			continue
		}
		mu := modelUsage{model: model}
		if v, ok := m["requests"].(int); ok {
			mu.requests = v
		}
		if v, ok := m["input_tokens"].(int); ok {
			mu.inputTokens = v
		}
		if v, ok := m["output_tokens"].(int); ok {
			mu.outputTokens = v
		}
		if v, ok := m["cost_usd"].(float64); ok {
			mu.costUSD = v
		}
		mu.provider = config.DetectProviderFromModel(model)
		h.byModel = append(h.byModel, mu)

		p := providers[mu.provider]
		p.provider = mu.provider
		p.requests += mu.requests
		p.tokens += mu.inputTokens + mu.outputTokens
		p.costUSD += mu.costUSD
		providers[mu.provider] = p
	}
	for _, p := range providers {
		h.byProv = append(h.byProv, p)
	}
	sort.Slice(h.byModel, func(i, j int) bool { return h.byModel[i].model < h.byModel[j].model })
	sort.Slice(h.byProv, func(i, j int) bool { return h.byProv[i].provider < h.byProv[j].provider })
	return h
}

// String renders the histogram as monospace summary text (kept for parity
// with the pre-004 usage pane and any external readers).
func (h *usageHistogram) String() string {
	var b string
	b += fmt.Sprintf("Requests: %d\n", h.requests)
	b += fmt.Sprintf("Input tokens: %d\n", h.inTokens)
	b += fmt.Sprintf("Output tokens: %d\n", h.outTokens)
	b += fmt.Sprintf("Cost: $%.4f\n", h.cost)

	if len(h.byProv) > 0 {
		b += "\nPer provider:\n"
		for _, p := range h.byProv {
			b += fmt.Sprintf("  %s: %d req, %d tok, $%.4f\n", p.provider, p.requests, p.tokens, p.costUSD)
		}
	}
	if len(h.byModel) > 0 {
		b += "\nPer model:\n"
		for _, m := range h.byModel {
			b += fmt.Sprintf("  %s: %d req, %d→%d tok, $%.4f\n",
				m.model, m.requests, m.inputTokens, m.outputTokens, m.costUSD)
		}
	}
	return b
}

// usageView is the usage/cost dashboard panel (FR-009). It aggregates the
// current period's provider usage from internal/tools.CostTracker into KPI
// cards (cost today, requests, tokens), a per-provider table and a budget
// progress bar (004 US2, contracts §4).
type usageView struct {
	win  fyne.Window
	root *fyne.Container

	cards         []*cardLabel
	provTable     *widget.Table
	budget        *widget.ProgressBar
	center        *emptyAware

	providerRows map[string]providerUsage
	rowOrder     []string

	source    usageSource
	sinceFunc func() time.Time
}

// cardLabel is one KPI headline: value (bold) over caption (muted).
type cardLabel struct {
	value   *widget.Label
	caption *widget.Label
}

func newCard(value, caption string) *cardLabel {
	return &cardLabel{
		value:   widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		caption: widget.NewLabelWithStyle(caption, fyne.TextAlignCenter, fyne.TextStyle{}),
	}
}

// set updates the value label (typographic ramp: heading-size headline).
func (c *cardLabel) set(value string) {
	c.value.SetText(value)
	c.value.SizeName = theme.SizeNameHeadingText
}

func (c *cardLabel) content() fyne.CanvasObject {
	c.caption.Importance = widget.MediumImportance
	c.caption.SizeName = theme.SizeNameCaptionText
	return container.NewVBox(c.value, c.caption)
}

// newUsageView builds the usage dashboard bound to the singleton CostTracker.
func newUsageView(win fyne.Window) *usageView {
	v := &usageView{
		win:       win,
		source:    costTrackerUsageSource{tracker: tools.GetCostTracker()},
		sinceFunc: startOfToday,
	}

	headers := []string{"Coste hoy", "Requests", "Tokens"}
	v.cards = make([]*cardLabel, len(headers))
	var cardObs []fyne.CanvasObject
	for i, label := range headers {
		v.cards[i] = newCard("", label)
		cardObs = append(cardObs, v.cards[i].content())
	}
	grid := container.NewGridWithColumns(3, cardObs...)

	// Per-provider table: Provider | Requests | Tokens | Cost.
	v.provTable = widget.NewTable(
		func() (int, int) { return 1, 4 },
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Alignment = fyne.TextAlignLeading
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				caption := []string{"Proveedor", "Requests", "Tokens", "Coste"}
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText(caption[id.Col])
				return
			}
			l.TextStyle = fyne.TextStyle{}
			p := v.providerRows[v.rowOrder[id.Row-1]]
			switch id.Col {
			case 0:
				l.SetText(p.provider)
			case 1:
				l.SetText(fmt.Sprintf("%d", p.requests))
			case 2:
				l.SetText(fmt.Sprintf("%d", p.tokens))
			default:
				l.SetText(fmt.Sprintf("$%.2f", p.costUSD))
			}
		},
	)
	// Recompute row count after refresh; table Length closure reads the live slices.
	v.provTable.Length = func() (int, int) { return len(v.rowOrder) + 1, 4 }
	v.provTable.SetColumnWidth(0, 150)
	v.provTable.SetColumnWidth(1, 90)
	v.provTable.SetColumnWidth(2, 90)
	v.provTable.SetColumnWidth(3, 90)

	v.budget = widget.NewProgressBar()

	top := container.NewHBox(
		widget.NewLabelWithStyle("Uso hoy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	refresh := widget.NewButton("Actualizar", v.refresh)
	top.Add(refresh)

	dashboard := container.NewVBox(
		grid,
		widget.NewLabelWithStyle("Por proveedor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.provTable,
		widget.NewLabelWithStyle("Presupuesto", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.budget,
	)
	v.center = newEmptyAware(dashboard, "Sin actividad registrada hoy.", "", nil)
	v.root = container.NewBorder(
		top,
		nil, nil, nil,
		v.center.content(),
	)
	return v
}

// content returns the dashboard for embedding in the window.
func (v *usageView) content() fyne.CanvasObject { return v.root }

// refresh re-reads CostTracker stats and repaints KPI cards, table and budget.
func (v *usageView) refresh() {
	now := v.sinceFunc()
	var h *usageHistogram
	if v.source != nil {
		h = aggregateUsage(v.source.Stats(now))
		total := v.source.TotalCost(now)
		if total > h.cost {
			h.cost = total
		}
	}
	if h == nil {
		h = aggregateUsage(nil)
	}
	v.cards[0].set(fmt.Sprintf("$%.2f", h.cost))
	v.cards[1].set(fmt.Sprintf("%d", h.requests))
	v.cards[2].set(fmt.Sprintf("%d", h.inTokens+h.outTokens))

	// Total budget for the bar: a generous cap so today's cost never pegs 100%.
	budgetCap := 25.0
	if h.cost >= budgetCap {
		v.budget.SetValue(1)
	} else {
		v.budget.SetValue(h.cost / budgetCap)
	}

	v.providerRows = map[string]providerUsage{}
	v.rowOrder = v.rowOrder[:0]
	for _, p := range h.byProv {
		v.providerRows[p.provider] = p
		v.rowOrder = append(v.rowOrder, p.provider)
	}
	v.provTable.Length = func() (int, int) { return len(v.rowOrder) + 1, 4 }
	v.provTable.Refresh()
	if v.center != nil {
		v.center.setEmpty(len(v.rowOrder) == 0 && h.cost == 0)
	}
}

// startOfToday rounds the clock to local midnight, the usage period start.
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}