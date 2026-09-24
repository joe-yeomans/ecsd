package ecs

import "time"

// Cluster represents an ECS cluster with summary information.
type Cluster struct {
	ARN                  string
	Name                 string
	Status               string
	RunningTasksCount    int
	PendingTasksCount    int
	ActiveServicesCount  int
	RegisteredContainers int
}

// Service represents an ECS service.
type Service struct {
	ARN           string
	Name          string
	Status        string
	RunningCount  int
	PendingCount  int
	DesiredCount  int
	LaunchType    string
	TaskDefFamily string
	ClusterARN    string
}

// Task represents an ECS task.
type Task struct {
	ARN               string
	TaskID            string
	TaskDefinition    string // e.g., "task-a:1"
	Status            string
	HealthStatus      string
	LastStatus        string
	DesiredStatus     string
	CPU               string
	Memory            string
	StartedAt         time.Time
	StoppedAt         time.Time
	StoppedReason     string
	LaunchType        string
	ContainerInstance string
	Containers        []Container
}

// Container represents a container within a task.
type Container struct {
	Name         string
	Status       string
	HealthStatus string
	ExitCode     *int32
	Reason       string
}

// Event represents a service event.
type Event struct {
	ID        string
	Message   string
	CreatedAt time.Time
}

// ScheduledTask represents an EventBridge rule that triggers an ECS task.
type ScheduledTask struct {
	RuleName       string
	ScheduleExpr   string // e.g., "cron(0 12 * * ? *)" or "rate(5 minutes)"
	State          string // ENABLED or DISABLED
	TaskDefinition string // e.g., "task-a:1"
	TaskCount      int
	Description    string
}

// TaskMetric represents CloudWatch metrics for a task at a point in time.
type TaskMetric struct {
	Timestamp         time.Time
	CPUUtilization    float64
	MemoryUtilization float64
	NetworkBytesIn    float64
	NetworkBytesOut   float64
}

// TaskMetrics holds all metrics data for a task.
type TaskMetrics struct {
	TaskID     string
	ClusterARN string
	Metrics    []TaskMetric
}
