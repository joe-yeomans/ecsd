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

// Tab represents a tab in the cluster detail view.
type Tab int

const (
	TabServices Tab = iota
	TabTasks
	TabScheduled
)

// ClusterDetailView displays services and tasks in a tabbed interface.
type ClusterDetailView struct {
	clusterARN  string
	clusterName string
	activeTab   Tab

	// Services tab data
	services        []ecs.Service
	serviceCursor   int
	servicesLoading bool
	servicesErr     error

	// Tasks tab data
	tasks          []ecs.Task
	taskCursor     int
	tasksLoading   bool
	tasksErr       error
	metrics        map[string]*ecs.TaskMetrics // taskID -> metrics
	metricsLoading bool

	// Scheduled tasks tab data
	scheduledTasks        []ecs.ScheduledTask
	scheduledCursor       int
	scheduledTasksLoading bool
	scheduledTasksErr     error

	spinner spinner.Model
	styles  styles.Styles
	width   int
	height  int
}

// NewClusterDetailView creates a new cluster detail view.
func NewClusterDetailView(s styles.Styles) ClusterDetailView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return ClusterDetailView{
		spinner: sp,
		styles:  s,
	}
}

// Init initializes the view.
func (v ClusterDetailView) Init() tea.Cmd {
	return v.spinner.Tick
}

// SetSize sets the view dimensions.
func (v *ClusterDetailView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// SetCluster sets the cluster context.
func (v *ClusterDetailView) SetCluster(arn, name string) {
	v.clusterARN = arn
	v.clusterName = name
}

// SetServices sets the services data.
func (v *ClusterDetailView) SetServices(services []ecs.Service) {
	v.services = services
	v.servicesLoading = false
	v.servicesErr = nil
	if v.serviceCursor >= len(services) {
		v.serviceCursor = max(0, len(services)-1)
	}
}

// SetServicesError sets a services error state.
func (v *ClusterDetailView) SetServicesError(err error) {
	v.servicesErr = err
	v.servicesLoading = false
}

// SetServicesLoading sets the services loading state.
func (v *ClusterDetailView) SetServicesLoading(loading bool) {
	v.servicesLoading = loading
}

// SetTasks sets the tasks data.
func (v *ClusterDetailView) SetTasks(tasks []ecs.Task) {
	v.tasks = tasks
	v.tasksLoading = false
	v.tasksErr = nil
	if v.taskCursor >= len(tasks) {
		v.taskCursor = max(0, len(tasks)-1)
	}
}

// SetTasksError sets a tasks error state.
func (v *ClusterDetailView) SetTasksError(err error) {
	v.tasksErr = err
	v.tasksLoading = false
}

// SetTasksLoading sets the tasks loading state.
func (v *ClusterDetailView) SetTasksLoading(loading bool) {
	v.tasksLoading = loading
}

// SetMetrics sets the metrics data for tasks.
func (v *ClusterDetailView) SetMetrics(metrics map[string]*ecs.TaskMetrics) {
	v.metrics = metrics
	v.metricsLoading = false
}

// SetMetricsLoading sets the metrics loading state.
func (v *ClusterDetailView) SetMetricsLoading(loading bool) {
	v.metricsLoading = loading
}

// SetScheduledTasks sets the scheduled tasks data.
func (v *ClusterDetailView) SetScheduledTasks(tasks []ecs.ScheduledTask) {
	v.scheduledTasks = tasks
	v.scheduledTasksLoading = false
	v.scheduledTasksErr = nil
	if v.scheduledCursor >= len(tasks) {
		v.scheduledCursor = max(0, len(tasks)-1)
	}
}

// SetScheduledTasksError sets a scheduled tasks error state.
func (v *ClusterDetailView) SetScheduledTasksError(err error) {
	v.scheduledTasksErr = err
	v.scheduledTasksLoading = false
}

// SetScheduledTasksLoading sets the scheduled tasks loading state.
func (v *ClusterDetailView) SetScheduledTasksLoading(loading bool) {
	v.scheduledTasksLoading = loading
}

// ClusterARN returns the cluster ARN.
func (v ClusterDetailView) ClusterARN() string {
	return v.clusterARN
}

// ClusterName returns the cluster name.
func (v ClusterDetailView) ClusterName() string {
	return v.clusterName
}

// ActiveTab returns the active tab.
func (v ClusterDetailView) ActiveTab() Tab {
	return v.activeTab
}

// SelectedService returns the currently selected service.
func (v ClusterDetailView) SelectedService() *ecs.Service {
	if v.activeTab != TabServices || len(v.services) == 0 || v.serviceCursor >= len(v.services) {
		return nil
	}
	return &v.services[v.serviceCursor]
}

// SelectedTask returns the currently selected task.
func (v ClusterDetailView) SelectedTask() *ecs.Task {
	if v.activeTab != TabTasks || len(v.tasks) == 0 || v.taskCursor >= len(v.tasks) {
		return nil
	}
	return &v.tasks[v.taskCursor]
}

// Update handles messages.
func (v ClusterDetailView) Update(msg tea.Msg) (ClusterDetailView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if v.servicesLoading || v.tasksLoading || v.scheduledTasksLoading {
			v.spinner, cmd = v.spinner.Update(msg)
		}
	}

	return v, cmd
}

