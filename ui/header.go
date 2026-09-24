package ui

import (
	"strings"

	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/lipgloss"
)

// Header renders the application header.
type Header struct {
	styles styles.Styles
	width  int
}

// NewHeader creates a new header component.
func NewHeader(s styles.Styles) Header {
	return Header{styles: s}
}

// SetWidth sets the header width.
func (h *Header) SetWidth(width int) {
	h.width = width
}

// Render renders the header with title and tags.
func (h Header) Render(profile, region string) string {
	title := h.styles.Title.Render("ECS Dashboard")

	var tags string
	if profile != "" || region != "" {
		var parts []string
		if profile != "" {
			parts = append(parts, h.styles.Tag.Render(profile))
		}
		if region != "" {
			regionTag := lipgloss.NewStyle().
				Foreground(styles.ColorText).
				Background(styles.ColorMuted).
				Padding(0, 1).
				Render(region)
			parts = append(parts, regionTag)
		}
		tags = strings.Join(parts, " ")
	}

	// Calculate spacing between title and tags
	titleWidth := lipgloss.Width(title)
	tagsWidth := lipgloss.Width(tags)
	spacing := h.width - titleWidth - tagsWidth - 4 // padding

	if spacing < 1 {
		spacing = 1
	}

	return h.styles.Header.Width(h.width).Render(
		title + strings.Repeat(" ", spacing) + tags,
	)
}

// Breadcrumb renders a breadcrumb navigation trail.
type Breadcrumb struct {
	styles styles.Styles
	items  []string
}

// NewBreadcrumb creates a new breadcrumb component.
func NewBreadcrumb(s styles.Styles) Breadcrumb {
	return Breadcrumb{styles: s}
}

// SetItems sets the breadcrumb items.
func (b *Breadcrumb) SetItems(items []string) {
	b.items = items
}

// Render renders the breadcrumb.
func (b Breadcrumb) Render() string {
	if len(b.items) == 0 {
		return ""
	}

	var parts []string
	for i, item := range b.items {
		if i == len(b.items)-1 {
			// Last item is active
			parts = append(parts, b.styles.ActiveCrumb.Render("► "+item))
		} else {
			parts = append(parts, b.styles.Breadcrumb.Render(item))
		}
	}

	return strings.Join(parts, b.styles.BreadSep.String())
}
