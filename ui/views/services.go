package views

import (
	"fmt"
	"strings"

	"github.com/joe-yeomans/ecsd/ecs"
	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ServicesView displays the list of ECS services for a cluster.
type ServicesView struct {
	services   []ecs.Service
	clusterARN string
	cursor     int
	loading    bool
	err        error
	spinner    spinner.Model
	styles     styles.Styles
	width      int
	height     int
}

// NewServicesView creates a new services view.
func NewServicesView(s styles.Styles) ServicesView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return ServicesView{
		loading: true,
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the services view.
func (v ServicesView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *ServicesView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// SetServices sets the service data.
func (v *ServicesView) SetServices(services []ecs.Service, clusterARN string) {
	v.services = services
	v.clusterARN = clusterARN
	v.loading = false
	v.err = nil
	if v.cursor >= len(services) {
		v.cursor = max(0, len(services)-1)
	}
}

// SetError sets an error state.
func (v *ServicesView) SetError(err error) {
	v.err = err
	v.loading = false
}

// SetLoading sets the loading state.
func (v *ServicesView) SetLoading(loading bool) {
	v.loading = loading
}

// SelectedService returns the currently selected service.
func (v ServicesView) SelectedService() *ecs.Service {
	if len(v.services) == 0 || v.cursor >= len(v.services) {
		return nil
	}
	return &v.services[v.cursor]
}

// ClusterARN returns the cluster ARN.
func (v ServicesView) ClusterARN() string {
	return v.clusterARN
}

// Update handles messages for the services view.
func (v ServicesView) Update(msg tea.Msg) (ServicesView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if v.loading {
			v.spinner, cmd = v.spinner.Update(msg)
		}
	}

	return v, cmd
}

// MoveUp moves the cursor up.
func (v *ServicesView) MoveUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

// MoveDown moves the cursor down.
func (v *ServicesView) MoveDown() {
	if v.cursor < len(v.services)-1 {
		v.cursor++
	}
}

// Reset resets the cursor position.
func (v *ServicesView) Reset() {
	v.cursor = 0
}

// View renders the services view.
func (v ServicesView) View() string {
	if v.loading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading services...")
	}

	if v.err != nil {
		return v.styles.Error.Render("Error: " + v.err.Error())
	}

	if len(v.services) == 0 {
		return v.styles.Muted.Render("No services found")
	}

	return v.renderTable()
}

func (v ServicesView) renderTable() string {
	// Fixed column widths for status/numeric columns
	statusWidth := 10
	runningWidth := 9
	pendingWidth := 9
	desiredWidth := 9

	// Name column takes remaining space
	fixedWidth := statusWidth + runningWidth + pendingWidth + desiredWidth
	nameWidth := v.width - fixedWidth - 4 // 4 spaces between columns
	if nameWidth < 20 {
		nameWidth = 20
	}

	colWidths := []int{nameWidth, statusWidth, runningWidth, pendingWidth, desiredWidth}
	headers := []string{"Name", "Status", "Running", "Pending", "Desired"}

	// Cell style without extra padding - just width constraint
	cellStyle := lipgloss.NewStyle()

	// Build header
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = cellStyle.Width(colWidths[i]).Render(h)
	}
	header := v.styles.TableHeader.Render(strings.Join(headerCells, " "))

	// Build rows
	var rows []string
	for i, service := range v.services {
		// Determine health status based on running vs desired
		healthIndicator := ""
		if service.RunningCount < service.DesiredCount {
			healthIndicator = v.styles.StatusStyle("UNHEALTHY").Render("!")
		}

		cells := []string{
			truncate(service.Name, colWidths[0]-1) + healthIndicator,
			v.styles.StatusStyle(service.Status).Render(service.Status),
			fmt.Sprintf("%d", service.RunningCount),
			fmt.Sprintf("%d", service.PendingCount),
			fmt.Sprintf("%d", service.DesiredCount),
		}

		rowCells := make([]string, len(cells))
		for j, cell := range cells {
			rowCells[j] = cellStyle.Width(colWidths[j]).Render(cell)
		}

		row := strings.Join(rowCells, " ")
		if i == v.cursor {
			row = v.styles.TableSelected.Render(row)
		} else {
			row = v.styles.TableRow.Render(row)
		}
		rows = append(rows, row)
	}

	return header + "\n" + strings.Join(rows, "\n")
}
