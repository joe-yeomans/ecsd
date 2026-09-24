package ui

import (
	"strings"

	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Help renders the help bar at the bottom of the screen.
type Help struct {
	styles   styles.Styles
	width    int
	showFull bool
}

// NewHelp creates a new help component.
func NewHelp(s styles.Styles) Help {
	return Help{styles: s}
}

// SetWidth sets the help bar width.
func (h *Help) SetWidth(width int) {
	h.width = width
}

// ToggleFull toggles between short and full help.
func (h *Help) ToggleFull() {
	h.showFull = !h.showFull
}

// Render renders the help bar with keybindings.
func (h Help) Render(bindings []key.Binding, autoRefresh bool) string {
	var parts []string

	keyStyle := lipgloss.NewStyle().
		Foreground(styles.ColorSecondary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted)

	for _, b := range bindings {
		help := b.Help()
		part := keyStyle.Render("["+help.Key+"]") + " " + descStyle.Render(help.Desc)
		parts = append(parts, part)
	}

	// Add auto-refresh indicator
	if autoRefresh {
		indicator := lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Render("● auto")
		parts = append(parts, indicator)
	}

	helpText := strings.Join(parts, "  ")
	return h.styles.Help.Width(h.width).Render(helpText)
}

// RenderWithExtra renders the help bar with additional context-specific keys.
func (h Help) RenderWithExtra(baseBindings []key.Binding, extraBindings []key.Binding, autoRefresh bool) string {
	allBindings := append(baseBindings, extraBindings...)
	return h.Render(allBindings, autoRefresh)
}
