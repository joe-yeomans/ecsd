package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/joe-yeomans/ecsd/ecs"
	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

// TaskDetailView displays detailed information about a single task including metrics graphs.
type TaskDetailView struct {
	task       *ecs.Task
	clusterARN string
	metrics    *ecs.TaskMetrics
	loading    bool
	metricsErr error
	spinner    spinner.Model
	styles     styles.Styles
	width      int
	height     int
}

// NewTaskDetailView creates a new task detail view.
func NewTaskDetailView(s styles.Styles) TaskDetailView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return TaskDetailView{
		loading: true,
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the task detail view.
func (v TaskDetailView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *TaskDetailView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// SetTask sets the task to display.
func (v *TaskDetailView) SetTask(task *ecs.Task, clusterARN string) {
	v.task = task
	v.clusterARN = clusterARN
	v.metrics = nil
	v.metricsErr = nil
}

// SetMetrics sets the metrics data.
func (v *TaskDetailView) SetMetrics(metrics *ecs.TaskMetrics) {
	v.metrics = metrics
	v.loading = false
	v.metricsErr = nil
}

// SetMetricsError sets a metrics error state.
func (v *TaskDetailView) SetMetricsError(err error) {
	v.metricsErr = err
	v.loading = false
}

// SetLoading sets the loading state.
func (v *TaskDetailView) SetLoading(loading bool) {
	v.loading = loading
}

// Task returns the current task.
func (v TaskDetailView) Task() *ecs.Task {
	return v.task
}

// ClusterARN returns the cluster ARN.
func (v TaskDetailView) ClusterARN() string {
	return v.clusterARN
}

// Reset resets the view state.
func (v *TaskDetailView) Reset() {
	v.task = nil
	v.clusterARN = ""
	v.metrics = nil
	v.metricsErr = nil
	v.loading = true
}

// Update handles messages.
func (v TaskDetailView) Update(msg tea.Msg) (TaskDetailView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if v.loading {
			v.spinner, cmd = v.spinner.Update(msg)
		}
	}

	return v, cmd
}

// View renders the task detail view.
func (v TaskDetailView) View() string {
	if v.task == nil {
		return v.styles.Muted.Render("No task selected")
	}

	var b strings.Builder

	// Task info header
	b.WriteString(v.renderTaskInfo())
	b.WriteString("\n\n")

	// Metrics graphs
	if v.loading {
		b.WriteString(v.styles.Loading.Render(v.spinner.View() + " Loading metrics..."))
	} else if v.metricsErr != nil {
		b.WriteString(v.styles.Muted.Render("Metrics unavailable: " + v.metricsErr.Error()))
		b.WriteString("\n\n")
		b.WriteString(v.renderContainers())
	} else if v.metrics != nil && len(v.metrics.Metrics) > 0 {
		b.WriteString(v.renderMetricsGraphs())
		b.WriteString("\n\n")
		b.WriteString(v.renderContainers())
	} else {
		b.WriteString(v.styles.Muted.Render("No metrics data available"))
		b.WriteString("\n\n")
		b.WriteString(v.renderContainers())
	}

	return b.String()
}

func (v TaskDetailView) renderTaskInfo() string {
	task := v.task

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary)

	labelStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted)

	valueStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText)

	var b strings.Builder

	// Title line
	b.WriteString(titleStyle.Render("Task: " + task.TaskID))
	b.WriteString("\n")

	// Status line
	status := v.styles.StatusStyle(task.Status).Render(task.Status)
	health := "-"
	if task.HealthStatus != "" && task.HealthStatus != "UNKNOWN" {
		health = v.styles.StatusStyle(task.HealthStatus).Render(task.HealthStatus)
	}
	started := "-"
	if !task.StartedAt.IsZero() {
		started = formatDuration(time.Since(task.StartedAt)) + " ago"
	}

	b.WriteString(labelStyle.Render("Status: "))
	b.WriteString(status)
	b.WriteString("    ")
	b.WriteString(labelStyle.Render("Health: "))
	b.WriteString(health)
	b.WriteString("    ")
	b.WriteString(labelStyle.Render("Started: "))
	b.WriteString(valueStyle.Render(started))
	b.WriteString("\n")

	// Task definition
	b.WriteString(labelStyle.Render("Task Definition: "))
	b.WriteString(valueStyle.Render(task.TaskDefinition))

	// Resources
	if task.CPU != "" || task.Memory != "" {
		b.WriteString("    ")
		b.WriteString(labelStyle.Render("CPU: "))
		b.WriteString(valueStyle.Render(task.CPU))
		b.WriteString("    ")
		b.WriteString(labelStyle.Render("Memory: "))
		b.WriteString(valueStyle.Render(task.Memory))
	}

	return b.String()
}

