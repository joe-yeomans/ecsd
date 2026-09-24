package ecs

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client wraps the ECS API client.
type Client struct {
	ecs         *ecs.Client
	sts         *sts.Client
	eventbridge *eventbridge.Client
	cloudwatch  *cloudwatch.Client
	region      string
	profile     string
}

// NewClient creates a new ECS client wrapper.
func NewClient(cfg aws.Config, profile string) *Client {
	return &Client{
		ecs:         ecs.NewFromConfig(cfg),
		sts:         sts.NewFromConfig(cfg),
		eventbridge: eventbridge.NewFromConfig(cfg),
		cloudwatch:  cloudwatch.NewFromConfig(cfg),
		region:      cfg.Region,
		profile:     profile,
	}
}

// Region returns the AWS region.
func (c *Client) Region() string {
	return c.region
}

// Profile returns the AWS profile name.
func (c *Client) Profile() string {
	return c.profile
}

// GetAccountAlias returns a stage name based on the AWS account.
func (c *Client) GetAccountAlias(ctx context.Context) (string, error) {
	identity, err := c.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	// Map account IDs to stage names (simplified - you may want to enhance this)
	account := aws.ToString(identity.Account)
	switch {
	case strings.HasSuffix(account, "dev"):
		return "dev", nil
	default:
		// Return last 4 digits of account as identifier
		if len(account) >= 4 {
			return account[len(account)-4:], nil
		}
		return account, nil
	}
}

// ListClusters retrieves all ECS clusters with their details.
func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	listOutput, err := c.ecs.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	if len(listOutput.ClusterArns) == 0 {
		return []Cluster{}, nil
	}

	descOutput, err := c.ecs.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: listOutput.ClusterArns,
	})
	if err != nil {
		return nil, err
	}

	clusters := make([]Cluster, 0, len(descOutput.Clusters))
	for _, c := range descOutput.Clusters {
		clusters = append(clusters, Cluster{
			ARN:                  aws.ToString(c.ClusterArn),
			Name:                 aws.ToString(c.ClusterName),
			Status:               aws.ToString(c.Status),
			RunningTasksCount:    int(c.RunningTasksCount),
			PendingTasksCount:    int(c.PendingTasksCount),
			ActiveServicesCount:  int(c.ActiveServicesCount),
			RegisteredContainers: int(c.RegisteredContainerInstancesCount),
		})
	}

	return clusters, nil
}

// ListServices retrieves all services for a cluster.
func (c *Client) ListServices(ctx context.Context, clusterARN string) ([]Service, error) {
	var serviceARNs []string
	var nextToken *string

	for {
		listOutput, err := c.ecs.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   aws.String(clusterARN),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		serviceARNs = append(serviceARNs, listOutput.ServiceArns...)

		if listOutput.NextToken == nil {
			break
		}
		nextToken = listOutput.NextToken
	}

	if len(serviceARNs) == 0 {
		return []Service{}, nil
	}

	// DescribeServices has a limit of 10 services per call
	var services []Service
	for i := 0; i < len(serviceARNs); i += 10 {
		end := i + 10
		if end > len(serviceARNs) {
			end = len(serviceARNs)
		}

		descOutput, err := c.ecs.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  aws.String(clusterARN),
			Services: serviceARNs[i:end],
		})
		if err != nil {
			return nil, err
		}

		for _, s := range descOutput.Services {
			launchType := string(s.LaunchType)
			if launchType == "" && len(s.CapacityProviderStrategy) > 0 {
				launchType = "CAPACITY_PROVIDER"
			}

			services = append(services, Service{
				ARN:           aws.ToString(s.ServiceArn),
				Name:          aws.ToString(s.ServiceName),
				Status:        aws.ToString(s.Status),
				RunningCount:  int(s.RunningCount),
				PendingCount:  int(s.PendingCount),
				DesiredCount:  int(s.DesiredCount),
				LaunchType:    launchType,
				TaskDefFamily: extractTaskDefFamily(aws.ToString(s.TaskDefinition)),
				ClusterARN:    clusterARN,
			})
		}
	}

	return services, nil
}

