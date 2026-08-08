package gui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/db"
)

// chatMessage is a single rendered message in the conversation.
type chatMessage struct {
	role      string // "user" | "assistant"
	text      string // markdown source (streaming accumulates)
	rich      *widget.RichText
	card      *widget.Card
	streaming bool
}

// chatView renders the conversation, streaming updates and the composer.
type chatView struct {
	container   *fyne.Container
	scroll      *container.Scroll
	messages    *fyne.Container
	input       *composerEntry
	composerBox *composerBox
	sendBtn     *widget.Button
	agentBtn    *widget.Button
	filesBtn    *widget.Button
	busy        bool

	// history holds the full session history so the view can lazily render
	// older windows on scroll-up (data-model.md scale assumptions).
	history      []db.Message
	windowStart  int  // index into history of the first rendered message
	loadingOlder bool // guards re-entrant prepend during scroll events

	current *chatMessage // assistant message being streamed
	blocks  map[string]*toolPanel

	// approval zone (FR-001/021): inline cards above the composer.
	approvalZone *fyne.Container
	pendingLabel *pendingIndicator
	queuedLabel  *widget.Label
	queuedCount  int

	onSend   func(text string)
	onSteer  func(text string)
	onQueue  func(text string)
	onCancel func()
	onAgent  func(text string)
	onSlash  func()
	onHole   func()

	// Plan callbacks (US4, FR-012/013) wired by the Controller.
	onPlanApprove func(planID string)
	onPlanReject  func(planID string)
	onPlanEdit    func(planID, instructions string)

	// Fork chooser (FR-015) mounted in the approval zone.
	fork   *forkChooser
	onFork func(sessionID string, messageID int64)
}

// newChatView builds the conversation area and composer bar.
func newChatView() *chatView {
	v := &chatView{
		blocks: make(map[string]*toolPanel),
	}
	v.messages = container.NewVBox()
	v.scroll = container.NewVScroll(v.messages)
	v.scroll.SetMinSize(fyne.NewSize(400, 300))
	v.startScrollWatch()

	v.input = &composerEntry{
		submit: func(text string) { v.submit(text) },
		cancel: func() {
			if v.onCancel != nil {
				v.onCancel()
			}
		},
		queue: func(text string) { v.queue(text) },
		busy:  func() bool { return v.busy },
		onSlash: func() {
			if v.onSlash != nil {
				v.onSlash()
			}
		},
		onHole: func() {
			if v.onHole != nil {
				v.onHole()
			}
		},
	}
	v.input.MultiLine = true
	v.input.Wrapping = fyne.TextWrapWord
	v.input.SetPlaceHolder("Escribe a LetsGO…")
	v.input.Resize(fyne.NewSize(600, 60))

	// US3 (T027): one primary button alternates Enviar <-> Detener in the
	// same place while a turn is active; the input stays enabled (FR-008).
	v.sendBtn = widget.NewButtonWithIcon("Enviar", theme.MailComposeIcon(), func() {
		if v.busy {
			if v.onCancel != nil {
				v.onCancel()
			}
			return
		}
		v.submit(v.input.Text)
	})
	v.agentBtn = widget.NewButton("Agent", func() {
		if v.onAgent != nil {
			v.onAgent(v.input.Text)
		}
	})
	v.agentBtn.Importance = widget.HighImportance

	v.filesBtn = widget.NewButtonWithIcon("Files", theme.FolderOpenIcon(), func() {
		v.showFiles()
	})

	v.composerBox = newComposerBox(v.input)
	v.input.onFocus = func(focused bool) { v.composerBox.setFocused(focused) }
	composer := container.NewBorder(nil, nil, nil, container.NewHBox(v.filesBtn, v.sendBtn, v.agentBtn), v.composerBox)
	v.approvalZone = container.NewVBox()
	v.queuedLabel = widget.NewLabel("")
	v.queuedLabel.Importance = widget.MediumImportance
	v.queuedLabel.TextStyle = fyne.TextStyle{Monospace: true}
	v.queuedLabel.Hidden = true
	v.pendingLabel = newPendingIndicator()
	// Approval queue sits between the transcript and the composer: cards render
	// inline above the input so the transcript stays visible (FR-001, SC-007).
	bottom := container.NewVBox(v.approvalZone, v.queuedLabel, v.pendingLabel, composer)
	v.container = container.NewBorder(nil, bottom, nil, nil, v.scroll)
	return v
}

