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
)

// TasksView displays the list of ECS tasks for a service.
type TasksView struct {
	tasks          []ecs.Task
	clusterARN     string
	serviceARN     string
	metrics        map[string]*ecs.TaskMetrics // taskID -> metrics
	metricsLoading bool
	cursor         int
	loading        bool
	err            error
	spinner        spinner.Model
	styles         styles.Styles
	width          int
	height         int
}

// NewTasksView creates a new tasks view.
func NewTasksView(s styles.Styles) TasksView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return TasksView{
		loading: true,
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the tasks view.
func (v TasksView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *TasksView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// SetTasks sets the task data.
func (v *TasksView) SetTasks(tasks []ecs.Task, clusterARN, serviceARN string) {
	v.tasks = tasks
	v.clusterARN = clusterARN
	v.serviceARN = serviceARN
	v.loading = false
	v.err = nil
	if v.cursor >= len(tasks) {
		v.cursor = max(0, len(tasks)-1)
	}
}

// SetError sets an error state.
func (v *TasksView) SetError(err error) {
	v.err = err
	v.loading = false
}

// SetLoading sets the loading state.
func (v *TasksView) SetLoading(loading bool) {
	v.loading = loading
}

// SetMetrics sets the metrics data for tasks.
func (v *TasksView) SetMetrics(metrics map[string]*ecs.TaskMetrics) {
	v.metrics = metrics
	v.metricsLoading = false
}

// SetMetricsLoading sets the metrics loading state.
func (v *TasksView) SetMetricsLoading(loading bool) {
	v.metricsLoading = loading
}

// TaskIDs returns the IDs of all tasks.
func (v TasksView) TaskIDs() []string {
	ids := make([]string, len(v.tasks))
	for i, t := range v.tasks {
		ids[i] = t.TaskID
	}
	return ids
}

// SelectedTask returns the currently selected task.
func (v TasksView) SelectedTask() *ecs.Task {
	if len(v.tasks) == 0 || v.cursor >= len(v.tasks) {
		return nil
	}
	return &v.tasks[v.cursor]
}

// ClusterARN returns the cluster ARN.
func (v TasksView) ClusterARN() string {
	return v.clusterARN
}

// ServiceARN returns the service ARN.
func (v TasksView) ServiceARN() string {
	return v.serviceARN
}

// Update handles messages for the tasks view.
func (v TasksView) Update(msg tea.Msg) (TasksView, tea.Cmd) {
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
func (v *TasksView) MoveUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

// MoveDown moves the cursor down.
func (v *TasksView) MoveDown() {
	if v.cursor < len(v.tasks)-1 {
		v.cursor++
	}
}

// Reset resets the view state.
func (v *TasksView) Reset() {
	v.cursor = 0
	v.metrics = nil
	v.metricsLoading = false
}

// View renders the tasks view.
func (v TasksView) View() string {
	if v.loading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading tasks...")
	}

	if v.err != nil {
		return v.styles.Error.Render("Error: " + v.err.Error())
	}

	if len(v.tasks) == 0 {
		return v.styles.Muted.Render("No tasks found")
	}

	return v.renderTable()
}

func (v TasksView) renderTable() string {
	// Fixed column widths for status/numeric columns
	taskDefWidth := 20
	statusWidth := 12
	healthWidth := 8
	startedWidth := 8
	sparkWidth := 15 // "▁▂▃▄▅▆▇█ 100%"

	// Task ID column takes remaining space
	fixedWidth := taskDefWidth + statusWidth + healthWidth + startedWidth + sparkWidth + sparkWidth
	taskIDWidth := v.width - fixedWidth - 6 // spaces between columns
	if taskIDWidth < 12 {
		taskIDWidth = 12
	}

	colWidths := []int{taskIDWidth, taskDefWidth, statusWidth, healthWidth, sparkWidth, sparkWidth, startedWidth}
	headers := []string{"Task ID", "Task Def", "Status", "Health", "CPU %", "Mem %", "Started"}

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
	for i, task := range v.tasks {
		startedAgo := ""
		if !task.StartedAt.IsZero() {
			startedAgo = formatDuration(time.Since(task.StartedAt))
		}

		healthStatus := task.HealthStatus
		if healthStatus == "" {
			healthStatus = "-"
		}

		cpuSpark, memSpark := v.getTaskSparklines(task.TaskID, sparkWidth-1)
		cells := []string{
			truncate(task.TaskID, colWidths[0]-1),
			truncate(task.TaskDefinition, colWidths[1]-1),
			v.styles.StatusStyle(task.Status).Render(truncate(task.Status, colWidths[2]-1)),
			v.styles.StatusStyle(healthStatus).Render(truncate(healthStatus, colWidths[3]-1)),
			cpuSpark,
			memSpark,
			startedAgo,
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

	// Add stopped reason if applicable
	result := header + "\n" + strings.Join(rows, "\n")

	selectedTask := v.SelectedTask()
	if selectedTask != nil && selectedTask.StoppedReason != "" {
		reasonStyle := lipgloss.NewStyle().
			Foreground(styles.ColorWarning).
			MarginTop(1).
			Padding(0, 1)
		result += "\n" + reasonStyle.Render("Stopped reason: "+selectedTask.StoppedReason)
	}

	// Show container info for selected task
	if selectedTask != nil && len(selectedTask.Containers) > 0 {
		containerInfo := v.renderContainerInfo(selectedTask.Containers)
		result += "\n\n" + containerInfo
	}

	return result
}

func (v TasksView) getTaskSparklines(taskID string, width int) (cpuSpark, memSpark string) {
	metrics, ok := v.metrics[taskID]
	if !ok || metrics == nil || len(metrics.Metrics) == 0 {
		// No metrics available - show placeholder
		placeholder := v.styles.Muted.Render("--")
		return placeholder, placeholder
	}

	// Extract CPU and memory values
	cpuValues := make([]float64, len(metrics.Metrics))
	memValues := make([]float64, len(metrics.Metrics))
	for i, m := range metrics.Metrics {
		cpuValues[i] = m.CPUUtilization
		memValues[i] = m.MemoryUtilization
	}

	currentCPU := GetLatestValue(cpuValues)
	currentMem := GetLatestValue(memValues)

	cpuSpark = SparklineWithValueStyled(cpuValues, width, currentCPU, v.styles)
	memSpark = SparklineWithValueStyled(memValues, width, currentMem, v.styles)
	return cpuSpark, memSpark
}

func (v TasksView) renderContainerInfo(containers []ecs.Container) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorSecondary)

	var lines []string
	lines = append(lines, titleStyle.Render("Containers:"))

	for _, c := range containers {
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
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
