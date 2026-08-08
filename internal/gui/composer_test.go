package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

// TestComposerSwapSendStop verifies US3 (T023, T027): the same primary button
// alternates Enviar <-> Detener while a turn is active, tapping it cancels
// while busy and submits in repose, and the placeholder is Spanish.
func TestComposerSwapSendStop(t *testing.T) {
	app := newTestApp(t)
	defer app.Quit()

	view := newChatView()
	win := test.NewWindow(view.container)
	defer win.Close()

	sent, stopped := "", false
	view.onSend = func(s string) { sent = s }
	view.onCancel = func() { stopped = true }

	if view.sendBtn.Text != "Enviar" {
		t.Fatalf("repose label = %q, want %q", view.sendBtn.Text, "Enviar")
	}
	if view.input.PlaceHolder != "Escribe a LetsGO…" {
		t.Fatalf("placeholder = %q, want Spanish copy", view.input.PlaceHolder)
	}

	view.SetBusy(true)
	if view.busy != true {
		t.Fatal("view must flip to busy")
	}
	if view.sendBtn.Text != "Detener" {
		t.Fatalf("busy label = %q, want %q", view.sendBtn.Text, "Detener")
	}

	test.Tap(view.sendBtn)
	if !stopped {
		t.Fatal("tapping the busy button must cancel the stream")
	}

	view.SetBusy(false)
	if view.sendBtn.Text != "Enviar" {
		t.Fatalf("idle label = %q, want %q", view.sendBtn.Text, "Enviar")
	}
	view.input.SetText("hola LetsGO")
	test.Tap(view.sendBtn)
	if sent != "hola LetsGO" {
		t.Fatalf("idle tap must submit the composer text, got %q", sent)
	}
}

// TestComposerRingTwoPx verifies the US3 focus ring: 2px accent stroke when
// focused, border token when idle (T027).
func TestComposerRingTwoPx(t *testing.T) {
	app := newTestApp(t)
	defer app.Quit()

	view := newChatView()
	win := test.NewWindow(view.container)
	defer win.Close()

	if view.composerBox == nil {
		t.Fatal("chatView must own a composerBox")
	}
	if view.composerBox.ring.StrokeWidth != 2 {
		t.Fatalf("ring stroke = %v, want 2px (US3)", view.composerBox.ring.StrokeWidth)
	}

	view.composerBox.setFocused(true)
	if view.composerBox.ring.StrokeColor != composerFocusColor() {
		t.Fatal("focused ring must use the theme focus token")
	}
	view.composerBox.setFocused(false)
	if view.composerBox.ring.StrokeColor == composerFocusColor() {
		t.Fatal("idle ring must drop the focus color")
	}
}