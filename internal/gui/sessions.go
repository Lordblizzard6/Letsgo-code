package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/db"
)

// sessionsView is the sidebar listing sessions with create/rename/delete/search
// (spec.md US3).
type sessionsView struct {
	win  fyne.Window
	list *widget.List
	srch *widget.Entry
	root *fyne.Container

	// empty aware stack (US4, T032): list + hidden empty-state overlay.
	center *emptyAware

	sessions []db.Session
	onSelect func(id string)
	onNew    func()
	onDelete func(id string)
}

// newSessionsView builds the sessions sidebar.
func newSessionsView(win fyne.Window) *sessionsView {
	s := &sessionsView{win: win}

	s.srch = widget.NewEntry()
	s.srch.SetPlaceHolder("Search messages…")
	s.srch.OnChanged = func(string) { s.reload() }

	newBtn := widget.NewButtonWithIcon("New", theme.ContentAddIcon(), func() {
		if s.onNew != nil {
			s.onNew()
		}
	})

	s.list = widget.NewList(
		func() int { return len(s.sessions) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel(""))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(s.sessions) {
				return
			}
			label := item.(*fyne.Container).Objects[1].(*widget.Label)
			label.SetText(sessionDisplayName(s.sessions[id]))
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(s.sessions) && s.onSelect != nil {
			s.onSelect(s.sessions[id].ID)
		}
	}
	s.list.OnUnselected = func(widget.ListItemID) {}

	s.list.OnUnselected = func(widget.ListItemID) {}

	// US4 (T032): empty sessions show a friendly empty state with a "New"
	// CTA instead of a bare list; the list underneath is preserved for the
	// next reload.
	s.center = newEmptyAware(s.list, "No hay conversaciones todavía.", "Nueva conversación", func() {
		if s.onNew != nil {
			s.onNew()
		}
	})

	s.root = container.NewBorder(
		s.srch,
		newBtn,
		nil, nil,
		s.center.content(),
	)
	return s
}

// reload re-fetches sessions and repaints the list.
func (s *sessionsView) reload() {
	sessions, err := db.ListSessions()
	if err != nil {
		s.sessions = nil
		s.list.Refresh()
		return
	}
	query := strings.TrimSpace(s.srch.Text)
	s.sessions = sessions
	if query != "" {
		s.sessions = filterSessionsBySearch(sessions, query)
	}
	s.list.Refresh()
	s.center.setEmpty(len(s.sessions) == 0)
}

// selectSession highlights the given session in the list.
func (s *sessionsView) selectSession(id string) {
	for i, session := range s.sessions {
		if session.ID == id {
			s.list.Select(i)
			return
		}
	}
}

// sessionIDs returns the currently listed session IDs in list order (Ctrl+1..9).
func (s *sessionsView) sessionIDs() []string {
	ids := make([]string, 0, len(s.sessions))
	for _, session := range s.sessions {
		ids = append(ids, session.ID)
	}
	return ids
}

// rename prompts for a new name and renames the session.
func (s *sessionsView) rename(sessionID string) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("New session name")
	dlg := dialog.NewForm("Rename session", "Save", "Cancel",
		[]*widget.FormItem{{Text: "Name", Widget: entry}},
		func(ok bool) {
			if !ok || strings.TrimSpace(entry.Text) == "" {
				return
			}
			_ = db.RenameSession(sessionID, strings.TrimSpace(entry.Text))
			s.reload()
		}, s.win)
	dlg.Show()
}

// delete confirms and deletes the session.
func (s *sessionsView) delete(sessionID string) {
	dlg := dialog.NewConfirm("Delete session",
		"Delete this session and its history? This cannot be undone.",
		func(ok bool) {
			if !ok {
				return
			}
			_ = db.DeleteSession(sessionID)
			if s.onDelete != nil {
				s.onDelete(sessionID)
			}
			s.reload()
		}, s.win)
	dlg.SetDismissText("Cancel")
	dlg.SetConfirmText("Delete")
	dlg.Show()
}

// rewind lets the user jump back to a point in the conversation, deleting all
// messages after it (cli-parity.md `rewind` / `teleport` rows).
func (s *sessionsView) rewind(sessionID string) {
	history, err := db.GetHistory(sessionID)
	if err != nil || len(history) == 0 {
		dialog.ShowInformation("Rewind", "No history to rewind.", s.win)
		return
	}

	var items []string
	byIndex := make(map[int]int64, len(history))
	for i, m := range history {
		items = append(items, fmt.Sprintf("%d. %s — %s", i+1, m.Role, messageContentText(m.Content)))
		byIndex[i+1] = m.ID
	}
	sel := widget.NewSelect(items, nil)
	sel.SetSelectedIndex(len(items) - 1)

	dlg := dialog.NewForm("Rewind to message", "Rewind", "Cancel",
		[]*widget.FormItem{{Text: "Message", Widget: sel}},
		func(ok bool) {
			if !ok || sel.SelectedIndex() < 0 {
				return
			}
			keep := byIndex[sel.SelectedIndex()+1]
			for _, m := range history {
				if m.ID > keep {
					_ = db.DeleteMessage(m.ID)
				}
			}
			if s.onSelect != nil {
				s.onSelect(sessionID)
			}
			s.reload()
		}, s.win)
	dlg.Show()
}

// teleport shows the full content of a message at a point in the conversation.
func (s *sessionsView) teleport(sessionID string) {
	history, err := db.GetHistory(sessionID)
	if err != nil || len(history) == 0 {
		dialog.ShowInformation("Teleport", "No history to teleport to.", s.win)
		return
	}

	var items []string
	for i, m := range history {
		items = append(items, fmt.Sprintf("%d. %s — %s", i+1, m.Role, messageContentText(m.Content)))
	}
	sel := widget.NewSelect(items, nil)
	sel.SetSelectedIndex(len(items) - 1)

	dlg := dialog.NewForm("Teleport to message", "View", "Cancel",
		[]*widget.FormItem{{Text: "Message", Widget: sel}},
		func(ok bool) {
			if !ok || sel.SelectedIndex() < 0 {
				return
			}
			m := history[sel.SelectedIndex()]
			dialog.ShowInformation(
				fmt.Sprintf("%s — %s", m.Role, m.Timestamp.Format("01-02 15:04")),
				messageContentText(m.Content), s.win)
		}, s.win)
	dlg.Show()
}

func sessionDisplayName(session db.Session) string {
	name := session.Name
	if name == "" {
		name = session.ID
	}
	ts := session.UpdatedAt.Format("01-02 15:04")
	cwd := session.ProjectPath
	if cwd == "" {
		cwd = "?"
	}
	return fmt.Sprintf("%s  (%s)  %s", name, ts, cwd)
}

func filterSessionsBySearch(sessions []db.Session, query string) []db.Session {
	var out []db.Session
	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Name), strings.ToLower(query)) {
			out = append(out, session)
			continue
		}
		results, _, err := db.SearchMessages(session.ID, query, 1, 0)
		if err == nil && len(results) > 0 {
			out = append(out, session)
		}
	}
	return out
}

func (s *sessionsView) content() fyne.CanvasObject { return s.root }
