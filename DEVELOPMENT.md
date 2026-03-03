# Development Setup

## Prerequisites

- **Go 1.25+**
- **Git**
- **golangci-lint** (optional, for local linting)

## Quick Start

```bash
# 1. Clone
git clone https://github.com/pinchtab/pinchtab.git
cd pinchtab

# 2. Setup (installs git hooks, downloads deps)
make setup

# 3. Build and run
make build
./pinchtab
```

That's it! Git hooks are installed automatically and will run on every commit.

## Detailed Setup

### 1. Clone the repository

```bash
git clone https://github.com/pinchtab/pinchtab.git
cd pinchtab
```

### 2. Run automated setup

```bash
make setup
```

This will:
- Install git hooks (gofmt + golangci-lint checks before commit)
- Download Go dependencies
- Verify your environment

### 3. (Optional) Install golangci-lint

For local linting (recommended):

```bash
# macOS/Linux
brew install golangci-lint

# Or via Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Without golangci-lint, commits will still work but won't check linting locally.

## Before Committing

Git hooks automatically run on `git commit`. To manually check your code:

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# All checks (format + lint + test)
make check
```

## Common Issues

### "Git hooks not running on commit"

Re-run setup:
```bash
make install-hooks
```

Verify hooks installed:
```bash
cat .git/hooks/pre-commit
```

### "golangci-lint: command not found" during commit

Hooks will warn but still allow commit. To fix:
```bash
brew install golangci-lint
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### gofmt fails in CI even though local commit worked

Run format before committing:
```bash
make fmt
```

### Tests failing locally

```bash
# Run full test suite
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestName ./...
```

## Running Tests

```bash
# All tests
make test

# With coverage report
make test-cover
# Opens coverage.html in browser
```

Or directly with Go:
```bash
go test ./...                    # All tests
go test -v ./...                 # Verbose
go test -run TestName ./...      # Specific test
```

## Code Style

- **Format:** `gofmt` (automatic via git hook, or run `make fmt`)
- **Lint:** `golangci-lint` (automatic via git hook, or run `make lint`)
- **Tests:** Must pass (`make test`)

## Git Workflow

```bash
# 1. Create branch
git checkout -b feature/your-feature

# 2. Make changes
# ... edit files ...

# 3. Check your work (optional, hooks will run on commit)
make check

# 4. Commit (git hooks run automatically: gofmt + lint)
git commit -m "feat: description"

# 5. Push
git push origin feature/your-feature

# 6. Create Pull Request on GitHub
```

**Note:** Git hooks automatically run `gofmt` and `golangci-lint` on staged files before each commit. If checks fail, the commit is blocked.

## Documentation

Update docs when adding features:

```bash
# Docs location
docs/
├── core-concepts.md
├── get-started.md
├── references/
├── architecture/
└── guides/
```

Validate docs: `./scripts/check-docs-json.sh`

## Useful Commands

```bash
# Development
make setup          # Setup dev environment (run once)
make build          # Build pinchtab binary
make run            # Build and run
make fmt            # Format code
make lint           # Run linter
make test           # Run tests
make test-cover     # Run tests with coverage
make check          # Format + lint + test
make clean          # Remove build artifacts
make help           # Show all commands

# Direct Go commands
gofmt -w .          # Format all files
gofmt -l .          # List files that need formatting
go test ./...       # Run all tests
go test -v ./...    # Verbose tests
go clean            # Clean build cache
go get -u ./...     # Update dependencies
```

## Getting Help

- Read the [Overview](docs/overview.md)
- Check [Architecture](docs/architecture/pinchtab-architecture.md)
- See [API Reference](docs/references/instance-api.md)
- Browse [Guides](docs/guides/)
