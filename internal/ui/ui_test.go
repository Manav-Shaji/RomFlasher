package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelView(t *testing.T) {
	m := NewModel()
	m.Width = 100
	m.Height = 30
	// Initialize menu dimensions since View depends on them
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	view := m.View()

	if !strings.Contains(view, "VoidFlasher PRIME") {
		t.Errorf("Expected view to contain branding, got missing")
	}

	if !strings.Contains(view, "Codename:") {
		t.Errorf("Expected view to contain Device Info panel, got missing")
	}

	if !strings.Contains(view, "Live Logs:") {
		t.Errorf("Expected view to contain Log Viewer panel, got missing")
	}

	if !strings.Contains(view, "Flash Recovery") {
		t.Errorf("Expected view to contain interactive menu options, got missing")
	}
}

func TestModelLogsTrimming(t *testing.T) {
	m := NewModel()
	m.Width = 100
	m.Height = 20

	for i := 0; i < 10; i++ {
		m.Logs = append(m.Logs, "Log Line")
	}

	view := m.View()
	if len(view) == 0 {
		t.Errorf("View should render")
	}
}

func TestModelUpdateKeys(t *testing.T) {
	m := NewModel()

	tests := []string{"q", "ctrl+c", "esc"}

	for _, key := range tests {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		if key == "ctrl+c" {
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		} else if key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		}

		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Errorf("Expected tea.Quit cmd for key %s", key)
		}
	}
}

func TestModelUpdateEnterKey(t *testing.T) {
	m := NewModel()
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Since "Flash Recovery" is the first option
	if !strings.Contains(newModel.Status, "Flash Recovery") {
		t.Errorf("Expected status to change to Flash Recovery, got %s", newModel.Status)
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.Width != 120 {
		t.Errorf("Expected width 120, got %d", newModel.Width)
	}
	if newModel.Height != 40 {
		t.Errorf("Expected height 40, got %d", newModel.Height)
	}
	if newModel.Menu.Width() != 56 { // (120/2) - 4 = 56
		t.Errorf("Expected menu width 56, got %d", newModel.Menu.Width())
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()
	if m.Init() != nil {
		t.Errorf("Expected Init to return nil")
	}
}

func TestModelViewUninitialized(t *testing.T) {
	m := NewModel()
	m.Width = 0 // simulate no window size yet
	if m.View() != "Initializing..." {
		t.Errorf("Expected Initializing...")
	}
}
