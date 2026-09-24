package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Select      key.Binding
	Back        key.Binding
	Events      key.Binding
	Refresh     key.Binding
	AutoRefresh key.Binding
	Quit        key.Binding
	Help        key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Events: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "events"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		AutoRefresh: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "auto-refresh"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Refresh, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select, k.Back},
		{k.Events, k.Refresh, k.AutoRefresh, k.Quit},
	}
}
