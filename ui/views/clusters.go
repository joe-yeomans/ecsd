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

// ClustersView displays the list of ECS clusters.
type ClustersView struct {
	clusters []ecs.Cluster
	cursor   int
	loading  bool
	err      error
	spinner  spinner.Model
	styles   styles.Styles
	width    int
	height   int
}

// NewClustersView creates a new clusters view.
func NewClustersView(s styles.Styles) ClustersView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return ClustersView{
		loading: true,
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the clusters view.
func (v ClustersView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *ClustersView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// SetClusters sets the cluster data.
func (v *ClustersView) SetClusters(clusters []ecs.Cluster) {
	v.clusters = clusters
	v.loading = false
	v.err = nil
	if v.cursor >= len(clusters) {
		v.cursor = max(0, len(clusters)-1)
	}
}

// SetError sets an error state.
func (v *ClustersView) SetError(err error) {
	v.err = err
	v.loading = false
}

// SetLoading sets the loading state.
func (v *ClustersView) SetLoading(loading bool) {
	v.loading = loading
}

// SelectedCluster returns the currently selected cluster.
func (v ClustersView) SelectedCluster() *ecs.Cluster {
	if len(v.clusters) == 0 || v.cursor >= len(v.clusters) {
		return nil
	}
	return &v.clusters[v.cursor]
}

// Update handles messages for the clusters view.
func (v ClustersView) Update(msg tea.Msg) (ClustersView, tea.Cmd) {
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
func (v *ClustersView) MoveUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

// MoveDown moves the cursor down.
func (v *ClustersView) MoveDown() {
	if v.cursor < len(v.clusters)-1 {
		v.cursor++
	}
}

// View renders the clusters view.
func (v ClustersView) View() string {
	if v.loading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading clusters...")
	}

	if v.err != nil {
		return v.styles.Error.Render("Error: " + v.err.Error())
	}

	if len(v.clusters) == 0 {
		return v.styles.Muted.Render("No clusters found")
	}

	return v.renderTable()
}

func (v ClustersView) renderTable() string {
	// Fixed column widths for status/numeric columns
	statusWidth := 10
	runningWidth := 9
	pendingWidth := 9
	servicesWidth := 9

	// Name column takes remaining space
	fixedWidth := statusWidth + runningWidth + pendingWidth + servicesWidth
	nameWidth := v.width - fixedWidth - 4 // 4 spaces between columns
	if nameWidth < 20 {
		nameWidth = 20
	}

	colWidths := []int{nameWidth, statusWidth, runningWidth, pendingWidth, servicesWidth}
	headers := []string{"Name", "Status", "Running", "Pending", "Services"}

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
	for i, cluster := range v.clusters {
		cells := []string{
			truncate(cluster.Name, colWidths[0]-1),
			v.styles.StatusStyle(cluster.Status).Render(cluster.Status),
			fmt.Sprintf("%d", cluster.RunningTasksCount),
			fmt.Sprintf("%d", cluster.PendingTasksCount),
			fmt.Sprintf("%d", cluster.ActiveServicesCount),
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

// truncate truncates a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
