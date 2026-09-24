package ui

import (
	"strings"
	"time"

	"github.com/joe-yeomans/ecsd/ecs"
	"github.com/joe-yeomans/ecsd/styles"
	"github.com/joe-yeomans/ecsd/ui/views"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType represents the current view.
type ViewType int

const (
	ViewClusters ViewType = iota
	ViewClusterDetail
	ViewTasks
	ViewTaskDetail
	ViewEvents
)

// RefreshTickMsg is sent on auto-refresh intervals.
type RefreshTickMsg struct{}

// Model is the root model for the TUI.
type Model struct {
	client *ecs.Client
	styles styles.Styles
	keys   KeyMap

	// Account info
	profile string
	region  string

	// Navigation stack
	viewStack []ViewType

	// Views
	clustersView      views.ClustersView
	clusterDetailView views.ClusterDetailView
	tasksView         views.TasksView
	taskDetailView    views.TaskDetailView
	eventsView        views.EventsView

	// UI components
	header     Header
	breadcrumb Breadcrumb
	help       Help

	// State
	autoRefresh bool
	width       int
	height      int
	err         error
}

// NewModel creates a new root model.
func NewModel(client *ecs.Client) Model {
	s := styles.DefaultStyles()

	return Model{
		client:            client,
		styles:            s,
		keys:              DefaultKeyMap(),
		viewStack:         []ViewType{ViewClusters},
		clustersView:      views.NewClustersView(s),
		clusterDetailView: views.NewClusterDetailView(s),
		tasksView:         views.NewTasksView(s),
		taskDetailView:    views.NewTaskDetailView(s),
		eventsView:        views.NewEventsView(s),
		header:            NewHeader(s),
		breadcrumb:        NewBreadcrumb(s),
		help:              NewHelp(s),
		autoRefresh:       true,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.clustersView.Init(),
		ecs.LoadClustersCmd(m.client),
		ecs.LoadAccountInfoCmd(m.client),
		m.scheduleRefresh(), // Start auto-refresh timer
	)
}

// CurrentView returns the current view type.
func (m Model) CurrentView() ViewType {
	if len(m.viewStack) == 0 {
		return ViewClusters
	}
	return m.viewStack[len(m.viewStack)-1]
}

// PushView adds a view to the navigation stack.
func (m *Model) PushView(view ViewType) {
	m.viewStack = append(m.viewStack, view)
}

// PopView removes the top view from the navigation stack.
func (m *Model) PopView() bool {
	if len(m.viewStack) <= 1 {
		return false
	}
	m.viewStack = m.viewStack[:len(m.viewStack)-1]
	return true
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ecs.AccountInfoMsg:
		if msg.Err == nil {
			m.profile = msg.Profile
			m.region = msg.Region
		}

	case ecs.ClustersLoadedMsg:
		if msg.Err != nil {
			m.clustersView.SetError(msg.Err)
		} else {
			m.clustersView.SetClusters(msg.Clusters)
		}

	case ecs.ServicesLoadedMsg:
		if msg.Err != nil {
			m.clusterDetailView.SetServicesError(msg.Err)
		} else {
			m.clusterDetailView.SetServices(msg.Services)
		}

	case ecs.TasksLoadedMsg:
		if msg.Err != nil {
			if m.CurrentView() == ViewClusterDetail {
				m.clusterDetailView.SetTasksError(msg.Err)
			} else {
				m.tasksView.SetError(msg.Err)
			}
		} else {
			if m.CurrentView() == ViewClusterDetail {
				m.clusterDetailView.SetTasks(msg.Tasks)
				// Load metrics for tasks
				if len(msg.Tasks) > 0 {
					taskInfos := make([]ecs.TaskInfo, len(msg.Tasks))
					for i, t := range msg.Tasks {
						taskInfos[i] = ecs.TaskInfo{
							TaskID:        t.TaskID,
							TaskDefFamily: ecs.ExtractTaskDefFamily(t.TaskDefinition),
						}
					}
					m.clusterDetailView.SetMetricsLoading(true)
					cmds = append(cmds, ecs.LoadBatchTaskMetricsCmd(m.client, msg.ClusterARN, taskInfos, 15*time.Minute))
				}
			} else {
				m.tasksView.SetTasks(msg.Tasks, msg.ClusterARN, msg.ServiceARN)
				// Load metrics for tasks
				if len(msg.Tasks) > 0 {
					taskInfos := make([]ecs.TaskInfo, len(msg.Tasks))
					for i, t := range msg.Tasks {
						taskInfos[i] = ecs.TaskInfo{
							TaskID:        t.TaskID,
							TaskDefFamily: ecs.ExtractTaskDefFamily(t.TaskDefinition),
						}
					}
					m.tasksView.SetMetricsLoading(true)
					cmds = append(cmds, ecs.LoadBatchTaskMetricsCmd(m.client, msg.ClusterARN, taskInfos, 15*time.Minute))
				}
			}
		}

	case ecs.EventsLoadedMsg:
		if msg.Err != nil {
			m.eventsView.SetError(msg.Err)
		} else {
			m.eventsView.SetEvents(msg.Events, msg.ClusterARN, msg.ServiceARN)
		}

	case ecs.ScheduledTasksLoadedMsg:
		if msg.Err != nil {
			m.clusterDetailView.SetScheduledTasksError(msg.Err)
		} else {
			m.clusterDetailView.SetScheduledTasks(msg.ScheduledTasks)
		}

	case ecs.TaskMetricsLoadedMsg:
		if m.CurrentView() == ViewTaskDetail {
			if msg.Err != nil {
				m.taskDetailView.SetMetricsError(msg.Err)
			} else {
				m.taskDetailView.SetMetrics(msg.Metrics)
			}
		}

	case ecs.BatchTaskMetricsLoadedMsg:
		// Handle batch metrics for task tables
		if msg.Err == nil && msg.Metrics != nil {
			if m.CurrentView() == ViewClusterDetail {
				m.clusterDetailView.SetMetrics(msg.Metrics)
			} else if m.CurrentView() == ViewTasks {
				m.tasksView.SetMetrics(msg.Metrics)
			}
		}

	case RefreshTickMsg:
		if m.autoRefresh {
			cmds = append(cmds, m.refreshCurrentView(), m.scheduleRefresh())
		}
	}

	// Update current view
	switch m.CurrentView() {
	case ViewClusters:
		var cmd tea.Cmd
		m.clustersView, cmd = m.clustersView.Update(msg)
		cmds = append(cmds, cmd)
	case ViewClusterDetail:
		var cmd tea.Cmd
		m.clusterDetailView, cmd = m.clusterDetailView.Update(msg)
		cmds = append(cmds, cmd)
	case ViewTasks:
		var cmd tea.Cmd
		m.tasksView, cmd = m.tasksView.Update(msg)
		cmds = append(cmds, cmd)
	case ViewTaskDetail:
		var cmd tea.Cmd
		m.taskDetailView, cmd = m.taskDetailView.Update(msg)
		cmds = append(cmds, cmd)
	case ViewEvents:
		var cmd tea.Cmd
		m.eventsView, cmd = m.eventsView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Global keys
	if key.Matches(msg, m.keys.Quit) {
		return tea.Quit
	}

	if key.Matches(msg, m.keys.AutoRefresh) {
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			return m.scheduleRefresh()
		}
		return nil
	}

	if key.Matches(msg, m.keys.Refresh) {
		return m.refreshCurrentView()
	}

	// View-specific keys
	switch m.CurrentView() {
	case ViewClusters:
		return m.handleClusterKeys(msg)
	case ViewClusterDetail:
		return m.handleClusterDetailKeys(msg)
	case ViewTasks:
		return m.handleTaskKeys(msg)
	case ViewTaskDetail:
		return m.handleTaskDetailKeys(msg)
	case ViewEvents:
		return m.handleEventKeys(msg)
	}

	return nil
}

