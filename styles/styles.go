package styles

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#3B82F6") // Blue
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorBorder    = lipgloss.Color("#374151") // Dark Gray
	ColorBg        = lipgloss.Color("#1F2937") // Dark background
	ColorText      = lipgloss.Color("#F9FAFB") // Light text
)

// Styles contains all UI styles.
type Styles struct {
	// Base styles
	App    lipgloss.Style
	Header lipgloss.Style
	Help   lipgloss.Style

	// Text styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Muted       lipgloss.Style
	Tag         lipgloss.Style
	Breadcrumb  lipgloss.Style
	BreadSep    lipgloss.Style
	ActiveCrumb lipgloss.Style

	// Status styles
	StatusActive    lipgloss.Style
	StatusInactive  lipgloss.Style
	StatusRunning   lipgloss.Style
	StatusPending   lipgloss.Style
	StatusStopped   lipgloss.Style
	StatusHealthy   lipgloss.Style
	StatusUnhealthy lipgloss.Style

	// Table styles
	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableSelected lipgloss.Style
	TableCell     lipgloss.Style

	// Other
	Error    lipgloss.Style
	Loading  lipgloss.Style
	Viewport lipgloss.Style
}

// DefaultStyles returns the default style configuration.
func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder).
			MarginBottom(1),

		Help: lipgloss.NewStyle().
			Foreground(ColorMuted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(ColorBorder).
			MarginTop(1).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary),

		Subtitle: lipgloss.NewStyle().
			Foreground(ColorSecondary),

		Muted: lipgloss.NewStyle().
			Foreground(ColorMuted),

		Tag: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSecondary).
			Padding(0, 1),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(ColorMuted),

		BreadSep: lipgloss.NewStyle().
			Foreground(ColorBorder).
			SetString(" > "),

		ActiveCrumb: lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true),

		StatusActive: lipgloss.NewStyle().
			Foreground(ColorSuccess),

		StatusInactive: lipgloss.NewStyle().
			Foreground(ColorDanger),

		StatusRunning: lipgloss.NewStyle().
			Foreground(ColorSuccess),

		StatusPending: lipgloss.NewStyle().
			Foreground(ColorWarning),

		StatusStopped: lipgloss.NewStyle().
			Foreground(ColorDanger),

		StatusHealthy: lipgloss.NewStyle().
			Foreground(ColorSuccess),

		StatusUnhealthy: lipgloss.NewStyle().
			Foreground(ColorDanger),

		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder),

		TableRow: lipgloss.NewStyle().
			Foreground(ColorText),

		TableSelected: lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorPrimary).
			Bold(true),

		TableCell: lipgloss.NewStyle().
			Padding(0, 1),

		Error: lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true),

		Loading: lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true),

		Viewport: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1),
	}
}

// StatusStyle returns the appropriate style for a status string.
func (s Styles) StatusStyle(status string) lipgloss.Style {
	switch status {
	case "ACTIVE", "RUNNING":
		return s.StatusRunning
	case "PENDING", "PROVISIONING", "ACTIVATING":
		return s.StatusPending
	case "INACTIVE", "STOPPED", "DEPROVISIONING":
		return s.StatusStopped
	case "HEALTHY":
		return s.StatusHealthy
	case "UNHEALTHY":
		return s.StatusUnhealthy
	default:
		return s.Muted
	}
}
