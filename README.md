# golog - Enhanced Logging Library

Simple yet powerful logging library for Go with structured logging support.

## Features

### ✅ Backward Compatible

All existing code works without changes!

### ✨ New Features (v2.0)

1. **Structured Logging** - Add key-value fields to logs
2. **Named Loggers** - Organize logs by component
3. **Context-Aware** - Pass loggers through context
4. **JSON Output** - For production log aggregation
5. **Stack Traces** - Debug errors with full stack traces
6. **Fatal Level** - Log and exit on critical errors

## Quick Start

### Basic Usage (Backward Compatible)

```go
import "gitlab.com/kopytkg/golog"

func main() {
    golog.LOGLEVEL = golog.INFO

    golog.Info("Server started")
    golog.Infof("Listening on port %d", 8080)
    golog.Error("Connection failed")
    golog.Debug("Debug info")
    golog.Warn("Warning message")
}
```

### Structured Logging (NEW!)

```go
// Add structured fields
golog.WithFields(golog.Fields{
    "user_id": 123,
    "action": "login",
    "duration_ms": 45,
}).Info("User logged in")

// Output: 15:30:45 INFO User logged in user_id=123 action=login duration_ms=45
```

### Named Loggers (NEW!)

```go
// Create component-specific loggers
scraper := golog.Named("scraper")
scraper.Info("Started")
// Output: 15:30:45 INFO [scraper] Started

// Combine with fields
scraper.WithFields(golog.Fields{
    "probe_id": "node-01",
    "duration_ms": 45,
}).Info("Probe scraped")
// Output: 15:30:45 INFO [scraper] Probe scraped probe_id=node-01 duration_ms=45
```

### Context-Aware Logging (NEW!)

```go
// Create logger and add to context
logger := golog.Named("api").WithFields(golog.Fields{
    "request_id": "abc-123",
})
ctx := logger.WithContext(context.Background())

// Use logger from context
golog.InfoContext(ctx, "Processing request")
// Automatically includes request_id field!
```

### JSON Output (NEW!)

```go
// Perfect for production - easily parsed by log aggregators
golog.SetOutputFormat("json")

golog.WithFields(golog.Fields{
    "event": "request_completed",
    "status": 200,
    "duration": 156,
}).Info("Request completed")

// Output: {"time":"2026-01-11T15:30:45Z","level":"info","event":"request_completed","status":200,"duration":156,"msg":"Request completed"}
```

### Error with Stack Trace (NEW!)

```go
if err != nil {
    golog.ErrorWithStack(err)  // Includes full stack trace
}
```

### Fatal Logging (NEW!)

```go
// Log error and exit with code 1
golog.Fatal("Cannot start server")
golog.Fatalf("Port %d is in use", 8080)
```

## Log Levels

```go
golog.SILENT  // No output
golog.ERROR   // Errors only
golog.WARN    // Warnings and errors
golog.INFO    // Info, warnings, and errors (default for production)
golog.DEBUG   // Everything (default for development)

golog.LOGLEVEL = golog.DEBUG  // Set level
```

## Output Control

```go
// CLI output
golog.CLILOG = golog.ENABLED   // Enable console output (default)
golog.CLILOG = golog.DISABLED  // Disable console output

// File output
golog.EnableLogFile("/var/log/app.log")  // Write to file
golog.DisableLogFile()                    // Stop file output

// Format
golog.SetOutputFormat("text")  // Human-readable (default)
golog.SetOutputFormat("json")  // Machine-readable
```

## Available Log Functions

### Simple Logging

```go
golog.Info(msg)      // Info message
golog.Success(msg)   // Success (green OK)
golog.Warn(msg)      // Warning (NEW!)
golog.SYSWarn(msg)   // System warning
golog.APPWarn(msg)   // Application warning
golog.Error(msg)     // Error
golog.Debug(msg)     // Debug
golog.Fatal(msg)     // Error + exit (NEW!)
```

### Formatted Logging

```go
golog.Infof(format, args...)
golog.Successf(format, args...)
golog.Warnf(format, args...)      // NEW!
golog.SYSWarnf(format, args...)
golog.APPWarnf(format, args...)
golog.Errorf(format, args...)
golog.Debugf(format, args...)
golog.Fatalf(format, args...)     // NEW!
```

### Logger Methods (NEW!)

```go
logger := golog.Named("component")

logger.Info(msg)
logger.Infof(format, args...)
logger.Success(msg)
logger.Successf(format, args...)
logger.Warn(msg)
logger.Warnf(format, args...)
logger.Error(msg)
logger.Errorf(format, args...)
logger.ErrorWithStack(err)         // With stack trace
logger.Debug(msg)
logger.Debugf(format, args...)
logger.Fatal(msg)
logger.Fatalf(format, args...)
```

## Real-World Example

```go
package main

import (
    "context"
    "gitlab.com/kopytkg/golog"
)

func main() {
    // Setup
    golog.LOGLEVEL = golog.INFO
    golog.SetOutputFormat("json")  // For production
    golog.EnableLogFile("/var/log/app.log")

    // Create component loggers
    api := golog.Named("api")
    db := golog.Named("database")

    // Structured logging
    api.WithFields(golog.Fields{
        "port": 8080,
        "workers": 4,
    }).Info("API server started")

    // Context-aware
    ctx := api.WithFields(golog.Fields{
        "request_id": "abc-123",
        "user": "alice",
    }).WithContext(context.Background())

    processRequest(ctx, db)
}

func processRequest(ctx context.Context, db *golog.Logger) {
    // Logger from context includes request_id and user
    golog.InfoContext(ctx, "Processing request")

    // Database operations with structured fields
    db.WithFields(golog.Fields{
        "query": "SELECT * FROM users",
        "duration_ms": 23,
        "rows": 42,
    }).Debug("Query executed")
}
```

## Migration from Old Code

No changes needed! All existing code works:

```go
// OLD CODE - Still works!
golog.Info("Server started")
golog.Errorf("Failed to connect: %v", err)
golog.Debug("Debug message")

// NEW CODE - Add features when you want
golog.WithFields(golog.Fields{"port": 8080}).Info("Server started")
```

## Performance

- Thread-safe (uses sync.RWMutex)
- Zero allocations for disabled log levels
- Lazy field formatting (only formats if level is enabled)
- No reflection - direct field access

## Color Codes (CLI Output)

- **INFO** - Cyan
- **OK** - Green (Success)
- **WARN/SYSW/APPW** - Yellow
- **ERRO** - Red
- **DEBG** - Magenta
- **FATAL** - Bold Red

## License

See [LICENSE](https://github.com/KopyTKG/GoLog?tab=CC0-1.0-1-ov-file) for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Author

kopytkg - gitlab.com/kopytkg/golog