func (m *Model) handleClusterKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.clustersView.MoveUp()
	case key.Matches(msg, m.keys.Down):
		m.clustersView.MoveDown()
	case key.Matches(msg, m.keys.Select):
		cluster := m.clustersView.SelectedCluster()
		if cluster != nil {
			m.PushView(ViewClusterDetail)
			m.clusterDetailView.Reset()
			m.clusterDetailView.SetCluster(cluster.ARN, cluster.Name)
			m.clusterDetailView.SetServicesLoading(true)
			m.clusterDetailView.SetTasksLoading(true)
			m.clusterDetailView.SetScheduledTasksLoading(true)
			// Load services, tasks, and scheduled tasks in parallel
			return tea.Batch(
				ecs.LoadServicesCmd(m.client, cluster.ARN),
				ecs.LoadClusterTasksCmd(m.client, cluster.ARN),
				ecs.LoadScheduledTasksCmd(m.client, cluster.ARN),
			)
		}
	}
	return nil
}

func (m *Model) handleClusterDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.clusterDetailView.MoveUp()
	case key.Matches(msg, m.keys.Down):
		m.clusterDetailView.MoveDown()
	case msg.String() == "tab", msg.String() == "shift+tab":
		m.clusterDetailView.NextTab()
	case msg.String() == "1":
		m.clusterDetailView.SetTab(views.TabServices)
	case msg.String() == "2":
		m.clusterDetailView.SetTab(views.TabTasks)
	case msg.String() == "3":
		m.clusterDetailView.SetTab(views.TabScheduled)
	case key.Matches(msg, m.keys.Select):
		// Services tab - drill into service tasks
		if m.clusterDetailView.ActiveTab() == views.TabServices {
			service := m.clusterDetailView.SelectedService()
			if service != nil {
				m.PushView(ViewTasks)
				m.tasksView.Reset()
				m.tasksView.SetLoading(true)
				return ecs.LoadTasksCmd(m.client, service.ClusterARN, service.ARN)
			}
		}
		// Tasks tab - drill into task detail
		if m.clusterDetailView.ActiveTab() == views.TabTasks {
			task := m.clusterDetailView.SelectedTask()
			if task != nil {
				m.PushView(ViewTaskDetail)
				m.taskDetailView.Reset()
				m.taskDetailView.SetTask(task, m.clusterDetailView.ClusterARN())
				m.taskDetailView.SetLoading(true)
				return ecs.LoadTaskMetricsCmd(m.client, m.clusterDetailView.ClusterARN(), task.TaskID, ecs.ExtractTaskDefFamily(task.TaskDefinition), time.Hour)
			}
		}
	case key.Matches(msg, m.keys.Events):
		// Events only available on Services tab
		if m.clusterDetailView.ActiveTab() == views.TabServices {
			service := m.clusterDetailView.SelectedService()
			if service != nil {
				m.PushView(ViewEvents)
				m.eventsView.Reset()
				m.eventsView.SetLoading(true)
				return ecs.LoadEventsCmd(m.client, service.ClusterARN, service.ARN)
			}
		}
	case key.Matches(msg, m.keys.Back):
		m.PopView()
	}
	return nil
}