// ListTasks retrieves all tasks for a service.
func (c *Client) ListTasks(ctx context.Context, clusterARN, serviceARN string) ([]Task, error) {
	var taskARNs []string
	var nextToken *string

	for {
		listOutput, err := c.ecs.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:     aws.String(clusterARN),
			ServiceName: aws.String(serviceARN),
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, err
		}

		taskARNs = append(taskARNs, listOutput.TaskArns...)

		if listOutput.NextToken == nil {
			break
		}
		nextToken = listOutput.NextToken
	}

	return c.describeTasks(ctx, clusterARN, taskARNs)
}

// ListClusterTasks retrieves all tasks in a cluster (including standalone/scheduled tasks).
func (c *Client) ListClusterTasks(ctx context.Context, clusterARN string) ([]Task, error) {
	var taskARNs []string

	// Fetch running tasks
	var nextToken *string
	for {
		listOutput, err := c.ecs.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:       aws.String(clusterARN),
			DesiredStatus: ecstypes.DesiredStatusRunning,
			NextToken:     nextToken,
		})
		if err != nil {
			return nil, err
		}

		taskARNs = append(taskARNs, listOutput.TaskArns...)

		if listOutput.NextToken == nil {
			break
		}
		nextToken = listOutput.NextToken
	}

	// Fetch stopped tasks (includes recently completed scheduled tasks)
	nextToken = nil
	for {
		listOutput, err := c.ecs.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:       aws.String(clusterARN),
			DesiredStatus: ecstypes.DesiredStatusStopped,
			NextToken:     nextToken,
		})
		if err != nil {
			return nil, err
		}

		taskARNs = append(taskARNs, listOutput.TaskArns...)

		if listOutput.NextToken == nil {
			break
		}
		nextToken = listOutput.NextToken
	}

	return c.describeTasks(ctx, clusterARN, taskARNs)
}

// describeTasks converts task ARNs to Task structs.
func (c *Client) describeTasks(ctx context.Context, clusterARN string, taskARNs []string) ([]Task, error) {

	if len(taskARNs) == 0 {
		return []Task{}, nil
	}

	// DescribeTasks has a limit of 100 tasks per call
	var tasks []Task
	for i := 0; i < len(taskARNs); i += 100 {
		end := i + 100
		if end > len(taskARNs) {
			end = len(taskARNs)
		}

		descOutput, err := c.ecs.DescribeTasks(ctx, &ecs.DescribeTasksInput{
			Cluster: aws.String(clusterARN),
			Tasks:   taskARNs[i:end],
		})
		if err != nil {
			return nil, err
		}

		for _, t := range descOutput.Tasks {
			var containers []Container
			for _, cont := range t.Containers {
				containers = append(containers, Container{
					Name:         aws.ToString(cont.Name),
					Status:       aws.ToString(cont.LastStatus),
					HealthStatus: string(cont.HealthStatus),
					ExitCode:     cont.ExitCode,
					Reason:       aws.ToString(cont.Reason),
				})
			}

			var startedAt, stoppedAt = t.StartedAt, t.StoppedAt
			task := Task{
				ARN:               aws.ToString(t.TaskArn),
				TaskID:            extractTaskID(aws.ToString(t.TaskArn)),
				TaskDefinition:    extractTaskDefFamilyRevision(aws.ToString(t.TaskDefinitionArn)),
				Status:            aws.ToString(t.LastStatus),
				HealthStatus:      string(t.HealthStatus),
				LastStatus:        aws.ToString(t.LastStatus),
				DesiredStatus:     aws.ToString(t.DesiredStatus),
				CPU:               aws.ToString(t.Cpu),
				Memory:            aws.ToString(t.Memory),
				StoppedReason:     aws.ToString(t.StoppedReason),
				LaunchType:        string(t.LaunchType),
				ContainerInstance: aws.ToString(t.ContainerInstanceArn),
				Containers:        containers,
			}
			if startedAt != nil {
				task.StartedAt = *startedAt
			}
			if stoppedAt != nil {
				task.StoppedAt = *stoppedAt
			}
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// GetServiceEvents retrieves events for a service.
func (c *Client) GetServiceEvents(ctx context.Context, clusterARN, serviceARN string) ([]Event, error) {
	descOutput, err := c.ecs.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterARN),
		Services: []string{serviceARN},
	})
	if err != nil {
		return nil, err
	}

	if len(descOutput.Services) == 0 {
		return []Event{}, nil
	}

	service := descOutput.Services[0]
	events := make([]Event, 0, len(service.Events))
	for _, e := range service.Events {
		var createdAt = e.CreatedAt
		event := Event{
			ID:      aws.ToString(e.Id),
			Message: aws.ToString(e.Message),
		}
		if createdAt != nil {
			event.CreatedAt = *createdAt
		}
		events = append(events, event)
	}

	return events, nil
}