// submit sends the composer text unless busy (repose). During an active turn
// the composer stays editable (FR-008) and Enter steers the running turn
// (FR-009). In repose it sends a fresh message as before.
func (v *chatView) submit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if v.busy {
		// FR-009: steer into the running turn.
		if v.onSteer != nil {
			v.input.SetText("")
			v.onSteer(text)
		}
		return
	}
	if v.onSend != nil {
		v.input.SetText("")
		v.onSend(text)
	}
}

// queue queues composer text as the next turn during an active turn (FR-010)
// and shows an inline "en cola" marker, clearing the input.
func (v *chatView) queue(text string) {
	text = strings.TrimSpace(text)
	if text == "" || !v.busy {
		return
	}
	if v.onQueue != nil {
		v.input.SetText("")
		v.onQueue(text)
	}
	v.queuedCount++
	v.queuedLabel.SetText(fmt.Sprintf("queued: %d", v.queuedCount))
	v.queuedLabel.Show()
}

// AppendQueued flags that an engine-queued message ran (decrements indicator).
func (v *chatView) AppendQueued() {
	if v.queuedCount > 0 {
		v.queuedCount--
	}
	if v.queuedCount <= 0 {
		v.queuedLabel.Hide()
	} else {
		v.queuedLabel.SetText(fmt.Sprintf("queued: %d", v.queuedCount))
	}
	v.queuedLabel.Refresh()
}

// ShowApproval mounts an inline approval card in the approval zone and
// refreshes the pending counter (FR-021).
func (v *chatView) ShowApproval(card *approvalCard) {
	v.approvalZone.Add(card)
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
	v.scrollToBottom()
	v.input.requestFocus()
}

// ShowPlan mounts an inline plan card (FR-013) above the composer.
func (v *chatView) ShowPlan(card *planCard) {
	v.approvalZone.Add(card)
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
	card.requestFocus(fyne.CurrentApp().Driver().AllWindows()[0])
	v.scrollToBottom()
}

// RemovePlanByID removes a settled plan card.
func (v *chatView) RemovePlanByID(planID string) {
	kept := v.approvalZone.Objects[:0]
	for _, o := range v.approvalZone.Objects {
		if pc, ok := o.(*planCard); ok && pc.planID == planID {
			continue
		}
		kept = append(kept, o)
	}
	v.approvalZone.Objects = kept
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
}

// DismissApproval removes a settled card and refreshes the pending counter.
func (v *chatView) DismissApproval() {
	if len(v.approvalZone.Objects) > 0 {
		v.approvalZone.Objects = v.approvalZone.Objects[:len(v.approvalZone.Objects)-1]
	}
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
}

// ShowFork mounts the fork-point chooser (FR-015) in the approval zone.
func (v *chatView) ShowFork() {
	if v.fork == nil {
		return
	}
	v.approvalZone.Add(v.fork)
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
	v.fork.requestFocus(fyne.CurrentApp().Driver().AllWindows()[0])
}

// RemoveFork unmounts the fork chooser.
func (v *chatView) RemoveFork() {
	if v.fork == nil {
		return
	}
	kept := v.approvalZone.Objects[:0]
	for _, o := range v.approvalZone.Objects {
		if o == v.fork {
			continue
		}
		kept = append(kept, o)
	}
	v.approvalZone.Objects = kept
	v.approvalZone.Refresh()
	v.pendingLabel.set(len(v.approvalZone.Objects))
	v.pendingLabel.Refresh()
}

// SetBusy alternates the primary button Enviar <-> Detener in the same place
// (US3, T027) and disables the secondary Agent command while a turn is
// active. The input stays enabled so the user can compose (FR-008).
func (v *chatView) SetBusy(busy bool) {
	v.busy = busy
	if busy {
		v.sendBtn.SetText("Detener")
		v.sendBtn.SetIcon(theme.MediaStopIcon())
		v.agentBtn.Disable()
	} else {
		v.sendBtn.SetText("Enviar")
		v.sendBtn.SetIcon(theme.MailComposeIcon())
		v.agentBtn.Enable()
	}
}