func (m *Model) handleTaskKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.tasksView.MoveUp()
	case key.Matches(msg, m.keys.Down):
		m.tasksView.MoveDown()
	case key.Matches(msg, m.keys.Select):
		task := m.tasksView.SelectedTask()
		if task != nil {
			m.PushView(ViewTaskDetail)
			m.taskDetailView.Reset()
			m.taskDetailView.SetTask(task, m.tasksView.ClusterARN())
			m.taskDetailView.SetLoading(true)
			return ecs.LoadTaskMetricsCmd(m.client, m.tasksView.ClusterARN(), task.TaskID, ecs.ExtractTaskDefFamily(task.TaskDefinition), time.Hour)
		}
	case key.Matches(msg, m.keys.Back):
		m.PopView()
	}
	return nil
}

func (m *Model) handleTaskDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.PopView()
	}
	return nil
}

func (m *Model) handleEventKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.PopView()
	}
	return nil
}

func (m Model) refreshCurrentView() tea.Cmd {
	switch m.CurrentView() {
	case ViewClusters:
		m.clustersView.SetLoading(true)
		return ecs.LoadClustersCmd(m.client)
	case ViewClusterDetail:
		clusterARN := m.clusterDetailView.ClusterARN()
		if clusterARN != "" {
			m.clusterDetailView.SetServicesLoading(true)
			m.clusterDetailView.SetTasksLoading(true)
			m.clusterDetailView.SetScheduledTasksLoading(true)
			return tea.Batch(
				ecs.LoadServicesCmd(m.client, clusterARN),
				ecs.LoadClusterTasksCmd(m.client, clusterARN),
				ecs.LoadScheduledTasksCmd(m.client, clusterARN),
			)
		}
	case ViewTasks:
		clusterARN := m.tasksView.ClusterARN()
		serviceARN := m.tasksView.ServiceARN()
		if clusterARN != "" && serviceARN != "" {
			m.tasksView.SetLoading(true)
			return ecs.LoadTasksCmd(m.client, clusterARN, serviceARN)
		}
	case ViewTaskDetail:
		task := m.taskDetailView.Task()
		clusterARN := m.taskDetailView.ClusterARN()
		if task != nil && clusterARN != "" {
			m.taskDetailView.SetLoading(true)
			return ecs.LoadTaskMetricsCmd(m.client, clusterARN, task.TaskID, ecs.ExtractTaskDefFamily(task.TaskDefinition), time.Hour)
		}
	case ViewEvents:
		clusterARN := m.eventsView.ClusterARN()
		serviceARN := m.eventsView.ServiceARN()
		if clusterARN != "" && serviceARN != "" {
			m.eventsView.SetLoading(true)
			return ecs.LoadEventsCmd(m.client, clusterARN, serviceARN)
		}
	}
	return nil
}