// NextTab switches to the next tab.
func (v *ClusterDetailView) NextTab() {
	switch v.activeTab {
	case TabServices:
		v.activeTab = TabTasks
	case TabTasks:
		v.activeTab = TabScheduled
	case TabScheduled:
		v.activeTab = TabServices
	}
}

// PrevTab switches to the previous tab.
func (v *ClusterDetailView) PrevTab() {
	switch v.activeTab {
	case TabServices:
		v.activeTab = TabScheduled
	case TabTasks:
		v.activeTab = TabServices
	case TabScheduled:
		v.activeTab = TabTasks
	}
}

// SetTab sets the active tab.
func (v *ClusterDetailView) SetTab(tab Tab) {
	v.activeTab = tab
}

// MoveUp moves the cursor up in the active tab.
func (v *ClusterDetailView) MoveUp() {
	switch v.activeTab {
	case TabServices:
		if v.serviceCursor > 0 {
			v.serviceCursor--
		}
	case TabTasks:
		if v.taskCursor > 0 {
			v.taskCursor--
		}
	case TabScheduled:
		if v.scheduledCursor > 0 {
			v.scheduledCursor--
		}
	}
}

// MoveDown moves the cursor down in the active tab.
func (v *ClusterDetailView) MoveDown() {
	switch v.activeTab {
	case TabServices:
		if v.serviceCursor < len(v.services)-1 {
			v.serviceCursor++
		}
	case TabTasks:
		if v.taskCursor < len(v.tasks)-1 {
			v.taskCursor++
		}
	case TabScheduled:
		if v.scheduledCursor < len(v.scheduledTasks)-1 {
			v.scheduledCursor++
		}
	}
}

// Reset resets the view state.
func (v *ClusterDetailView) Reset() {
	v.activeTab = TabServices
	v.serviceCursor = 0
	v.taskCursor = 0
	v.scheduledCursor = 0
	v.services = nil
	v.tasks = nil
	v.scheduledTasks = nil
	v.servicesErr = nil
	v.tasksErr = nil
	v.scheduledTasksErr = nil
	v.metrics = nil
	v.metricsLoading = false
}

// View renders the cluster detail view.
func (v ClusterDetailView) View() string {
	var b strings.Builder

	// Render tabs
	b.WriteString(v.renderTabs())
	b.WriteString("\n\n")

	// Render active tab content
	switch v.activeTab {
	case TabServices:
		b.WriteString(v.renderServicesTab())
	case TabTasks:
		b.WriteString(v.renderTasksTab())
	case TabScheduled:
		b.WriteString(v.renderScheduledTab())
	}

	return b.String()
}

