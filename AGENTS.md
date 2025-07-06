# SnakeLSP Development Context

## Build/Test Commands

- **Build**: `go build -o snakelsp`
- **Build with Sentry**: `go build -ldflags "-X main.sentryDSN=$SENTRY_DSN" -o snakelsp`
- **Run**: `./snakelsp` (LSP server via stdio)
- **Test all**: `go test ./...`
- **Test single package**: `go test ./internal/workspace`
- **Test single file**: `go test ./internal/workspace -run TestParseSymbols`
- **Format**: `go fmt ./...`
- **Vet**: `go vet ./...`

## Code Style Guidelines

- **Package naming**: lowercase, single word (e.g., `workspace`, `protocol`, `messages`)
- **Imports**: Group stdlib, third-party, local packages with blank lines between groups
- **Types**: PascalCase for exported, camelCase for unexported
- **Functions**: PascalCase for exported handlers (e.g., `HandleInitialize`), camelCase for internal
- **Variables**: camelCase, descriptive names (e.g., `pythonCode`, `mockFile`)
- **Constants**: PascalCase or UPPER_SNAKE_CASE for package-level
- **Error handling**: Always check errors, return early on error
- **Testing**: Use testify/assert, create mock data in test functions
- **Logging**: Use structured logging with slog package
- **JSON**: Use struct tags for LSP message serialization