// extractTaskDefFamily extracts the family name from a task definition ARN.
func extractTaskDefFamily(taskDef string) string {
	// Format: arn:aws:ecs:region:account:task-definition/family:revision
	parts := strings.Split(taskDef, "/")
	if len(parts) < 2 {
		return taskDef
	}
	familyRev := parts[len(parts)-1]
	familyParts := strings.Split(familyRev, ":")
	return familyParts[0]
}

// extractTaskDefFamilyRevision extracts "family:revision" from a task definition ARN.
func extractTaskDefFamilyRevision(taskDef string) string {
	// Format: arn:aws:ecs:region:account:task-definition/family:revision
	parts := strings.Split(taskDef, "/")
	if len(parts) < 2 {
		return taskDef
	}
	return parts[len(parts)-1] // Returns "family:revision"
}

// extractTaskID extracts the task ID from a task ARN.
func extractTaskID(taskARN string) string {
	// Format: arn:aws:ecs:region:account:task/cluster/task-id
	parts := strings.Split(taskARN, "/")
	if len(parts) < 1 {
		return taskARN
	}
	return parts[len(parts)-1]
}

// ListScheduledTasks retrieves EventBridge rules that target a specific ECS cluster.
func (c *Client) ListScheduledTasks(ctx context.Context, clusterARN string) ([]ScheduledTask, error) {
	// Extract cluster name from ARN for matching
	clusterName := extractClusterName(clusterARN)

	var rules []string
	var nextToken *string

	// List all rules
	for {
		listOutput, err := c.eventbridge.ListRules(ctx, &eventbridge.ListRulesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, rule := range listOutput.Rules {
			rules = append(rules, aws.ToString(rule.Name))
		}

		if listOutput.NextToken == nil {
			break
		}
		nextToken = listOutput.NextToken
	}

	if len(rules) == 0 {
		return []ScheduledTask{}, nil
	}

	var scheduledTasks []ScheduledTask

	// Check each rule's targets for ECS tasks targeting our cluster
	for _, ruleName := range rules {
		targetsOutput, err := c.eventbridge.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
			Rule: aws.String(ruleName),
		})
		if err != nil {
			continue // Skip rules we can't read targets for
		}

		for _, target := range targetsOutput.Targets {
			if target.EcsParameters == nil {
				continue
			}

			// Check if this target is for our cluster
			targetClusterName := extractClusterName(aws.ToString(target.Arn))
			if targetClusterName != clusterName {
				continue
			}

			// Get rule details
			ruleOutput, err := c.eventbridge.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
				Name: aws.String(ruleName),
			})
			if err != nil {
				continue
			}

			taskDef := extractTaskDefFamilyRevision(aws.ToString(target.EcsParameters.TaskDefinitionArn))
			taskCount := 1
			if target.EcsParameters.TaskCount != nil {
				taskCount = int(*target.EcsParameters.TaskCount)
			}

			scheduledTasks = append(scheduledTasks, ScheduledTask{
				RuleName:       ruleName,
				ScheduleExpr:   aws.ToString(ruleOutput.ScheduleExpression),
				State:          string(ruleOutput.State),
				TaskDefinition: taskDef,
				TaskCount:      taskCount,
				Description:    aws.ToString(ruleOutput.Description),
			})
		}
	}

	return scheduledTasks, nil
}

