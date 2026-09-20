# Contributing to SensorFlow

Thank you for improving SensorFlow. Bug reports, documentation corrections, reproducible Docker issues, and focused pull requests are welcome.

## Development setup

1. Fork and clone the repository.
2. Start the local stack with `cd deploy/docker && docker compose up -d --build`.
3. Install a valid customer license at `binaries/sensors-payload-license` when testing the production ingestion path.
4. Send an `integration_test` event from an official Sensors Data SDK and verify its ClickHouse row.
5. Run the checks below before opening a pull request.

```bash
go test ./...
go build -o sensors main.go
docker compose -f deploy/docker/docker-compose.yml config
```

Keep changes small and explain their operational impact. Add or update tests when behavior changes. Never include real event data, secrets, tokens, passwords, or license files.

## Reporting bugs

Use the bug report template and include the operating system, CPU architecture, Docker/Go versions, exact commands, logs with secrets removed, expected behavior, and actual behavior. Security vulnerabilities should not be posted as public issues; contact the maintainers privately through the repository owner's published contact channel.

## Pull requests

- Link the issue or explain the problem being solved.
- Document new configuration and migration steps.
- Preserve backward compatibility where practical and call out breaking changes.
- Confirm that generated files, local volumes, credentials, and customer data are excluded.

By contributing, you agree that your contribution is licensed under the repository's Apache License 2.0.