func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

func (m *Model) updateSizes() {
	contentHeight := m.height - 6 // header, breadcrumb, help, margins
	contentWidth := m.width - 6   // outer padding + buffer

	m.header.SetWidth(contentWidth)
	m.help.SetWidth(contentWidth)
	m.clustersView.SetSize(contentWidth, contentHeight)
	m.clusterDetailView.SetSize(contentWidth, contentHeight)
	m.tasksView.SetSize(contentWidth, contentHeight)
	m.taskDetailView.SetSize(contentWidth, contentHeight)
	m.eventsView.SetSize(contentWidth, contentHeight)
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Header
	s.WriteString(m.header.Render(m.profile, m.region))
	s.WriteString("\n")

	// Breadcrumb
	m.breadcrumb.SetItems(m.getBreadcrumbs())
	if bc := m.breadcrumb.Render(); bc != "" {
		s.WriteString(bc)
		s.WriteString("\n\n")
	}

	// Main content
	switch m.CurrentView() {
	case ViewClusters:
		s.WriteString(m.clustersView.View())
	case ViewClusterDetail:
		s.WriteString(m.clusterDetailView.View())
	case ViewTasks:
		s.WriteString(m.tasksView.View())
	case ViewTaskDetail:
		s.WriteString(m.taskDetailView.View())
	case ViewEvents:
		s.WriteString(m.eventsView.View())
	}

	// Help bar
	s.WriteString("\n")
	s.WriteString(m.help.Render(m.getHelpBindings(), m.autoRefresh))

	return lipgloss.NewStyle().Padding(1, 2).Render(s.String())
}

func (m Model) getBreadcrumbs() []string {
	var crumbs []string

	for _, view := range m.viewStack {
		switch view {
		case ViewClusters:
			crumbs = append(crumbs, "Clusters")
		case ViewClusterDetail:
			crumbs = append(crumbs, m.clusterDetailView.ClusterName())
		case ViewTasks:
			if service := m.clusterDetailView.SelectedService(); service != nil {
				crumbs = append(crumbs, service.Name)
			}
		case ViewTaskDetail:
			if task := m.taskDetailView.Task(); task != nil {
				// Show truncated task ID
				taskID := task.TaskID
				if len(taskID) > 12 {
					taskID = taskID[:12]
				}
				crumbs = append(crumbs, taskID)
			}
		case ViewEvents:
			if service := m.clusterDetailView.SelectedService(); service != nil {
				crumbs = append(crumbs, service.Name+" (events)")
			}
		}
	}

	return crumbs
}

func (m Model) getHelpBindings() []key.Binding {
	baseBindings := []key.Binding{m.keys.Up, m.keys.Down}

	switch m.CurrentView() {
	case ViewClusters:
		return append(baseBindings, m.keys.Select, m.keys.Refresh, m.keys.AutoRefresh, m.keys.Quit)
	case ViewClusterDetail:
		if m.clusterDetailView.ActiveTab() == views.TabServices {
			return append(baseBindings, m.keys.Select, m.keys.Events, m.keys.Back, m.keys.Refresh, m.keys.Quit)
		}
		if m.clusterDetailView.ActiveTab() == views.TabTasks {
			return append(baseBindings, m.keys.Select, m.keys.Back, m.keys.Refresh, m.keys.Quit)
		}
		return append(baseBindings, m.keys.Back, m.keys.Refresh, m.keys.Quit)
	case ViewTasks:
		return append(baseBindings, m.keys.Select, m.keys.Back, m.keys.Refresh, m.keys.Quit)
	case ViewTaskDetail:
		return []key.Binding{m.keys.Back, m.keys.Refresh, m.keys.Quit}
	case ViewEvents:
		return []key.Binding{m.keys.Back, m.keys.Refresh, m.keys.Quit}
	}

	return baseBindings
}