// extractClusterName extracts the cluster name from a cluster ARN or returns the input if it's already a name.
func extractClusterName(clusterARN string) string {
	// Format: arn:aws:ecs:region:account:cluster/cluster-name
	if strings.HasPrefix(clusterARN, "arn:") {
		parts := strings.Split(clusterARN, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}
	return clusterARN
}

// ExtractTaskDefFamily extracts the family name from a task definition string like "family:revision".
func ExtractTaskDefFamily(taskDef string) string {
	parts := strings.Split(taskDef, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return taskDef
}

// GetTaskMetrics retrieves CloudWatch metrics for a specific task.
// Uses Container Insights metrics from ECS/ContainerInsights namespace.
// Returns CPU and Memory utilization as percentages (0-100).
// taskDefFamily is the task definition family name (without revision).
func (c *Client) GetTaskMetrics(ctx context.Context, clusterARN, taskID, taskDefFamily string, duration time.Duration) (*TaskMetrics, error) {
	clusterName := extractClusterName(clusterARN)
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	// Determine period based on duration
	// For sparklines (15m), use 1-minute periods
	// For detail view (1h), use 5-minute periods
	period := int32(60)
	if duration > 30*time.Minute {
		period = 300
	}

	// Build dimension filter for this task (requires all 3 dimensions)
	dimensions := []cwtypes.Dimension{
		{Name: aws.String("ClusterName"), Value: aws.String(clusterName)},
		{Name: aws.String("TaskId"), Value: aws.String(taskID)},
		{Name: aws.String("TaskDefinitionFamily"), Value: aws.String(taskDefFamily)},
	}

	// Fetch CPU and Memory utilization metrics (these are already percentages)
	cpuMetrics, err := c.getMetricData(ctx, "ECS/ContainerInsights", "TaskCpuUtilization", dimensions, startTime, endTime, period)
	if err != nil {
		return nil, err
	}

	memMetrics, err := c.getMetricData(ctx, "ECS/ContainerInsights", "TaskMemoryUtilization", dimensions, startTime, endTime, period)
	if err != nil {
		return nil, err
	}

	// Combine metrics by timestamp
	metricsMap := make(map[time.Time]*TaskMetric)

	for i, ts := range cpuMetrics.timestamps {
		metricsMap[ts] = &TaskMetric{
			Timestamp:      ts,
			CPUUtilization: cpuMetrics.values[i],
		}
	}

	for i, ts := range memMetrics.timestamps {
		if m, ok := metricsMap[ts]; ok {
			m.MemoryUtilization = memMetrics.values[i]
		} else {
			metricsMap[ts] = &TaskMetric{
				Timestamp:         ts,
				MemoryUtilization: memMetrics.values[i],
			}
		}
	}

	// Convert map to sorted slice
	metrics := make([]TaskMetric, 0, len(metricsMap))
	for _, m := range metricsMap {
		metrics = append(metrics, *m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.Before(metrics[j].Timestamp)
	})

	return &TaskMetrics{
		TaskID:     taskID,
		ClusterARN: clusterARN,
		Metrics:    metrics,
	}, nil
}

type metricResult struct {
	timestamps []time.Time
	values     []float64
}

func (c *Client) getMetricData(ctx context.Context, namespace, metricName string, dimensions []cwtypes.Dimension, startTime, endTime time.Time, period int32) (*metricResult, error) {
	output, err := c.cloudwatch.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: dimensions,
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(period),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage},
	})
	if err != nil {
		return nil, err
	}

	// Sort datapoints by timestamp
	sort.Slice(output.Datapoints, func(i, j int) bool {
		return output.Datapoints[i].Timestamp.Before(*output.Datapoints[j].Timestamp)
	})

	result := &metricResult{
		timestamps: make([]time.Time, len(output.Datapoints)),
		values:     make([]float64, len(output.Datapoints)),
	}

	for i, dp := range output.Datapoints {
		result.timestamps[i] = *dp.Timestamp
		if dp.Average != nil {
			result.values[i] = *dp.Average
		}
	}

	return result, nil
}

// TaskInfo holds the info needed to fetch metrics for a task.
type TaskInfo struct {
	TaskID        string
	TaskDefFamily string
}

// GetTasksMetricsBatch retrieves metrics for multiple tasks in parallel.
func (c *Client) GetTasksMetricsBatch(ctx context.Context, clusterARN string, tasks []TaskInfo, duration time.Duration) (map[string]*TaskMetrics, error) {
	results := make(map[string]*TaskMetrics)

	// Fetch metrics for each task (could be parallelized further if needed)
	for _, task := range tasks {
		metrics, err := c.GetTaskMetrics(ctx, clusterARN, task.TaskID, task.TaskDefFamily, duration)
		if err != nil {
			// Skip tasks with metric errors but continue fetching others
			continue
		}
		results[task.TaskID] = metrics
	}

	return results, nil
}
