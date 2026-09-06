package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
	"github.com/user/go-claude-code/internal/mcp"
)

// TestMain isolates every test behind a temp HOME/DB so no test touches the
// developer's ~/.letsGo (config.yaml, history.db, costs.json, agents.json).
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "wails-services-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("TEST_DB_PATH", filepath.Join(dir, "test.db"))
	os.Setenv("USERPROFILE", dir)
	os.Setenv("HOME", dir)
	_ = config.LoadConfig()
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// capture collects events emitted through the hub.
type capture struct {
	mu     sync.Mutex
	events []emitted
}

type emitted struct {
	name string
	data any
}

func (c *capture) add(name string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, emitted{name: name, data: data})
}

func (c *capture) names() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []string
	for _, e := range c.events {
		out = append(out, e.name)
	}
	return out
}

func (c *capture) has(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.events {
		if e.name == name {
			return true
		}
	}
	return false
}

func (c *capture) payload(name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.events {
		if e.name == name {
			return e.data, true
		}
	}
	return nil, false
}

// payloadAt returns the i-th payload of the given event name.
func (c *capture) payloadAt(name string, i int) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var n int
	for _, e := range c.events {
		if e.name == name {
			if n == i {
				return e.data, true
			}
			n++
		}
	}
	return nil, false
}

// count returns how many times an event was emitted.
func (c *capture) count(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	var n int
	for _, e := range c.events {
		if e.name == name {
			n++
		}
	}
	return n
}

// waitFor polls the capture until pred(name) or the deadline; helper for the
// async pump (research.md D3 batches ~50ms).
func waitFor(t *testing.T, c *capture, pred func(string) bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, name := range c.names() {
			if pred(name) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for event; got %v", c.names())
}

// script is the streaming script fed to the engine via SetStreamFunc.
type script struct {
	deltas     []string
	err        error
	blockOnCh chan struct{}
}

func scriptSink(s script) engine.StreamFunc {
	return func(req api.Request,
		onDelta func(string),
		onToolUse func(api.ToolUse),
		onToolInput func(string, string),
		onUsage func(int, int)) error {
		for _, d := range s.deltas {
			onDelta(d)
		}
		if s.blockOnCh != nil {
			<-s.blockOnCh
			return context.Canceled
		}
		return s.err
	}
}

// newTestChat wires hub + real toolbarDriver + engine + chat pump; the
// capture records every hub emission.
func newTestChat(t *testing.T, s script) (*Hub, *ChatService, *capture) {
	t.Helper()
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	driver := NewToolbarDriver(hub)
	e := engine.New(driver)
	hub.SetEngine(e)
	e.SetStreamFunc(scriptSink(s))
	chat := NewChatService(hub)
	e.Start()
	chat.Start()
	t.Cleanup(func() {
		chat.Stop()
		e.Stop()
	})
	return hub, chat, c
}

// TestChatServiceStreamBatchesAndEnds — C-001 at service level: send →
// chat:start → chat:delta* (batched ≤50ms) → chat:end with the persisted
// turn; the full text survives concatenation.
func TestChatServiceStreamBatchesAndEnds(t *testing.T) {
	_, chat, c := newTestChat(t, script{deltas: []string{"hel", "lo", " w", "orld"}})
	chat.Send("hola")

	waitFor(t, c, func(name string) bool { return name == "chat:end" }, 10*time.Second)

	if !c.has("chat:start") {
		t.Fatal("expected chat:start before deltas")
	}
	var text strings.Builder
	c.mu.Lock()
	for _, e := range c.events {
		if e.name == "chat:delta" {
			td, _ := e.data.(map[string]any)
			text.WriteString(td["text"].(string))
		}
	}
	c.mu.Unlock()
	if text.String() != "hello world" {
		t.Fatalf("deltas must reassemble the text, got %q", text.String())
	}
	if c.has("chat:error") {
		t.Fatal("no chat:error expected on a healthy stream")
	}

	// The turn is persisted by the engine (frontend-contract §1 stream:end).
	hist, err := db.GetHistory(chat.SessionID())
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(hist) < 2 || hist[1].Role != "assistant" {
		t.Fatalf("expected persisted assistant message, got %d msgs", len(hist))
	}
}

// TestChatServiceCancelNoPartialPersist — C-002 at service level: cancelling
// mid-stream emits chat:cancelled (no chat:end) and nothing partial persists.
func TestChatServiceCancelNoPartialPersist(t *testing.T) {
	block := make(chan struct{})
	_, chat, c := newTestChat(t, script{deltas: []string{"partial"}, blockOnCh: block})
	sid := chat.SessionID()
	chat.Send("mensaje a cancelar")

	waitFor(t, c, func(name string) bool { return name == "chat:start" }, 5*time.Second)
	chat.Cancel()
	close(block)

	waitFor(t, c, func(name string) bool { return name == "chat:cancelled" }, 5*time.Second)
	if c.has("chat:end") {
		t.Fatal("chat:end must not fire after cancellation")
	}
	hist, err := db.GetHistory(sid)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	for _, m := range hist {
		if m.Role == "assistant" {
			t.Fatalf("partial assistant text persisted after cancel: %q", m.Content)
		}
	}
}

// TestToolbarDriverPromptApprovalFlow — tool:request surfaces the approval,
// ChatService.Approve decides it and Prompt returns the decision.
func TestToolbarDriverPromptApprovalFlow(t *testing.T) {
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	driver := NewToolbarDriver(hub)

	var allow bool
	var err error
	done := make(chan struct{})
	go func() {
		allow, err = driver.Prompt("bash", map[string]any{"command": "ls"})
		close(done)
	}()

	waitFor(t, c, func(name string) bool { return name == "tool:request" }, 3*time.Second)
	payload, _ := c.payload("tool:request")
	callID := payload.(map[string]any)["call_id"].(string)
	if callID == "" {
		t.Fatal("tool:request must carry a call_id")
	}

	chat := NewChatService(hub)
	chat.Approve(callID, true)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Prompt did not return after approve")
	}
	if err != nil || !allow {
		t.Fatalf("expected allow=true, got %v (%v)", allow, err)
	}

	// Reject path.
	rej := make(chan struct{})
	go func() {
		allow, err = driver.Prompt("edit", map[string]any{})
		close(rej)
	}()
	waitFor(t, c, func(name string) bool { return c.count("tool:request") >= 2 }, 3*time.Second)
	payload2, _ := c.payloadAt("tool:request", 1)
	chat.Approve(payload2.(map[string]any)["call_id"].(string), false)
	select {
	case <-rej:
	case <-time.After(3 * time.Second):
		t.Fatal("Prompt did not return after reject")
	}
	if err != nil || allow {
		t.Fatalf("expected allow=false, got %v (%v)", allow, err)
	}
}

