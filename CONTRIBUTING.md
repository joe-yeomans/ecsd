# Contributing

Bug reports, small fixes, and improvements are welcome. For larger changes, open an issue first to discuss the proposed behaviour.

## Local checks

Use Go 1.26.1 or later and run:

```sh
gofmt -w .
go mod tidy -diff
go vet ./...
go test -race ./...
go build -o /tmp/ecsd .
```

Tests must run without AWS credentials or a live account. Use simulated responses for AWS API behaviour, and include a regression test when fixing a bug.

Keep pull requests focused. Explain the behaviour being changed and the checks you ran. Include screenshots for visual changes, using fictional resource names and account details.

## Bug reports

Include the operating system, terminal, Go version, steps to reproduce, and expected versus actual behaviour. Remove credentials, account IDs, resource names, and other private information from logs and screenshots.

Report security issues using the instructions in [SECURITY.md](SECURITY.md).

Contributions are licensed under the project's [MIT licence](LICENSE).
