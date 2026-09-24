package ecs

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Message types for async operations.

// ClustersLoadedMsg contains the result of loading clusters.
type ClustersLoadedMsg struct {
	Clusters []Cluster
	Err      error
}

// ServicesLoadedMsg contains the result of loading services.
type ServicesLoadedMsg struct {
	Services   []Service
	ClusterARN string
	Err        error
}

// TasksLoadedMsg contains the result of loading tasks.
type TasksLoadedMsg struct {
	Tasks      []Task
	ClusterARN string
	ServiceARN string
	Err        error
}

// EventsLoadedMsg contains the result of loading events.
type EventsLoadedMsg struct {
	Events     []Event
	ClusterARN string
	ServiceARN string
	Err        error
}

// ScheduledTasksLoadedMsg contains the result of loading scheduled tasks.
type ScheduledTasksLoadedMsg struct {
	ScheduledTasks []ScheduledTask
	ClusterARN     string
	Err            error
}

// AccountInfoMsg contains account information.
type AccountInfoMsg struct {
	Profile string
	Region  string
	Err     error
}

// Commands for async data fetching.

// LoadClustersCmd returns a command that loads ECS clusters.
func LoadClustersCmd(client *Client) tea.Cmd {
	return func() tea.Msg {
		clusters, err := client.ListClusters(context.Background())
		return ClustersLoadedMsg{Clusters: clusters, Err: err}
	}
}

// LoadServicesCmd returns a command that loads services for a cluster.
func LoadServicesCmd(client *Client, clusterARN string) tea.Cmd {
	return func() tea.Msg {
		services, err := client.ListServices(context.Background(), clusterARN)
		return ServicesLoadedMsg{Services: services, ClusterARN: clusterARN, Err: err}
	}
}

// LoadTasksCmd returns a command that loads tasks for a service.
func LoadTasksCmd(client *Client, clusterARN, serviceARN string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := client.ListTasks(context.Background(), clusterARN, serviceARN)
		return TasksLoadedMsg{Tasks: tasks, ClusterARN: clusterARN, ServiceARN: serviceARN, Err: err}
	}
}

// LoadClusterTasksCmd returns a command that loads all tasks for a cluster.
func LoadClusterTasksCmd(client *Client, clusterARN string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := client.ListClusterTasks(context.Background(), clusterARN)
		return TasksLoadedMsg{Tasks: tasks, ClusterARN: clusterARN, ServiceARN: "", Err: err}
	}
}

// LoadEventsCmd returns a command that loads events for a service.
func LoadEventsCmd(client *Client, clusterARN, serviceARN string) tea.Cmd {
	return func() tea.Msg {
		events, err := client.GetServiceEvents(context.Background(), clusterARN, serviceARN)
		return EventsLoadedMsg{Events: events, ClusterARN: clusterARN, ServiceARN: serviceARN, Err: err}
	}
}

// LoadAccountInfoCmd returns a command that loads account information.
func LoadAccountInfoCmd(client *Client) tea.Cmd {
	return func() tea.Msg {
		profile := client.Profile()
		return AccountInfoMsg{Profile: profile, Region: client.Region(), Err: nil}
	}
}

// LoadScheduledTasksCmd returns a command that loads scheduled tasks for a cluster.
func LoadScheduledTasksCmd(client *Client, clusterARN string) tea.Cmd {
	return func() tea.Msg {
		scheduledTasks, err := client.ListScheduledTasks(context.Background(), clusterARN)
		return ScheduledTasksLoadedMsg{ScheduledTasks: scheduledTasks, ClusterARN: clusterARN, Err: err}
	}
}

// TaskMetricsLoadedMsg contains the result of loading task metrics.
type TaskMetricsLoadedMsg struct {
	Metrics    *TaskMetrics
	ClusterARN string
	TaskID     string
	Err        error
}

// BatchTaskMetricsLoadedMsg contains the result of loading metrics for multiple tasks.
type BatchTaskMetricsLoadedMsg struct {
	Metrics    map[string]*TaskMetrics
	ClusterARN string
	Err        error
}

// LoadTaskMetricsCmd returns a command that loads metrics for a single task.
// taskDefFamily is the task definition family name (without revision).
func LoadTaskMetricsCmd(client *Client, clusterARN, taskID, taskDefFamily string, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		metrics, err := client.GetTaskMetrics(context.Background(), clusterARN, taskID, taskDefFamily, duration)
		return TaskMetricsLoadedMsg{Metrics: metrics, ClusterARN: clusterARN, TaskID: taskID, Err: err}
	}
}

// LoadBatchTaskMetricsCmd returns a command that loads metrics for multiple tasks.
func LoadBatchTaskMetricsCmd(client *Client, clusterARN string, tasks []TaskInfo, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		metrics, err := client.GetTasksMetricsBatch(context.Background(), clusterARN, tasks, duration)
		return BatchTaskMetricsLoadedMsg{Metrics: metrics, ClusterARN: clusterARN, Err: err}
	}
}
