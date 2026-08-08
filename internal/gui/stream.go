package gui

import (
	"context"
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"github.com/user/go-claude-code/internal/engine"
)

const (
	deltaFlushInterval = 100 * time.Millisecond
	// staleAfter is the FR-020 threshold: context/rate data older than this is
	// flagged "desactualizado" (contracts/gui-contract.md, appearance).
	staleAfter = 15 * time.Minute
)

// streamPump reads engine events on a goroutine and dispatches them to the
// UI thread. Stream deltas are batched and flushed at most every 100ms so the
// UI never freezes on a burst of tokens.
type streamPump struct {
	events <-chan engine.Event
	view   *chatView
	status *statusBar
	tasks  *agentTasksView
	cancel context.CancelFunc

	pending string
	timer   *time.Timer

	// stale timer (FR-020): marks context/rate as "desactualizado" after
	// 15 minutes without a UsageUpdate.
	staleTimer *time.Timer
	lastUsage  time.Time
}

// newStreamPump starts pumping engine events into the chat view.
func newStreamPump(events <-chan engine.Event, view *chatView, status *statusBar, cancel context.CancelFunc) *streamPump {
	p := &streamPump{events: events, view: view, status: status, cancel: cancel}
	go p.run()
	return p
}

func (p *streamPump) run() {
	for ev := range p.events {
		switch e := ev.(type) {
		case engine.StreamDelta:
			if p.status != nil {
				fyne.Do(func() { p.status.MarkStreaming() })
			}
			p.bufferDelta(e.Text)
		case engine.StreamStart:
			if p.status != nil {
				fyne.Do(func() { p.status.BeginWork() })
			}
			fyne.Do(func() { p.view.StartAssistant() })
		case engine.StreamDone:
			p.flush()
			fyne.Do(func() {
				p.view.FinishAssistant()
				if p.status != nil {
					p.status.EndWork()
				}
			})
		case engine.StreamCancelled:
			p.flush()
			fyne.Do(func() {
				p.view.FinishAssistant()
				if p.status != nil {
					p.status.EndWork()
				}
			})
		case engine.UserMessageAppended:
			fyne.Do(func() { p.view.AppendUser(messageText(e.Message)) })
		case engine.ToolRequested:
			p.toolRequested(e)
			p.markTooling(e.Name)
		case engine.ToolExecuting:
			p.toolStatus(e.ToolCallID, "executing")
			p.markTooling("running tool")
		case engine.ToolResult:
			p.toolStatus(e.ToolCallID, "result:"+e.Content)
			p.markTooling("tool done")
			fyne.Do(func() {
				if b, ok := p.boundBlock(e.ToolCallID); ok {
					b.SetResult(e.Content, e.IsError)
				}
			})
		case engine.ToolRejected:
			p.toolStatus(e.ToolCallID, "rejected")
			p.markTooling("rejected")
			fyne.Do(func() {
				if b, ok := p.boundBlock(e.ToolCallID); ok {
					b.SetRejected()
				}
			})
		case engine.ToolTimedOut:
			p.toolStatus(e.ToolCallID, "timed out")
			p.markTooling("timed out")
			fyne.Do(func() {
				if b, ok := p.boundBlock(e.ToolCallID); ok {
					b.SetTimedOut()
				}
			})
		case engine.UsageUpdate:
			if p.status != nil {
				p.lastUsage = time.Now()
				p.resetStaleTimer()
				fyne.Do(func() { p.status.SetUsage(e.InputTokens, e.OutputTokens) })
			}
		case engine.AgentTaskUpdate:
			if p.tasks != nil {
				fyne.Do(func() { p.tasks.update(e.TaskID, e.Status, e.Summary) })
			}
			p.markTooling("agent " + e.Status)
		case engine.ErrorEvent:
			p.flush()
			fyne.Do(func() {
				p.view.ShowError(errors.New(e.Message))
				if p.status != nil {
					p.status.EndWork()
				}
			})
		case engine.Idle:
			fyne.Do(func() {
				p.view.SetBusy(false)
				if p.status != nil {
					p.status.EndWork()
				}
			})
		case engine.ModeChanged:
			if p.status != nil {
				fyne.Do(func() { p.status.SetMode(e.Mode) })
			}
		case engine.PlanProposed:
			fyne.Do(func() {
				card := newPlanCard(e)
				card.onApprove = func(planID string) {
					if p.view.onPlanApprove != nil {
						p.view.onPlanApprove(planID)
					}
					p.view.RemovePlanByID(planID)
				}
				card.onReject = func(planID string) {
					if p.view.onPlanReject != nil {
						p.view.onPlanReject(planID)
					}
					p.view.RemovePlanByID(planID)
				}
				card.onEdit = func(planID, instructions string) {
					if p.view.onPlanEdit != nil {
						p.view.onPlanEdit(planID, instructions)
					}
					p.view.RemovePlanByID(planID)
				}
				p.view.ShowPlan(card)
			})
		case engine.PlanDecided:
			fyne.Do(func() { p.view.RemovePlanByID(e.PlanID) })
		case engine.SteerQueued:
			fyne.Do(func() { p.view.AppendUser("Steer: " + e.Text) })
		case engine.QueueRuns:
			fyne.Do(func() { p.view.AppendQueued() })
		}
	}
}

// markTooling brings the work strip back between tool bursts (FR-006).
func (p *streamPump) markTooling(step string) {
	if p.status == nil {
		return
	}
	fyne.Do(func() { p.status.MarkTooling(step) })
}

// resetStaleTimer arms the 15-minute stale marker after fresh usage data (FR-020).
func (p *streamPump) resetStaleTimer() {
	if p.staleTimer != nil {
		p.staleTimer.Stop()
	}
	p.staleTimer = time.AfterFunc(staleAfter, func() {
		if p.status == nil {
			return
		}
		p.lastUsage = time.Time{}
		fyne.Do(func() { p.status.MarkStale(true, true) })
	})
}

// toolRequested attaches a panel to the current assistant message and shows it.
func (p *streamPump) toolRequested(e engine.ToolRequested) {
	fyne.Do(func() {
		p.view.StartToolBlock(e.ToolCallID, e.Name, formatInput(e.Input))
	})
}

// toolExecuting marks the bound block as executing.
func (p *streamPump) toolStatus(id, status string) {
	fyne.Do(func() {
		if b, ok := p.boundBlock(id); ok {
			switch status {
			case "executing":
				b.SetExecuting()
			case "requested":
				b.SetRequested("")
			}
		}
	})
}

// boundBlock resolves a tool panel by ID, falling back to the first child card.
func (p *streamPump) boundBlock(id string) (*toolPanel, bool) {
	return p.view.toolPanel(id)
}

// bufferDelta accumulates stream text and schedules a flush every ≤100ms.
func (p *streamPump) bufferDelta(delta string) {
	p.pending += delta
	if p.timer == nil {
		p.timer = time.AfterFunc(deltaFlushInterval, func() {
			p.flush()
		})
	}
}

// flush delivers pending deltas to the UI thread.
func (p *streamPump) flush() {
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	if p.pending == "" {
		return
	}
	txt := p.pending
	p.pending = ""
	fyne.Do(func() {
		p.view.AppendDelta(txt)
	})
}

// Stop halts the pump and cancels any in-flight stream.
func (p *streamPump) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}
