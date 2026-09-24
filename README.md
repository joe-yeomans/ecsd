# ecsd

A terminal dashboard for Amazon ECS.

Browse clusters, services, running and recently stopped tasks, container status, service events, and scheduled tasks without leaving your terminal. Task views include CPU and memory graphs from CloudWatch.

ecsd makes read-only AWS API calls. It uses your configured AWS credentials and region.

## Install

Requires Go 1.26.1 or later; use a current patched Go release.

```sh
go install github.com/joe-yeomans/ecsd@latest
```

Go installs the executable into `GOBIN`, or `$(go env GOPATH)/bin` when `GOBIN` is unset. Add that directory to your `PATH`.

To build from a local checkout and install to `$HOME/bin/ecsd`:

```sh
bash build/build.sh
```

## Run

Configure an AWS profile and region using the standard AWS SDK configuration sources, then run:

```sh
AWS_PROFILE=my-profile AWS_REGION=us-east-1 ecsd
```

For a profile using IAM Identity Center, sign in first with `aws sso login --profile my-profile`. If your region is already configured in the profile, you can omit `AWS_REGION`.

During development, run from the project directory:

```sh
AWS_PROFILE=my-profile go run .
```

The dashboard refreshes every five seconds. Press `a` to turn automatic refresh off or on.

### Keyboard controls

| Key | Action |
| --- | --- |
| Up / `k`, Down / `j` | Move selection |
| Enter | Open the selected cluster, service, or task |
| Esc / Backspace | Return to the previous view |
| Tab | Cycle cluster tabs |
| `1`, `2`, `3` | Select services, tasks, or scheduled tasks |
| `e` | Show events for the selected service on the services tab |
| `r` | Refresh the current view |
| `a` | Toggle automatic refresh |
| `q` / Ctrl+C | Quit |

## AWS permissions and metrics

[docs/iam-policy.json](docs/iam-policy.json) lists the read permissions used by the dashboard. It is an example policy for discovery across an account; restrict resources where the API supports it. Apply it to an existing identity rather than putting credentials in the project. Profiles that assume a role also need their normal role access configured.

Task graphs use `TaskCpuUtilization` and `TaskMemoryUtilization` in the `ECS/ContainerInsights` namespace. These require [Container Insights with enhanced observability](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Container-Insights-enhanced-observability-metrics-ECS.html) and available task datapoints. Empty graphs can mean metrics are unavailable; cluster and task browsing still works without them.

## Current limitations

- A session uses one AWS profile and region. Restart with different configuration to switch.
- Scheduled tasks come from EventBridge rules on the default event bus. EventBridge Scheduler schedules are not included.
- Cluster listing and each rule's target listing currently read only the first API page.
- Rules or task metrics that cannot be read may be omitted from aggregate views.

## Development

Tests use simulated AWS responses and do not require AWS credentials or a live account.

```sh
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidance and [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## Licence

[MIT](LICENSE), copyright Joe Yeomans. Third-party dependencies retain their own licences.
