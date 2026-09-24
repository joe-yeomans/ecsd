package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type testHTTPClient func(*http.Request) (*http.Response, error)

func (client testHTTPClient) Do(request *http.Request) (*http.Response, error) {
	return client(request)
}

func newTestClient(t *testing.T, handler testHTTPClient) *Client {
	t.Helper()
	return NewClient(aws.Config{
		Region:           "us-east-1",
		Credentials:      aws.AnonymousCredentials{},
		HTTPClient:       handler,
		RetryMaxAttempts: 1,
	}, "test")
}

func jsonResponse(t *testing.T, status int, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/x-amz-json-1.1"}},
		Body:       io.NopCloser(strings.NewReader(string(body))),
	}
}

func TestListServicesPaginatesAndBatchesDescriptions(t *testing.T) {
	const cluster = "arn:aws:ecs:us-east-1:000000000000:cluster/demo"
	var serviceARNs []string
	for i := range 12 {
		serviceARNs = append(serviceARNs, fmt.Sprintf("arn:aws:ecs:us-east-1:000000000000:service/demo/service-%02d", i))
	}
	var listCalls int
	var describeSizes []int
	var described []string
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		var input struct {
			Cluster   string   `json:"cluster"`
			NextToken string   `json:"nextToken"`
			Services  []string `json:"services"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Cluster != cluster {
			t.Fatalf("request cluster = %q, want %q", input.Cluster, cluster)
		}
		switch operation := request.Header.Get("X-Amz-Target"); {
		case strings.HasSuffix(operation, ".ListServices"):
			listCalls++
			switch listCalls {
			case 1:
				if input.NextToken != "" {
					t.Fatalf("first request token = %q", input.NextToken)
				}
				return jsonResponse(t, http.StatusOK, map[string]any{
					"serviceArns": serviceARNs[:7], "nextToken": "second-page",
				}), nil
			case 2:
				if input.NextToken != "second-page" {
					t.Fatalf("second request token = %q", input.NextToken)
				}
				return jsonResponse(t, http.StatusOK, map[string]any{"serviceArns": serviceARNs[7:]}), nil
			default:
				t.Fatalf("unexpected extra listing request: %d", listCalls)
			}
		case strings.HasSuffix(operation, ".DescribeServices"):
			describeSizes = append(describeSizes, len(input.Services))
			described = append(described, input.Services...)
			var services []map[string]any
			for _, arn := range input.Services {
				services = append(services, map[string]any{
					"serviceArn":               arn,
					"serviceName":              arn[strings.LastIndex(arn, "/")+1:],
					"status":                   "ACTIVE",
					"runningCount":             3,
					"desiredCount":             3,
					"taskDefinition":           "arn:aws:ecs:us-east-1:000000000000:task-definition/demo:7",
					"capacityProviderStrategy": []map[string]any{{"capacityProvider": "FARGATE", "weight": 1}},
				})
			}
			return jsonResponse(t, http.StatusOK, map[string]any{"services": services}), nil
		default:
			t.Fatalf("unexpected API operation: %s", operation)
		}
		return nil, fmt.Errorf("unexpected request")
	})

	services, err := client.ListServices(context.Background(), cluster)
	if err != nil {
		t.Fatal(err)
	}
	if listCalls != 2 || !reflect.DeepEqual(describeSizes, []int{10, 2}) {
		t.Fatalf("list calls = %d, describe batches = %v; want 2 and [10 2]", listCalls, describeSizes)
	}
	if !reflect.DeepEqual(described, serviceARNs) {
		t.Fatalf("described services = %v, want %v", described, serviceARNs)
	}
	if len(services) != len(serviceARNs) {
		t.Fatalf("got %d services, want %d", len(services), len(serviceARNs))
	}
	for i, service := range services {
		if service.ARN != serviceARNs[i] || service.ClusterARN != cluster ||
			service.Name != fmt.Sprintf("service-%02d", i) || service.Status != "ACTIVE" ||
			service.RunningCount != 3 || service.DesiredCount != 3 ||
			service.LaunchType != "CAPACITY_PROVIDER" || service.TaskDefFamily != "demo" {
			t.Errorf("unexpected service %d: %+v", i, service)
		}
	}
}

func TestListServicesEmptyClusterSkipsDescriptions(t *testing.T) {
	var calls int
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		calls++
		if operation := request.Header.Get("X-Amz-Target"); !strings.HasSuffix(operation, ".ListServices") {
			t.Fatalf("unexpected API operation: %s", operation)
		}
		return jsonResponse(t, http.StatusOK, map[string]any{"serviceArns": []string{}}), nil
	})
	services, err := client.ListServices(context.Background(), "demo")
	if err != nil || len(services) != 0 || calls != 1 {
		t.Fatalf("services = %v, error = %v, calls = %d; want empty result and one call", services, err, calls)
	}
}

func TestListServicesPropagatesAPIErrors(t *testing.T) {
	for _, operation := range []string{"ListServices", "DescribeServices"} {
		t.Run(operation, func(t *testing.T) {
			client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
				if strings.HasSuffix(request.Header.Get("X-Amz-Target"), "."+operation) {
					return jsonResponse(t, http.StatusBadRequest, map[string]any{
						"__type": "AccessDeniedException", "message": "test access denied",
					}), nil
				}
				return jsonResponse(t, http.StatusOK, map[string]any{"serviceArns": []string{"demo"}}), nil
			})
			services, err := client.ListServices(context.Background(), "demo")
			if err == nil || !strings.Contains(err.Error(), "test access denied") {
				t.Fatalf("error = %v, want the API error", err)
			}
			if len(services) != 0 {
				t.Fatalf("services = %v, want no result on error", services)
			}
		})
	}
}