// TestToolbarDriverAutoApproveHonorsConfig — driver.AutoApprove mirrors the
// persisted config categories.
func TestToolbarDriverAutoApproveHonorsConfig(t *testing.T) {
	prev := config.AppConfig.AutoApprove
	config.AppConfig.AutoApprove = map[string]bool{"bash": true, "file-edit": false}
	t.Cleanup(func() { config.AppConfig.AutoApprove = prev })
	driver := NewToolbarDriver(NewHub())
	if !driver.AutoApprove("bash") {
		t.Fatal("bash expected auto-approved from config")
	}
	if driver.AutoApprove("file-edit") {
		t.Fatal("file-edit expected to require approval")
	}
}

// TestSessionsServiceLifecycle — C-003 at service level: create/list/open/
// get-messages across the contract surface.
func TestSessionsServiceLifecycle(t *testing.T) {
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	driver := NewToolbarDriver(hub)
	e := engine.New(driver)
	e.SetStreamFunc(scriptSink(script{deltas: []string{"resp"}}))
	hub.SetEngine(e)
	chat := NewChatService(hub)
	e.Start()
	chat.Start()
	t.Cleanup(func() { chat.Stop(); e.Stop() })
	svc := NewSessionsService(hub)

	sess, err := svc.Create("s1", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("create must return the new session id")
	}
	if e.SessionID() != sess.ID {
		t.Fatalf("engine must switch to the new session, got %s", e.SessionID())
	}
	waitFor(t, c, func(name string) bool { return name == "session:list" }, 3*time.Second)

	// Chat in the created session, then read it back paginated.
	chat.Send("hola desde el servicio")
	waitFor(t, c, func(name string) bool { return name == "chat:idle" }, 10*time.Second)
	msgs, err := svc.GetMessages(sess.ID, 10, 0)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(msgs) < 2 || msgs[0].Role != "user" || msgs[1].Role != "assistant" {
		t.Fatalf("expected user+assistant in session, got %d messages", len(msgs))
	}

	// Resume from another surface: Open returns full history and emits
	// session:loaded.
	got, err := svc.Open(sess.ID)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if len(got) != len(msgs) {
		t.Fatalf("open must return the full history, got %d (want %d)", len(got), len(msgs))
	}
	waitFor(t, c, func(name string) bool { return name == "session:loaded" }, 3*time.Second)
	active, err := db.GetActiveSession()
	if err != nil || active == nil || active.ID != sess.ID {
		t.Fatalf("active session must be the resumed one (err %v)", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) < 1 {
		t.Fatal("list must contain the created session")
	}
}

// TestSettingsSaveConfigPersistsAndEmits — C-004 at service level.
func TestSettingsSaveConfigPersistsAndEmits(t *testing.T) {
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	svc := NewSettingsService(hub)

	prev := config.AppConfig.Model
	t.Cleanup(func() { config.AppConfig.Model = prev })

	cfg, err := svc.SaveConfig(map[string]any{"model": "claude-sonnet-4-test", "temperature": 0.3})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if cfg.Model != "claude-sonnet-4-test" || cfg.Temperature != 0.3 {
		t.Fatalf("config not applied: %+v", cfg)
	}
	waitFor(t, c, func(name string) bool { return name == "config:changed" }, 3*time.Second)
	if payload, ok := c.payload("config:changed"); ok {
		pcfg, _ := payload.(map[string]any)["config"].(config.Config)
		if pcfg.Model != "claude-sonnet-4-test" {
			t.Fatal("config:changed must carry the refreshed config")
		}
	}
}

// TestThemeServiceSetPersistsAndValidates.
func TestThemeServiceSetPersistsAndValidates(t *testing.T) {
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	svc := NewThemeService(hub)

	prev := config.AppConfig.ThemeVariant
	t.Cleanup(func() { config.AppConfig.ThemeVariant = prev })

	got, err := svc.Set("light")
	if err != nil || got != "light" {
		t.Fatalf("set light: %v (%v)", got, err)
	}
	if config.AppConfig.ThemeVariant != "light" {
		t.Fatal("theme must be persisted in config")
	}
	waitFor(t, c, func(name string) bool { return name == "theme:changed" }, 3*time.Second)

	if _, err := svc.Set("neon"); err == nil {
		t.Fatal("invalid variant must be rejected")
	}
	if config.AppConfig.ThemeVariant != "light" {
		t.Fatal("failed set must leave the theme untouched")
	}
}

// TestUsageServiceEmptyDay — a fresh isolated home reports zero KPIs.
func TestUsageServiceEmptyDay(t *testing.T) {
	u := NewUsageService().GetUsageToday()
	if u.CostUSD != 0 || u.Requests != 0 || u.Tokens != 0 {
		t.Fatalf("expected an empty day, got %+v", u)
	}
	if u.BudgetCap != 25.0 {
		t.Fatalf("budget cap must match the GUI reference ($25), got %v", u.BudgetCap)
	}
}

// TestAccountStatusAndSignOut.
func TestAccountStatusAndSignOut(t *testing.T) {
	prevModel := config.AppConfig.Model
	prevKey := config.AppConfig.GroqAPIKey
	t.Cleanup(func() {
		config.AppConfig.Model = prevModel
		config.AppConfig.GroqAPIKey = prevKey
	})

	config.AppConfig.Model = "llama-3.3-70b-versatile"
	config.AppConfig.GroqAPIKey = "gsk_test_key"

	svc := NewAccountService(NewHub())
	st := svc.GetStatus()
	if !st.Online {
		t.Fatal("status must be online")
	}
	if st.Busy {
		t.Fatal("no engine → must idle")
	}
	if st.Provider != "groq" || st.Model == "" {
		t.Fatalf("unexpected provider/model: %+v", st)
	}

	if err := svc.SignOut(); err != nil {
		t.Fatalf("sign out: %v", err)
	}
	if config.AppConfig.GroqAPIKey != "" || config.AppConfig.APIKey != "" {
		t.Fatal("sign out must clear every API key")
	}
}

// TestGitServiceBranchLifecycle — real git against a temp repository.
func TestGitServiceBranchLifecycle(t *testing.T) {
	repo := t.TempDir()
	for _, cmd := range []struct {
		args []string
	}{
		{[]string{"init", "-b", "main"}},
		{[]string{"config", "user.email", "test@example.com"}},
		{[]string{"config", "user.name", "Test"}},
	} {
		if _, err := runGit(repo, cmd.args...); err != nil {
			t.Fatalf("git %v: %v", cmd.args, err)
		}
	}

	svc := NewGitService(NewHub())
	svc.SetDir(repo)

	if st := svc.Status(); !strings.Contains(st, "Working tree clean") {
		t.Fatalf("fresh repo expected clean, got %q", st)
	}
	// A repo with an unborn HEAD lists no branches; seed the first commit.
	if err := os.WriteFile(filepath.Join(repo, "start.md"), []byte("# start\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Commit("start"); err != nil {
		t.Fatalf("start commit: %v", err)
	}

	if _, err := svc.CreateBranch("feature/x"); err != nil {
		t.Fatalf("create branch: %v", err)
	}
	if svc.CurrentBranch() != "feature/x" {
		t.Fatalf("current branch = %q, want feature/x", svc.CurrentBranch())
	}
	branches := svc.Branches()
	if len(branches) < 2 {
		t.Fatalf("expected main + feature/x, got %v", branches)
	}

	// Commit lifecycle: file → Status shows change → Commit clean → Undo.
	file := filepath.Join(repo, "readme.md")
	if err := os.WriteFile(file, []byte("# repo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if st := svc.Status(); strings.Contains(st, "Working tree clean") {
		t.Fatalf("status must show the new file, got %q", st)
	}
	if _, err := svc.Commit("initial"); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if st := svc.Status(); !strings.Contains(st, "Working tree clean") {
		t.Fatalf("status must be clean after commit, got %q", st)
	}
	if _, err := svc.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	tags, err := svc.Tags()
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if tags != "" {
		t.Fatalf("no tags expected in a fresh repo, got %q", tags)
	}
	if _, err := svc.Diff(); err != nil {
		t.Fatalf("diff: %v", err)
	}
	if _, err := svc.Log(); err != nil {
		t.Fatalf("log: %v", err)
	}
	if _, err := svc.Hooks(); err != nil {
		t.Fatalf("hooks: %v", err)
	}
}

// TestTasksMCPServicesListEmpty — the panes render empty states without
// touching a real registry (hermetic HOME).
func TestTasksMCPServicesListEmpty(t *testing.T) {
	if agents := NewTasksService().List(); len(agents) != 0 {
		t.Fatalf("expected no agents, got %d", len(agents))
	}
	if servers := NewMCPService().List(); len(servers) != 0 {
		t.Fatalf("expected no mcp servers, got %d", len(servers))
	}
	if plugins := NewPluginsService().List(); len(plugins) != 0 {
		t.Fatalf("expected no plugins, got %d", len(plugins))
	}
}

// TestMCPServiceAddServer — Add creates a server row and duplicates fail.
func TestMCPServiceAddServer(t *testing.T) {
	svc := NewMCPService()
	name := "svc-" + time.Now().Format("150405.000")
	if err := svc.Add(&mcp.ServerConfig{Name: name, Command: "echo", Args: []string{"hi"}}); err != nil {
		t.Fatalf("add: %v", err)
	}
	var found bool
	for _, s := range svc.List() {
		if s.Name == name {
			found = true
			if s.Command != "echo" {
				t.Fatalf("command mismatch: %q", s.Command)
			}
		}
	}
	if !found {
		t.Fatal("added server must appear in List")
	}
	if err := svc.Add(&mcp.ServerConfig{Name: name, Command: "echo"}); err == nil {
		t.Fatal("duplicate server must be rejected")
	}
}

// TestSessionsServiceProjectSynchronization tests that SetCurrentProject and Open
// synchronize the process working directory and emit project:changed events.
func TestSessionsServiceProjectSynchronization(t *testing.T) {
	_ = db.InitDB()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	tempDirA, err := os.MkdirTemp("", "project-a-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	defer os.RemoveAll(tempDirA)

	tempDirB, err := os.MkdirTemp("", "project-b-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	defer os.RemoveAll(tempDirB)

	hub := NewHub()
	cap := &capture{}
	hub.Emit = cap.add

	svc := NewSessionsService(hub)

	// Initially unscoped
	if p := svc.GetCurrentProject(); p != "" {
		t.Fatalf("expected empty initial project, got %q", p)
	}

	// SetCurrentProject to tempDirA
	if err := svc.SetCurrentProject(tempDirA); err != nil {
		t.Fatalf("SetCurrentProject A: %v", err)
	}
	if p := svc.GetCurrentProject(); !strings.EqualFold(filepath.Clean(p), filepath.Clean(tempDirA)) {
		t.Fatalf("expected project A %q, got %q", tempDirA, p)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if !strings.EqualFold(filepath.Clean(wd), filepath.Clean(tempDirA)) {
		t.Fatalf("expected os.Getwd to be %q, got %q", tempDirA, wd)
	}

	// Create session with project B
	sessB, err := svc.Create("Session B", tempDirB)
	if err != nil {
		t.Fatalf("create session B: %v", err)
	}
	wdB, _ := os.Getwd()
	if !strings.EqualFold(filepath.Clean(wdB), filepath.Clean(tempDirB)) {
		t.Fatalf("expected os.Getwd to be %q after Create, got %q", tempDirB, wdB)
	}

	// Switch back to project A explicitly
	if err := svc.SetCurrentProject(tempDirA); err != nil {
		t.Fatalf("SetCurrentProject A: %v", err)
	}

	// Open session B: must switch working directory to tempDirB
	if _, err := svc.Open(sessB.ID); err != nil {
		t.Fatalf("open session B: %v", err)
	}
	wdAfterOpen, _ := os.Getwd()
	if !strings.EqualFold(filepath.Clean(wdAfterOpen), filepath.Clean(tempDirB)) {
		t.Fatalf("expected os.Getwd to be %q after Open, got %q", tempDirB, wdAfterOpen)
	}
}

// TestGitService_StagingAndNumstat tests git diff numstat parsing, file staging/unstaging, and commit counting.
func TestGitService_StagingAndNumstat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if _, err := runGit(tempDir, "init"); err != nil {
		t.Skipf("git not available or init failed: %v", err)
	}
	_, _ = runGit(tempDir, "config", "user.name", "Test User")
	_, _ = runGit(tempDir, "config", "user.email", "test@example.com")

	// Commit 1: initial file
	f1 := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(f1, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	_, _ = runGit(tempDir, "add", "file1.txt")
	if _, err := runGit(tempDir, "commit", "-m", "initial commit"); err != nil {
		t.Skipf("git commit failed: %v", err)
	}

	svc := NewGitService(nil)
	svc.SetDir(tempDir)

	// Initially clean
	summary := svc.DiffSummary()
	if summary.CommittedCount < 1 {
		t.Errorf("expected CommittedCount >= 1, got %d", summary.CommittedCount)
	}
	if len(summary.Files) != 0 {
		t.Errorf("expected 0 changed files initially, got %d", len(summary.Files))
	}

	// Modify file1.txt and add file2.txt in a subfolder
	subDir := filepath.Join(tempDir, "sub")
	_ = os.MkdirAll(subDir, 0o755)
	f2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(f2, []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}
	if err := os.WriteFile(f1, []byte("hello\nworld\nnew line\n"), 0o644); err != nil {
		t.Fatalf("modify file1: %v", err)
	}

	// Verify DiffSummary sees unstaged changes
	summary = svc.DiffSummary()
	if len(summary.Files) == 0 {
		t.Fatal("expected changed files, got 0")
	}

	// Stage sub/file2.txt
	if _, err := svc.StageFile(filepath.Join("sub", "file2.txt")); err != nil {
		t.Fatalf("StageFile failed: %v", err)
	}

	summary = svc.DiffSummary()
	var stagedFound bool
	for _, f := range summary.Files {
		if f.Name == "file2.txt" && f.Staged {
			stagedFound = true
			if f.Additions != 3 {
				t.Errorf("expected 3 additions for file2.txt, got %d", f.Additions)
			}
		}
	}
	if !stagedFound {
		t.Errorf("expected file2.txt to be marked staged in summary")
	}

	// Unstage sub/file2.txt
	if _, err := svc.UnstageFile(filepath.Join("sub", "file2.txt")); err != nil {
		t.Fatalf("UnstageFile failed: %v", err)
	}

	// StageAll
	if _, err := svc.StageAll(); err != nil {
		t.Fatalf("StageAll failed: %v", err)
	}
	summary = svc.DiffSummary()
	for _, f := range summary.Files {
		if !f.Staged {
			t.Errorf("expected file %s to be staged after StageAll", f.Path)
		}
	}

	// UnstageAll
	if _, err := svc.UnstageAll(); err != nil {
		t.Fatalf("UnstageAll failed: %v", err)
	}
	summary = svc.DiffSummary()
	for _, f := range summary.Files {
		if f.Staged {
			t.Errorf("expected file %s to not be staged after UnstageAll", f.Path)
		}
	}
}