package gui

import (
	"fmt"
	"image/color"
	"os"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

// providerModelIDs mirrors internal/tui model tables (Anthropic, OpenAI, Groq,
// OpenRouter, Ollama).
var providerModelIDs = [][]string{
	{"claude-sonnet-4-20250514", "claude-3-opus-20240229"},
	{"gpt-4o", "gpt-4o-mini"},
	{"llama-3.3-70b-versatile", "llama-3.1-8b-instant", "openai/gpt-oss-120b", "openai/gpt-oss-20b"},
	{"anthropic/claude-sonnet-4", "openai/gpt-4o", "meta-llama/llama-3.3-70b-instruct"},
	{"llama3.2", "codellama"},
}

var providerNames = []string{"Anthropic", "OpenAI", "Groq", "OpenRouter", "Ollama"}

// effortLabels mirrors cmd/effort.go level names.
var effortLabels = []string{"0 - minimal", "1 - low", "2 - normal", "3 - high", "4 - very high", "5 - maximum"}

// outputStyles mirrors cmd/output_style.go available styles.
var outputStyles = []string{"markdown", "plain", "compact"}

// platformModes mirrors cmd/config.go platform mode values.
var platformModes = []string{"desktop", "mobile", "ide", "chrome"}

// permTools mirrors the tool names managed by internal/tools permission modes.
var permTools = []string{"bash", "edit", "web_search", "web_fetch"}

// settingsView is the settings surface (FR-007/FR-011): provider, API key,
// model, plus parity rows for the CLI config group (cli-parity.md) and
// provider auth. 004 (US2) regroups the content into a 4-tab TabContainer
// (Cuenta · Apariencia · Preferencias · Uso) while preserving every field.
type settingsView struct {
	tabs *container.AppTabs

	providerSel  *widget.RadioGroup
	modelSel     *widget.Select
	apiKey       *widget.Entry
	effortSel    *widget.Select
	styleSel     *widget.Select
	themeSel     *widget.Select
	railChk      *widget.Check
	vimChk       *widget.Check
	statusLine   *widget.Entry
	platformSel  *widget.Select
	sandboxChk   *widget.Check
	analyticsChk *widget.Check
	rateLimit    *widget.Entry
	envLabel     *widget.Label

	permSelects map[string]*widget.Select

	// Status line field toggles (FR-018).
	statusChecks map[string]*widget.Check

	// Session grants section (FR-017): list/revoke per-session approval grants.
	grantsBox *fyne.Container

	saveBtn *widget.Button

	contentObj fyne.CanvasObject
	win        fyne.Window

	// onSaved is invoked after a successful config save (used to refresh clients).
	onSaved func()

	// onStatusLineChanged fires when a status line field is toggled (FR-018).
	onStatusLineChanged func()

	// onCancel closes the overlay without persisting (US2); today it mirrors
	// the pane behaviour so Esc never leaves the user mid-edit.
	onCancel func()

	// onOpenUsage jumps to the usage dashboard from the Uso tab (US2).
	onOpenUsage func()
}

// newSettingsView builds the settings form bound to the current config.
func newSettingsView(win fyne.Window) *settingsView {
	v := &settingsView{win: win, permSelects: map[string]*widget.Select{}}
	cfg := config.AppConfig

	v.modelSel = widget.NewSelect(nil, nil)

	v.providerSel = widget.NewRadioGroup(providerNames, func(selected string) {
		for i, name := range providerNames {
			if name == selected {
				v.modelSel.Options = providerModelIDs[i]
				v.modelSel.SetSelected(currentModelForProvider(i))
				return
			}
		}
	})
	v.providerSel.Horizontal = true
	provider := config.DetectProviderFromModel(cfg.Model)
	selectedProvider := providerNameFor(provider)
	v.modelSel.SetSelected(cfg.Model)
	v.providerSel.SetSelected(selectedProvider)

	v.apiKey = widget.NewEntry()
	v.apiKey.SetPlaceHolder("API key (per provider)")
	v.apiKey.Password = true
	v.apiKey.SetText(currentAPIKey())

	// Parity config rows (cli-parity.md Config group).
	v.effortSel = widget.NewSelect(effortLabels, nil)
	if lvl := viper.GetInt("effort"); lvl >= 0 && lvl < len(effortLabels) {
		v.effortSel.SetSelected(effortLabels[lvl])
	}

	v.styleSel = widget.NewSelect(outputStyles, nil)
	if style := viper.GetString("output_style"); style != "" {
		v.styleSel.SetSelected(style)
	}

	// theme: dark is the only shipped variant in v1 (gui-contract.md).
	v.themeSel = widget.NewSelect([]string{"dark", "light"}, nil)
	if cfg.ThemeVariant != "" {
		v.themeSel.SetSelected(cfg.ThemeVariant)
	} else {
		v.themeSel.SetSelected("dark")
	}

	v.railChk = widget.NewCheck("Collapse side rail", func(collapsed bool) {
		cfg.Rail.Collapsed = collapsed
	})
	v.railChk.SetChecked(cfg.Rail.Collapsed)

	v.vimChk = widget.NewCheck("Vim editing mode", nil)
	v.vimChk.SetChecked(viper.GetBool("vim"))

	v.statusLine = widget.NewEntry()
	if sl := viper.GetString("statusline"); sl != "" {
		v.statusLine.SetText(sl)
	}

	v.platformSel = widget.NewSelect(platformModes, nil)
	if mode := viper.GetString("platform"); mode != "" && containsStr(platformModes, mode) {
		v.platformSel.SetSelected(mode)
	} else {
		v.platformSel.SetSelected("desktop")
	}

	v.sandboxChk = widget.NewCheck("Enable sandbox mode", nil)
	v.sandboxChk.SetChecked(viper.GetBool("sandbox"))

	v.analyticsChk = widget.NewCheck("Share usage analytics", nil)
	v.analyticsChk.SetChecked(!viper.GetBool("privacy_analytics"))

	v.rateLimit = widget.NewEntry()
	v.rateLimit.SetPlaceHolder("requests / minute (0 = unlimited)")
	if rpm := viper.GetInt("rate_limit_requests"); rpm > 0 {
		v.rateLimit.SetText(fmt.Sprintf("%d", rpm))
	}

	v.envLabel = widget.NewLabel("")
	v.envLabel.TextStyle = fyne.TextStyle{Monospace: true}
	v.envLabel.Wrapping = fyne.TextWrapWord
	v.envLabel.SetText(formatEnvList(os.Environ()))

	form := widget.NewForm(
		widget.NewFormItem("Provider", v.providerSel),
		widget.NewFormItem("Model", v.modelSel),
		widget.NewFormItem("API key", v.apiKey),
	)
	parityForm := widget.NewForm(
		widget.NewFormItem("Effort", v.effortSel),
		widget.NewFormItem("Output style", v.styleSel),
		widget.NewFormItem("Theme", v.themeSel),
		widget.NewFormItem("Rail", v.railChk),
		widget.NewFormItem("Vim mode", v.vimChk),
		widget.NewFormItem("Status line", v.statusLine),
		widget.NewFormItem("Platform", v.platformSel),
		widget.NewFormItem("Sandbox", v.sandboxChk),
		widget.NewFormItem("Analytics", v.analyticsChk),
		widget.NewFormItem("Rate limit", v.rateLimit),
	)

	authRow := container.NewHBox(
		widget.NewButton("Login", v.login),
		widget.NewButton("Logout", v.logout),
		widget.NewButton("OAuth refresh", v.oauthRefresh),
	)

	envScroll := container.NewScroll(v.envLabel)
	envScroll.SetMinSize(fyne.NewSize(380, 120))

	v.buildPermissionRows()
	v.grantsBox = container.NewVBox()
	v.refreshGrants()

	// FR-018: status line field toggles bound to statusline.fields.
	v.statusChecks = map[string]*widget.Check{}
	fields := map[string]bool{}
	for _, f := range config.AppConfig.Statusline.Fields {
		fields[f] = true
	}
	fieldNames := []string{"model", "branch", "mode", "context", "rate", "version"}
	fieldRows := container.NewVBox()
	for _, name := range fieldNames {
		chk := widget.NewCheck(name, nil)
		chk.SetChecked(fields[name])
		chk.OnChanged = func(enabled bool) {
			v.applyStatusField(name, enabled)
		}
		v.statusChecks[name] = chk
		fieldRows.Add(chk)
	}

	v.saveBtn = widget.NewButton("Save", v.save)

	// 004 (US2): TabContainer with Cuenta, Apariencia, Preferencias, Uso.
	// Cuenta = provider/credentials/auth; Apariencia = visual + rail;
	// Preferencias = the parity/advanced rows; Uso = budget link.
	title := widget.NewLabelWithStyle("Configuración", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameHeadingText
	toolBar := container.NewHBox(title)
	saveRow := container.NewHBox(v.saveBtn, widget.NewButton("Cancel", v.cancel))
	appearance := container.NewVBox(
		widget.NewLabelWithStyle("Tema", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.themeSel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Rail", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.railChk,
	)
	usageTab := container.NewVBox(
		widget.NewLabelWithStyle("Presupuesto y actividad", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Consulta el detalle de coste de hoy, requests y tokens en el panel Uso (Alt+7 / Ctrl+U)."),
		container.NewHBox(widget.NewButton("Abrir Uso", func() {
			if v.onOpenUsage != nil {
				v.onOpenUsage()
			}
		})),
	)

	v.tabs = container.NewAppTabs(
		container.NewTabItem("Cuenta", container.NewVBox(
			form,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Provider auth", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			authRow,
		)),
		container.NewTabItem("Apariencia", appearance),
		container.NewTabItem("Preferencias", container.NewVBox(
			widget.NewLabelWithStyle("Advanced (parity)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			parityForm,
			v.permSection(),
			widget.NewLabelWithStyle("Status line fields (FR-018)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fieldRows,
			widget.NewLabelWithStyle("Session grants (FR-017)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			v.grantsBox,
			widget.NewLabelWithStyle("Environment", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			envScroll,
		)),
		container.NewTabItem("Uso", usageTab),
	)
	v.contentObj = container.NewBorder(
		container.NewVBox(toolBar, widget.NewSeparator()),
		saveRow, nil, nil,
		v.tabs,
	)
	return v
}

// cancel closes the settings surface without persisting pending edits.
func (v *settingsView) cancel() {
	if v.onCancel != nil {
		v.onCancel()
	}
}

// settingsEsc is a focus-only invisible key sink: while it holds focus, Esc
// closes the settings overlay (US2). The composer keeps its own Esc semantics
// because focus returns there after close (SC-009).
type settingsEsc struct {
	widget.BaseWidget
	esc func()
}

func newSettingsEsc(esc func()) *settingsEsc {
	e := &settingsEsc{esc: esc}
	e.ExtendBaseWidget(e)
	return e
}

func (e *settingsEsc) TypedRune(r rune)                             {}
func (e *settingsEsc) FocusGained()                                 {}
func (e *settingsEsc) FocusLost()                                   {}
func (e *settingsEsc) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape && e.esc != nil {
		e.esc()
	}
}

func (e *settingsEsc) MinSize() fyne.Size { return fyne.NewSize(1, 1) }

func (e *settingsEsc) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(r)
}

var _ fyne.Focusable = (*settingsEsc)(nil)

// settingsOverlay is the non-modal settings card (US2, contracts §2): the
// TabContainer floats above the conversation (~660x560) and Esc closes it.
// It lives in viewStack so the headless test driver renders it like any
// other overlay (research D2).
type settingsOverlay struct {
	win fyne.Window

	root fyne.CanvasObject
	esc  *settingsEsc

	onClose func()
}

// newSettingsOverlay builds the floating settings card over the chat shell.
func newSettingsOverlay(win fyne.Window, view *settingsView) *settingsOverlay {
	o := &settingsOverlay{win: win}
	o.esc = newSettingsEsc(func() { o.close() })

	// Cap the card at ~660x520; center it with generous padding.
	card := &fixedSize{CanvasObject: view.content(), w: 660, h: 520}
	surface := container.NewCenter(card)
	o.root = container.NewBorder(
		container.NewHBox(o.esc),
		nil, nil, nil, surface,
	)
	return o
}

// fixedSize enforces a minimum box for the settings card.
type fixedSize struct {
	fyne.CanvasObject
	w, h float32
}

func (f *fixedSize) MinSize() fyne.Size { return fyne.NewSize(f.w, f.h) }

// open makes the card visible and gives Esc the focus.
func (o *settingsOverlay) open() {
	if o.root == nil {
		return
	}
	o.root.Show()
	o.win.Canvas().Focus(o.esc)
}

// close hides the card without acting (Esc / Cancel).
func (o *settingsOverlay) close() {
	if o.root != nil {
		o.root.Hide()
	}
	if o.esc != nil {
		o.win.Canvas().Unfocus()
	}
	if o.onClose != nil {
		o.onClose()
	}
}

// content returns the settings pane for embedding in the main window.
func (v *settingsView) content() fyne.CanvasObject { return v.contentObj }

// permSection renders per-tool permission mode selects.
func (v *settingsView) permSection() fyne.CanvasObject {
	rows := container.NewVBox()
	for _, tool := range permTools {
		tool := tool
		sel := widget.NewSelect([]string{"ask", "auto", "auto-bp", "opt-out"}, nil)
		if tp := tools.GetAdvancedPermissionManager().GetToolConfig(tool); tp != nil {
			mode := string(tp.Mode)
			if mode == "" {
				mode = "ask"
			}
			sel.SetSelected(mode)
		} else {
			sel.SetSelected("ask")
		}
		sel.OnChanged = func(mode string) {
			_ = tools.GetAdvancedPermissionManager().SetToolMode(tool, tools.PermissionMode(mode), tools.ScopeGlobal)
		}
		v.permSelects[tool] = sel
		rows.Add(container.NewHBox(widget.NewLabel(tool), sel))
	}
	return rows
}

// buildPermissionRows ensures permission selects exist for all known tools.
func (v *settingsView) buildPermissionRows() {}

// applyStatusField updates statusline.show/fields live when the user toggles
// a status line piece (FR-018). Saving persists through config.SaveConfig.
func (v *settingsView) applyStatusField(name string, enabled bool) {
	cfg := &config.AppConfig
	if cfg.Statusline.Fields == nil {
		cfg.Statusline.Fields = []string{}
	}
	if enabled {
		for _, f := range cfg.Statusline.Fields {
			if f == name {
				return
			}
		}
		cfg.Statusline.Fields = append(cfg.Statusline.Fields, name)
	} else {
		kept := cfg.Statusline.Fields[:0]
		for _, f := range cfg.Statusline.Fields {
			if f != name && f != "" {
				kept = append(kept, f)
			}
		}
		cfg.Statusline.Fields = kept
	}
	cfg.Statusline.Show = len(cfg.Statusline.Fields) > 0
	if v.onStatusLineChanged != nil {
		v.onStatusLineChanged()
	}
}

// refreshGrants re-renders the per-session grants list (FR-017). Every grant
// is scoped to one session; revoking restores prompts for that category.
func (v *settingsView) refreshGrants() {
	if v.grantsBox == nil {
		return
	}
	v.grantsBox.Objects = nil

	type grantRow struct {
		sessionLabel string
		sessionID    string
		category     string
	}

	header := widget.NewLabelWithStyle("Session grants", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.grantsBox.Add(header)

	sessions, err := db.ListSessions()
	if err != nil {
		v.grantsBox.Add(widget.NewLabel("(cannot load sessions: " + err.Error() + ")"))
		v.grantsBox.Refresh()
		return
	}
	byID := make(map[string]db.Session, len(sessions))
	var rows []grantRow
	for _, ses := range sessions {
		byID[ses.ID] = ses
		grants, err := db.ListGrants(ses.ID)
		if err != nil {
			continue
		}
		for _, g := range grants {
			rows = append(rows, grantRow{
				sessionLabel: shortSessionLabel(ses),
				sessionID:    ses.ID,
				category:     g.Category,
			})
		}
	}
	if len(rows) == 0 {
		v.grantsBox.Add(widget.NewLabel("No session grants. Approve a tool with \"Allow for this session\" to create one."))
		v.grantsBox.Refresh()
		return
	}
	for _, g := range rows {
		g := g
		row := container.NewHBox(
			widget.NewLabel(g.sessionLabel),
			widget.NewLabel(g.category),
			widget.NewButton("Revoke", func() {
				_ = db.RevokeCategory(g.sessionID, g.category)
				v.refreshGrants()
			}),
		)
		v.grantsBox.Add(row)
	}
	v.grantsBox.Refresh()
}

// shortSessionLabel renders a compact session identifier for grant listings.
func shortSessionLabel(s db.Session) string {
	if s.Name != "" {
		return s.Name
	}
	if len(s.ID) > 12 {
		return s.ID[:12] + "…"
	}
	return s.ID
}

// save persists config changes and parity keys, then signals client refresh.
func (v *settingsView) save() {
	cfg := &config.AppConfig
	provider := providerKey(v.providerSel.Selected)
	cfg.Model = v.modelSel.Selected
	if cfg.Model == "" {
		idx := providerIndex(provider)
		if idx >= 0 && len(providerModelIDs[idx]) > 0 {
			cfg.Model = providerModelIDs[idx][0]
		}
	}
	switch provider {
	case "anthropic":
		cfg.AnthropicAPIKey = v.apiKey.Text
	case "openai":
		cfg.OpenAIAPIKey = v.apiKey.Text
	case "groq":
		cfg.GroqAPIKey = v.apiKey.Text
	case "openrouter":
		cfg.OpenRouterAPIKey = v.apiKey.Text
	case "ollama":
		cfg.BaseURL = "http://localhost:11434"
	}
	viper.Set("effort", effortLevel(v.effortSel.Selected))
	viper.Set("output_style", v.styleSel.Selected)
	viper.Set("theme", v.themeSel.Selected)
	cfg.ThemeVariant = v.themeSel.Selected
	cfg.Rail.Collapsed = v.railChk.Checked
	viper.Set("vim", v.vimChk.Checked)
	viper.Set("statusline", v.statusLine.Text)
	viper.Set("statusline.show", config.AppConfig.Statusline.Show)
	viper.Set("statusline.fields", config.AppConfig.Statusline.Fields)
	viper.Set("platform", v.platformSel.Selected)
	viper.Set("sandbox", v.sandboxChk.Checked)
	viper.Set("privacy_analytics", v.analyticsChk.Checked)
	viper.Set("rate_limit_requests", rateLimitValue(v.rateLimit.Text))
	if err := config.SaveConfig(); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save config: %w", err), v.win)
		return
	}
	v.envLabel.SetText(formatEnvList(os.Environ()))
	if v.onSaved != nil {
		v.onSaved()
	}
	dialog.ShowInformation("Settings", "Configuration saved.", v.win)
}

// login saves the current API key for the selected provider.
func (v *settingsView) login() {
	v.save()
	dialog.ShowInformation("Login", "API key saved for "+v.providerSel.Selected, v.win)
}

// logout clears the API key for the selected provider.
func (v *settingsView) logout() {
	cfg := &config.AppConfig
	switch providerKey(v.providerSel.Selected) {
	case "anthropic":
		cfg.AnthropicAPIKey = ""
	case "openai":
		cfg.OpenAIAPIKey = ""
	case "groq":
		cfg.GroqAPIKey = ""
	case "openrouter":
		cfg.OpenRouterAPIKey = ""
	}
	cfg.APIKey = ""
	_ = config.SaveConfig()
	v.apiKey.SetText("")
	dialog.ShowInformation("Settings", "Logged out. API key cleared.", v.win)
}

// oauthRefresh mirrors cmd/oauth_refresh.go informational flow.
func (v *settingsView) oauthRefresh() {
	cfg := config.AppConfig
	if cfg.APIKey == "" && !hasAnyKey() {
		dialog.ShowInformation("OAuth refresh",
			"No API key configured. Sign in first.", v.win)
		return
	}
	dialog.ShowInformation("OAuth refresh",
		fmt.Sprintf("Credential state checked for %s. Tokens auto-refresh; repeat only on auth issues.",
			detectProviderDescription()), v.win)
}

// ---- helpers ----

func providerNameFor(provider string) string {
	switch provider {
	case "anthropic":
		return "Anthropic"
	case "openai":
		return "OpenAI"
	case "groq":
		return "Groq"
	case "openrouter":
		return "OpenRouter"
	case "ollama":
		return "Ollama"
	}
	return "OpenAI"
}

func providerKey(name string) string {
	switch name {
	case "Anthropic":
		return "anthropic"
	case "OpenAI":
		return "openai"
	case "Groq":
		return "groq"
	case "OpenRouter":
		return "openrouter"
	case "Ollama":
		return "ollama"
	}
	return "openai"
}

func providerIndex(provider string) int {
	switch provider {
	case "anthropic":
		return 0
	case "openai":
		return 1
	case "groq":
		return 2
	case "openrouter":
		return 3
	case "ollama":
		return 4
	}
	return -1
}

func currentModelForProvider(providerIdx int) string {
	if providerIdx < 0 || providerIdx >= len(providerModelIDs) {
		return ""
	}
	for _, m := range providerModelIDs[providerIdx] {
		if m == config.AppConfig.Model {
			return m
		}
	}
	if len(providerModelIDs[providerIdx]) > 0 {
		return providerModelIDs[providerIdx][0]
	}
	return config.AppConfig.Model
}

func currentAPIKey() string {
	provider := detectProvider()
	return config.GetAPIKeyForProvider(provider)
}

func effortValue(label string) int {
	for i, l := range effortLabels {
		if l == label {
			return i
		}
	}
	return 2
}

func rateLimitValue(text string) int {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(text), "%d", &n); err != nil || n < 0 {
		return 0
	}
	return n
}

// effortLevel returns the numeric level for the given label (used by save).
func effortLevel(label string) int { return effortValue(label) }

// containsStr reports whether the slice contains the given string.
func containsStr(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// formatEnvList renders CLAUDE_/ANTHROPIC_/OPENAI_/GROQ_/OLLAMA_ env vars,
// masking keys like cmd/env.go list does.
func formatEnvList(env []string) string {
	prefixes := []string{"CLAUDE_", "ANTHROPIC_", "OPENAI_", "GROQ_", "OLLAMA_"}
	var rows []string
	for _, e := range env {
		for _, prefix := range prefixes {
			if strings.HasPrefix(e, prefix) {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					continue
				}
				val := parts[1]
				if strings.Contains(strings.ToLower(parts[0]), "key") && len(val) > 10 {
					val = val[:5] + "..." + val[len(val)-4:]
				}
				rows = append(rows, parts[0]+"="+val)
				break
			}
		}
	}
	sort.Strings(rows)
	if len(rows) == 0 {
		return "No provider environment variables set."
	}
	return strings.Join(rows, "\n")
}

// detectProvider resolves the provider for the currently selected model,
// falling back to base_url inspection (mirrors cmd/oauth_refresh.go).
func detectProvider() string {
	return detectProviderForModel(config.AppConfig.Model)
}

func detectProviderForModel(model string) string {
	if model != "" {
		return config.DetectProviderFromModel(model)
	}
	url := strings.ToLower(config.AppConfig.BaseURL)
	switch {
	case strings.Contains(url, "anthropic"):
		return "anthropic"
	case strings.Contains(url, "openai"):
		return "openai"
	case strings.Contains(url, "groq"):
		return "groq"
	case strings.Contains(url, "openrouter"):
		return "openrouter"
	}
	return "openai"
}

func detectProviderDescription() string {
	p := detectProvider()
	cfg := config.AppConfig
	if cfg.BaseURL != "" {
		return fmt.Sprintf("%s (%s)", p, cfg.BaseURL)
	}
	return p
}
