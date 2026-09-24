package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/joe-yeomans/ecsd/ecs"
	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EventsView displays service events in a scrollable viewport.
type EventsView struct {
	events     []ecs.Event
	clusterARN string
	serviceARN string
	viewport   viewport.Model
	loading    bool
	err        error
	spinner    spinner.Model
	styles     styles.Styles
	ready      bool
	width      int
	height     int
}

// NewEventsView creates a new events view.
func NewEventsView(s styles.Styles) EventsView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return EventsView{
		loading: true,
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the events view.
func (v EventsView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *EventsView) SetSize(width, height int) {
	v.width = width
	v.height = height

	if !v.ready {
		v.viewport = viewport.New(width, height-2)
		v.viewport.Style = v.styles.Viewport
		v.ready = true
	} else {
		v.viewport.Width = width
		v.viewport.Height = height - 2
	}

	if len(v.events) > 0 {
		v.viewport.SetContent(v.formatEvents())
	}
}

// SetEvents sets the event data.
func (v *EventsView) SetEvents(events []ecs.Event, clusterARN, serviceARN string) {
	v.events = events
	v.clusterARN = clusterARN
	v.serviceARN = serviceARN
	v.loading = false
	v.err = nil

	if v.ready {
		v.viewport.SetContent(v.formatEvents())
		v.viewport.GotoTop()
	}
}

// SetError sets an error state.
func (v *EventsView) SetError(err error) {
	v.err = err
	v.loading = false
}

// SetLoading sets the loading state.
func (v *EventsView) SetLoading(loading bool) {
	v.loading = loading
}

// ClusterARN returns the cluster ARN.
func (v EventsView) ClusterARN() string {
	return v.clusterARN
}

// ServiceARN returns the service ARN.
func (v EventsView) ServiceARN() string {
	return v.serviceARN
}

// Update handles messages for the events view.
func (v EventsView) Update(msg tea.Msg) (EventsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if v.loading {
			v.spinner, cmd = v.spinner.Update(msg)
			return v, cmd
		}
	case tea.KeyMsg:
		if v.ready && !v.loading {
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	}

	return v, nil
}

// Reset resets the viewport.
func (v *EventsView) Reset() {
	if v.ready {
		v.viewport.GotoTop()
	}
}

// View renders the events view.
func (v EventsView) View() string {
	if v.loading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading events...")
	}

	if v.err != nil {
		return v.styles.Error.Render("Error: " + v.err.Error())
	}

	if len(v.events) == 0 {
		return v.styles.Muted.Render("No events found")
	}

	if !v.ready {
		return ""
	}

	return v.viewport.View()
}

func (v EventsView) formatEvents() string {
	var lines []string

	timeStyle := lipgloss.NewStyle().
		Foreground(styles.ColorSecondary).
		Width(12)

	for _, event := range v.events {
		timeAgo := formatRelativeTime(event.CreatedAt)
		styledTime := timeStyle.Render(timeAgo)

		// Color-code based on event content
		messageStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		msg := event.Message

		if strings.Contains(msg, "has started") || strings.Contains(msg, "registered") {
			messageStyle = messageStyle.Foreground(styles.ColorSuccess)
		} else if strings.Contains(msg, "stopped") || strings.Contains(msg, "failed") || strings.Contains(msg, "unhealthy") {
			messageStyle = messageStyle.Foreground(styles.ColorDanger)
		} else if strings.Contains(msg, "draining") || strings.Contains(msg, "deregistering") {
			messageStyle = messageStyle.Foreground(styles.ColorWarning)
		}

		line := styledTime + " " + messageStyle.Render(msg)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatRelativeTime formats a time as a relative duration.
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return t.Format("Jan 2")
}