func (v ClusterDetailView) renderTabs() string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Background(styles.ColorPrimary).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Padding(0, 2)

	var servicesTab, tasksTab, scheduledTab string
	switch v.activeTab {
	case TabServices:
		servicesTab = activeStyle.Render("1:Services")
		tasksTab = inactiveStyle.Render("2:Tasks")
		scheduledTab = inactiveStyle.Render("3:Scheduled")
	case TabTasks:
		servicesTab = inactiveStyle.Render("1:Services")
		tasksTab = activeStyle.Render("2:Tasks")
		scheduledTab = inactiveStyle.Render("3:Scheduled")
	case TabScheduled:
		servicesTab = inactiveStyle.Render("1:Services")
		tasksTab = inactiveStyle.Render("2:Tasks")
		scheduledTab = activeStyle.Render("3:Scheduled")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, servicesTab, " ", tasksTab, " ", scheduledTab)

	hint := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render("  [tab/1/2/3] switch")

	return tabs + hint
}

func (v ClusterDetailView) renderServicesTab() string {
	if v.servicesLoading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading services...")
	}

	if v.servicesErr != nil {
		return v.styles.Error.Render("Error: " + v.servicesErr.Error())
	}

	if len(v.services) == 0 {
		return v.styles.Muted.Render("No services found")
	}

	return v.renderServicesTable()
}

func (v ClusterDetailView) renderServicesTable() string {
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
		if i == v.serviceCursor {
			row = v.styles.TableSelected.Render(row)
		} else {
			row = v.styles.TableRow.Render(row)
		}
		rows = append(rows, row)
	}

	return header + "\n" + strings.Join(rows, "\n")
}

func (v ClusterDetailView) renderTasksTab() string {
	if v.tasksLoading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading tasks...")
	}

	if v.tasksErr != nil {
		return v.styles.Error.Render("Error: " + v.tasksErr.Error())
	}

	if len(v.tasks) == 0 {
		return v.styles.Muted.Render("No tasks found")
	}

	return v.renderTasksTable()
}

func (v ClusterDetailView) renderTasksTable() string {
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
		if i == v.taskCursor {
			row = v.styles.TableSelected.Render(row)
		} else {
			row = v.styles.TableRow.Render(row)
		}
		rows = append(rows, row)
	}

	return header + "\n" + strings.Join(rows, "\n")
}

func (v ClusterDetailView) getTaskSparklines(taskID string, width int) (cpuSpark, memSpark string) {
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

func (v ClusterDetailView) renderScheduledTab() string {
	if v.scheduledTasksLoading {
		return v.styles.Loading.Render(v.spinner.View() + " Loading scheduled tasks...")
	}

	if v.scheduledTasksErr != nil {
		return v.styles.Error.Render("Error: " + v.scheduledTasksErr.Error())
	}

	if len(v.scheduledTasks) == 0 {
		return v.styles.Muted.Render("No scheduled tasks found")
	}

	return v.renderScheduledTable()
}

func (v ClusterDetailView) renderScheduledTable() string {
	// Fixed column widths
	taskDefWidth := 25
	scheduleWidth := 25
	stateWidth := 10
	countWidth := 6

	// Rule name column takes remaining space
	fixedWidth := taskDefWidth + scheduleWidth + stateWidth + countWidth
	nameWidth := v.width - fixedWidth - 4 // 4 spaces between columns
	if nameWidth < 20 {
		nameWidth = 20
	}

	colWidths := []int{nameWidth, taskDefWidth, scheduleWidth, stateWidth, countWidth}
	headers := []string{"Rule Name", "Task Def", "Schedule", "State", "Count"}

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
	for i, task := range v.scheduledTasks {
		cells := []string{
			truncate(task.RuleName, colWidths[0]-1),
			truncate(task.TaskDefinition, colWidths[1]-1),
			truncate(task.ScheduleExpr, colWidths[2]-1),
			v.styles.StatusStyle(task.State).Render(task.State),
			fmt.Sprintf("%d", task.TaskCount),
		}

		rowCells := make([]string, len(cells))
		for j, cell := range cells {
			rowCells[j] = cellStyle.Width(colWidths[j]).Render(cell)
		}

		row := strings.Join(rowCells, " ")
		if i == v.scheduledCursor {
			row = v.styles.TableSelected.Render(row)
		} else {
			row = v.styles.TableRow.Render(row)
		}
		rows = append(rows, row)
	}

	return header + "\n" + strings.Join(rows, "\n")
}