// AppendUser adds a user message to the conversation.
func (v *chatView) AppendUser(text string) {
	v.messages.Add(userCard(text))
	v.messages.Refresh()
	v.scrollToBottom()
}

// loadWindowSize is how many history messages render on resume.
const loadWindowSize = 50

// Clear empties the conversation (used on session switch).
func (v *chatView) Clear() {
	v.current = nil
	v.blocks = map[string]*toolPanel{}
	v.messages.Objects = nil
	v.messages.Refresh()
	v.history = nil
	v.windowStart = 0
	v.loadingOlder = false
}

// LoadHistory keeps the full session history and renders only the last N
// messages (data-model.md scale assumptions: windowed, not full history).
// Older messages are appended lazily on scroll-up via loadOlder.
func (v *chatView) LoadHistory(history []db.Message) {
	v.Clear()
	v.history = history
	start := 0
	if len(history) > loadWindowSize {
		start = len(history) - loadWindowSize
	}
	v.windowStart = start
	v.renderWindow()
	v.scrollToBottom()
}

// renderWindow renders history[windowStart:] as message cards.
func (v *chatView) renderWindow() {
	for _, m := range v.history[v.windowStart:] {
		v.renderMessage(m)
	}
}

// renderMessage appends one history message card without scrolling.
func (v *chatView) renderMessage(m db.Message) {
	text := messageContentText(m.Content)
	if text == "" {
		return
	}
	if m.Role == "user" {
		v.appendUserCard(text)
		return
	}
	if m.Role == "assistant" {
		v.appendAssistantCard(text)
	}
}

// appendUserCard adds a static user card (no scroll).
func (v *chatView) appendUserCard(text string) {
	v.messages.Add(withCopyButton("You", text, newMarkdownRichText(text)))
}

// appendAssistantCard adds a static assistant card (no scroll).
func (v *chatView) appendAssistantCard(text string) {
	v.messages.Add(withCopyButton("Assistant", text, newMarkdownRichText(text)))
}

// loadOlder prepends the previous window of messages when the user scrolls to
// the top, preserving the current scroll position.
func (v *chatView) loadOlder() {
	if v.loadingOlder || v.windowStart <= 0 {
		return
	}
	v.loadingOlder = true
	defer func() { v.loadingOlder = false }()

	oldHeight := v.messages.MinSize().Height
	start := v.windowStart - loadWindowSize
	if start < 0 {
		start = 0
	}

	var prepend []fyne.CanvasObject
	for _, m := range v.history[start:v.windowStart] {
		text := messageContentText(m.Content)
		if text == "" {
			continue
		}
		if m.Role == "user" {
			prepend = append(prepend, userCard(text))
		} else if m.Role == "assistant" {
			prepend = append(prepend, assistantCard(text))
		}
	}

	// Prepend keeps the newest messages at the bottom of the container.
	v.messages.Objects = append(prepend, v.messages.Objects...)
	v.windowStart = start
	v.messages.Refresh()

	// Keep the viewport anchored: content above grew, so shift the offset.
	newHeight := v.messages.MinSize().Height
	delta := newHeight - oldHeight
	if delta > 0 {
		v.scroll.Offset = fyne.NewPos(v.scroll.Offset.X, v.scroll.Offset.Y+delta)
		v.scroll.Refresh()
	}
}

// userCard builds a static user message card.
func userCard(text string) *widget.Card {
	return withCopyButton("You", text, newMarkdownRichText(text))
}

// assistantCard builds a static assistant message card.
func assistantCard(text string) *widget.Card {
	return withCopyButton("Assistant", text, newMarkdownRichText(text))
}

// withCopyButton wraps a message body with a header row containing a role
// label and a copy-to-clipboard button (cli-parity.md `copy` row). The card
// gets an accent-left bar, accent-blue for the user and a surface tint for the
// assistant (T008: differentiated bubbles).
func withCopyButton(role, text string, rich *widget.RichText) *widget.Card {
	label := widget.NewLabelWithStyle(role, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		clip := fyne.CurrentApp().Clipboard()
		if clip != nil {
			clip.SetContent(text)
		}
	})
	copyBtn.Importance = widget.LowImportance
	header := container.NewHBox(label, copyBtn)

	card := widget.NewCard("", "", container.NewVBox(header, rich))
	barColor := darkRaised
	if role != "Assistant" {
		barColor = AccentColor
	}
	bar := canvas.NewRectangle(barColor)
	body := container.NewBorder(nil, nil, &minWidth{CanvasObject: bar, w: 3}, nil,
		container.NewVBox(header, rich))
	// US3 (T024/T025): the transcript column wraps at proseWidth inside the
	// common card frame.
	card.SetContent(&proseClamp{CanvasObject: cardFrame(body), max: proseWidth})
	return card
}

