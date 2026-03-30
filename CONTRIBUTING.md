# Contributing

Thanks for contributing to `git-read-mcp`.

## Development setup

- Go 1.25+
- Git

```bash
git checkout -b feature/my-task
go test ./...
make fmt
make lint
make build
```

## Contribution workflow

1. Create a branch from `develop`.
2. Keep changes focused and small.
3. Add or update tests for behavioral changes.
4. Update docs when inputs/outputs or behavior change.
5. Open a PR into `develop` with:
   - concise summary
   - test plan
   - risk notes (if any)

## Commit message style

Use imperative titles (examples):

- `Add summarize_diff parser limits`
- `Fix handler error mapping`
- `Update release docs`
- `Document compare_text constraints`

## Quality bar

Before opening a PR, ensure:

```bash
go test ./...
go vet ./...
gofmt -w .
```

## Security issues

Do not open public issues for vulnerabilities. Follow `SECURITY.md`.

## License

By contributing, you agree that your contributions are licensed under Apache-2.0.
