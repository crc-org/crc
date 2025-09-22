# unqueryvet

[![Go Report Card](https://goreportcard.com/badge/github.com/MirrexOne/unqueryvet)](https://goreportcard.com/report/github.com/MirrexOne/unqueryvet)
[![GoDoc](https://godoc.org/github.com/MirrexOne/unqueryvet?status.svg)](https://godoc.org/github.com/MirrexOne/unqueryvet)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

unqueryvet is a Go static analysis tool (linter) that detects `SELECT *` usage in SQL queries and SQL builders, encouraging explicit column selection for better performance, maintainability, and API stability.

## Features

- **Detects `SELECT *` in string literals** - Finds problematic queries in your Go code
- **SQL Builder support** - Works with popular SQL builders like Squirrel, GORM, etc.
- **Highly configurable** - Extensive configuration options for different use cases
- **Supports `//nolint:unqueryvet`** - Standard Go linting suppression
- **golangci-lint integration** - Works seamlessly with golangci-lint
- **Zero false positives** - Smart pattern recognition for acceptable `SELECT *` usage
- **Fast and lightweight** - Built on golang.org/x/tools/go/analysis

## Why avoid `SELECT *`?

- **Performance**: Selecting unnecessary columns wastes network bandwidth and memory
- **Maintainability**: Schema changes can break your application unexpectedly  
- **Security**: May expose sensitive data that shouldn't be returned
- **API Stability**: Adding new columns can break clients that depend on column order

## Informative Error Messages

Unqueryvet provides context-specific messages that explain WHY you should avoid `SELECT *`:

```go
// Basic queries
query := "SELECT * FROM users"
// avoid SELECT * - explicitly specify needed columns for better performance, maintainability and stability

// SQL Builders
query := squirrel.Select("*").From("users")
// avoid SELECT * in SQL builder - explicitly specify columns to prevent unnecessary data transfer and schema change issues

// Empty Select()
query := squirrel.Select()
// SQL builder Select() without columns defaults to SELECT * - add specific columns with .Columns() method
```

## Quick Start

### As a standalone tool

```bash
go install github.com/MirrexOne/unqueryvet/cmd/unqueryvet@latest
unqueryvet ./...
```

### With golangci-lint (Recommended)

Add to your `.golangci.yml`:

```yaml
linters:
  enable:
    - unqueryvet

linters-settings:
  unqueryvet:
    check-sql-builders: true
    # By default, no functions are ignored - minimal configuration
    # ignored-functions:
    #   - "fmt.Printf"
    #   - "log.Printf"  
    # allowed-patterns:
    #   - "SELECT \\* FROM information_schema\\..*" 
    #   - "SELECT \\* FROM pg_catalog\\..*"
```

## Examples

### Problematic code (will trigger warnings)

```go
// String literals with SELECT *
query := "SELECT * FROM users"
rows, err := db.Query("SELECT * FROM orders WHERE status = ?", "active")

// SQL builders with SELECT *
query := squirrel.Select("*").From("products")
query := builder.Select().Columns("*").From("inventory")
```

### Good code (recommended)

```go
// Explicit column selection
query := "SELECT id, name, email FROM users"
rows, err := db.Query("SELECT id, total FROM orders WHERE status = ?", "active")

// SQL builders with explicit columns
query := squirrel.Select("id", "name", "price").From("products")
query := builder.Select().Columns("id", "quantity", "location").From("inventory")
```

### Acceptable SELECT * usage (won't trigger warnings)

```go
// System/meta queries
"SELECT * FROM information_schema.tables"
"SELECT * FROM pg_catalog.pg_tables"

// Aggregate functions
"SELECT COUNT(*) FROM users"
"SELECT MAX(*) FROM scores" 

// With nolint suppression
query := "SELECT * FROM debug_table" //nolint:unqueryvet
```

## Configuration

Unqueryvet is highly configurable to fit your project's needs:

```yaml
linters-settings:
  unqueryvet:
    # Enable/disable SQL builder checking (default: true)
    check-sql-builders: true
    
    # Optional: Functions to ignore during analysis (empty by default - minimal config)
    # ignored-functions:
    #   - "fmt.Printf"
    #   - "log.Printf"
    #   - "debug.Query"
    
    # Optional: Packages to ignore completely (empty by default)  
    # ignored-packages:
    #   - "testing"
    #   - "debug"
    
    # Default allowed patterns (automatically included):
    # - COUNT(*), MAX(*), MIN(*) functions
    # - information_schema, pg_catalog, sys schema queries
    # You can add more patterns if needed:
    # allowed-patterns:
    #   - "SELECT \\* FROM temp_.*"
    
    # Default ignored file patterns (automatically included):
    # *_test.go, *.pb.go, *_gen.go, *.gen.go, *_generated.go
    # You can add more patterns if needed:
    # ignored-file-patterns:
    #   - "my_special_pattern.go"
    
    # Default ignored directories (automatically included):
    # vendor, testdata, migrations, generated, .git, node_modules  
    # You can add more directories if needed:
    # ignored-directories:
    #   - "my_special_dir"
```

## Supported SQL Builders

Unqueryvet supports popular SQL builders out of the box:

- **Squirrel** - `squirrel.Select("*")`, `Select().Columns("*")`
- **GORM** - Custom query methods
- **SQLBoiler** - Generated query methods
- **Custom builders** - Any builder using `Select()` patterns

## Integration Examples

### GitHub Actions

```yaml
name: Lint
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: 1.23
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest
        args: --enable unqueryvet
```

## Command Line Options

When used as a standalone tool:

```bash
# Check all packages
unqueryvet ./...

# Check specific packages
unqueryvet ./cmd/... ./internal/...

# With custom config file
unqueryvet -config=.unqueryvet.yml ./...

# Verbose output
unqueryvet -v ./...
```

## Performance

Unqueryvet is designed to be fast and lightweight:

- **Parallel processing** - Analyzes multiple files concurrently
- **Incremental analysis** - Only analyzes changed files when possible
- **Minimal memory footprint** - Efficient AST traversal
- **Smart caching** - Reuses analysis results when appropriate

## Advanced Usage

### Custom Patterns

You can define custom regex patterns for acceptable `SELECT *` usage:

```yaml
allowed-patterns:
  # Allow SELECT * from temporary tables
  - "SELECT \\* FROM temp_\\w+"
  # Allow SELECT * in migration scripts  
  - "SELECT \\* FROM.*-- migration"
  # Allow SELECT * for specific schemas
  - "SELECT \\* FROM audit\\..+"
```

### Integration with Custom SQL Builders

For custom SQL builders, Unqueryvet looks for these patterns:

```go
// Method chaining
builder.Select("*")          // Direct SELECT *
builder.Select().Columns("*") // Chained SELECT *

// Variable tracking  
query := builder.Select()    // Empty select
// If no .Columns() call follows, triggers warning
```

### Running Tests

```bash
go test ./...
go test -race ./...
go test -bench=. ./...
```

### Development Setup

```bash
git clone https://github.com/MirrexOne/unqueryvet.git
cd unqueryvet
go mod tidy
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Bug Reports**: [GitHub Issues](https://github.com/MirrexOne/unqueryvet/issues)

---