func (v TaskDetailView) renderMetricsGraphs() string {
	if v.metrics == nil || len(v.metrics.Metrics) == 0 {
		return ""
	}

	// Extract CPU and memory values
	cpuValues := make([]float64, len(v.metrics.Metrics))
	memValues := make([]float64, len(v.metrics.Metrics))

	for i, m := range v.metrics.Metrics {
		cpuValues[i] = m.CPUUtilization
		memValues[i] = m.MemoryUtilization
	}

	// Get current values (latest)
	currentCPU := GetLatestValue(cpuValues)
	currentMem := GetLatestValue(memValues)

	graphWidth := v.width - 4
	if graphWidth < 40 {
		graphWidth = 40
	}
	graphHeight := 8

	var b strings.Builder

	// CPU Graph
	b.WriteString(v.renderGraph("CPU Utilization (last 1h)", cpuValues, currentCPU, graphWidth, graphHeight))
	b.WriteString("\n\n")

	// Memory Graph
	b.WriteString(v.renderGraph("Memory Utilization (last 1h)", memValues, currentMem, graphWidth, graphHeight))

	return b.String()
}

func (v TaskDetailView) renderGraph(title string, values []float64, current float64, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorSecondary)

	currentStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText)

	// Title with current value
	header := titleStyle.Render(title)
	currentStr := fmt.Sprintf("Current: %.0f%%", current)
	padding := width - lipgloss.Width(title) - len(currentStr)
	if padding < 1 {
		padding = 1
	}
	header += strings.Repeat(" ", padding) + currentStyle.Render(currentStr)

	// Render graph using asciigraph
	graphOpts := []asciigraph.Option{
		asciigraph.Height(height),
		asciigraph.Width(width - 10), // Leave room for Y-axis labels
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(100),
	}

	// Color the graph based on current value
	switch {
	case current >= 90:
		graphOpts = append(graphOpts, asciigraph.SeriesColors(asciigraph.Red))
	case current >= 70:
		graphOpts = append(graphOpts, asciigraph.SeriesColors(asciigraph.Yellow))
	default:
		graphOpts = append(graphOpts, asciigraph.SeriesColors(asciigraph.Green))
	}

	graph := asciigraph.Plot(values, graphOpts...)

	return header + "\n" + graph
}

func (v TaskDetailView) renderContainers() string {
	if v.task == nil || len(v.task.Containers) == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorSecondary)

	var b strings.Builder
	b.WriteString(titleStyle.Render("Containers:"))
	b.WriteString("\n")

	for _, c := range v.task.Containers {
		status := v.styles.StatusStyle(c.Status).Render(c.Status)
		health := ""
		if c.HealthStatus != "" && c.HealthStatus != "UNKNOWN" {
			health = " [" + v.styles.StatusStyle(c.HealthStatus).Render(c.HealthStatus) + "]"
		}
		exitInfo := ""
		if c.ExitCode != nil {
			exitInfo = fmt.Sprintf(" (exit: %d)", int(*c.ExitCode))
		}
		reason := ""
		if c.Reason != "" {
			reason = " - " + c.Reason
		}

		line := fmt.Sprintf("  • %s: %s%s%s%s", c.Name, status, health, exitInfo, reason)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}