// startScrollWatch triggers loadOlder when the user scrolls near the top.
func (v *chatView) startScrollWatch() {
	v.scroll.OnScrolled = func(pos fyne.Position) {
		if pos.Y <= 0 && !v.busy {
			v.loadOlder()
		}
	}
}

// StartAssistant begins a new streaming assistant message.
func (v *chatView) StartAssistant() {
	v.current = &chatMessage{
		role: "assistant",
		text: "",
		rich: widget.NewRichText(&widget.TextSegment{Style: widget.RichTextStyleInline, Text: ""}),
		card: widget.NewCard("", "", nil),
	}
	v.messages.Add(v.current.card)
	v.messages.Refresh()
}

// AppendDelta appends streamed text to the current assistant message.
func (v *chatView) AppendDelta(delta string) {
	if v.current == nil {
		v.StartAssistant()
	}
	v.current.text += delta
	v.renderCurrent()
}

// FinishAssistant marks the current assistant message as complete.
func (v *chatView) FinishAssistant() {
	if v.current == nil {
		return
	}
	v.current.streaming = false
	v.current = nil
	v.messages.Refresh()
	v.scrollToBottom()
}

// renderCurrent re-renders the in-progress assistant message.
func (v *chatView) renderCurrent() {
	if v.current == nil {
		return
	}
	segments := markdownSegments(v.current.text)
	if len(segments) == 0 {
		segments = []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleInline, Text: ""}}
	}
	v.current.rich.Segments = segments
	v.current.rich.Refresh()
	label := widget.NewLabelWithStyle("Assistant", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	var content *fyne.Container
	if len(v.blocks) > 0 {
		children := []fyne.CanvasObject{label, v.current.rich}
		for _, b := range v.blocks {
			children = append(children, b)
		}
		content = container.NewVBox(children...)
	} else {
		content = container.NewVBox(label, v.current.rich)
	}
	v.current.card.SetContent(&proseClamp{CanvasObject: cardFrame(content), max: proseWidth})
	v.current.card.Refresh()
	v.scrollToBottom()
}

// StartToolBlock creates a tool panel bound to the current assistant message.
func (v *chatView) StartToolBlock(id, name, input string) {
	if v.current == nil {
		v.StartAssistant()
	}
	panel := newToolPanel(name)
	panel.SetRequested(input)
	v.blocks[id] = panel
	v.renderCurrent()
}

// toolPanel returns the bound panel for a tool call ID.
func (v *chatView) toolPanel(id string) (*toolPanel, bool) {
	p, ok := v.blocks[id]
	return p, ok
}

// ShowError appends an error message from the engine.
func (v *chatView) ShowError(err error) {
	label := widget.NewLabelWithStyle(fmt.Sprintf("Error: %v", err), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Importance = widget.DangerImportance
	v.messages.Add(label)
	v.messages.Refresh()
	v.scrollToBottom()
}

// scrollToBottom scrolls the conversation to the newest message.
func (v *chatView) scrollToBottom() {
	v.scroll.Offset = fyne.NewPos(0, v.messages.MinSize().Height)
	v.scroll.Refresh()
}

// showFiles lists the working directory contents inline in the conversation
// (parity `files` row).
func (v *chatView) showFiles() {
	dir := "."
	entries, err := os.ReadDir(dir)
	if err != nil {
		v.ShowError(fmt.Errorf("files: %w", err))
		return
	}
	var rows []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		rows = append(rows, name)
	}
	sort.Strings(rows)
	msg := strings.Join(rows, "\n")
	if msg == "" {
		msg = "(empty directory)"
	}
	label := widget.NewLabelWithStyle("Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	rich := widget.NewRichTextFromMarkdown(msg)
	v.messages.Add(container.NewVBox(label, rich))
	v.messages.Refresh()
	v.scrollToBottom()
}
