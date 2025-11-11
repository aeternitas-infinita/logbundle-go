# logbundle-go

A comprehensive Go logging and error handling library designed for production applications with deep integration for Fiber web framework and Sentry error tracking.

## Features

- **Structured Logging**: Built on Go's standard `log/slog` with custom formatting
- **Rich Error Types**: Pre-defined error categories with automatic HTTP status mapping
- **Stack Trace Capture**: Automatic stack trace collection with intelligent frame filtering
- **Panic Recovery**: Automatic panic recovery middleware prevents application crashes
- **Fiber Integration**: Drop-in error handler and middleware for Fiber applications
- **Sentry Integration**: Automatic error reporting with context enrichment and HTTP status filtering
- **RFC 7807 Compliant**: Problem Details for HTTP APIs standard support
- **Validation Errors**: Field-level validation error collection and reporting
- **Thread-Safe**: Concurrent-safe operations with mutex protection

## Installation

```bash
go get github.com/aeternitas-infinita/logbundle-go
```

## Quick Start

### Basic Logging

```go
package main

import (
    "github.com/aeternitas-infinita/logbundle-go"
)

func main() {
    // Simple logging
    logbundle.Info("Application started")
    logbundle.Debug("Debug information", "key", "value")
    logbundle.Warn("Warning message")
    logbundle.Error("Error occurred", logbundle.ErrAttr(err))
}
```

### Context-Aware Logging

```go
func handler(ctx context.Context) error {
    logbundle.InfoCtx(ctx, "Processing request", "user_id", userID)
    logbundle.ErrorCtx(ctx, "Failed to process", logbundle.ErrAttr(err))
    return nil
}
```

### Custom Logger Configuration

```go
import (
    "log/slog"
    "github.com/aeternitas-infinita/logbundle-go"
)

func main() {
    // Configure global logger
    logbundle.InitLog(logbundle.LoggerConfig{
        Level:     slog.LevelDebug,
        AddSource: true,
    })

    // Or create a custom logger instance
    logger := logbundle.CreateLogger(logbundle.LoggerConfig{
        Level:     slog.LevelInfo,
        AddSource: false,
    })
}
```

## Error Handling

### Creating Errors

```go
import "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"

// Using factory functions (recommended)
err := lgerr.NotFound("User", 123)
err := lgerr.Validation("email", "invalid format")
err := lgerr.Internal("database connection failed")

// Using builder pattern
err := lgerr.New("custom error message").
    WithType(lgerr.TypeBadInput).
    WithTitle("Invalid Request").
    WithDetail("The provided data is invalid").
    WithContext("field", "email").
    WithContext("value", "invalid@")
```

### Error Types and HTTP Status Codes

| Error Type | HTTP Status | Use Case |
|------------|-------------|----------|
| `TypeInternal` | 500 | Internal server errors |
| `TypeNotFound` | 404 | Resource not found |
| `TypeValidation` | 400 | Validation failures |
| `TypeDatabase` | 500 | Database errors |
| `TypeBusy` | 503 | Service unavailable |
| `TypeForbidden` | 403 | Access forbidden |
| `TypeBadInput` | 400 | Bad request input |
| `TypeUnauth` | 401 | Unauthorized access |
| `TypeConflict` | 409 | Resource conflicts |
| `TypeExternal` | 502 | External service errors |
| `TypeTimeout` | 504 | Request timeouts |

### Custom Error Types

```go
// Register a custom error type
const TypeRateLimited lgerr.ErrorType = "rate_limited"
lgerr.RegisterErrorType(TypeRateLimited, 429)

// Or override existing mappings
lgerr.SetHTTPStatusMap(map[lgerr.ErrorType]int{
    lgerr.TypeNotFound: 410,  // Use 410 Gone instead of 404
    lgerr.TypeBusy:     429,  // Use 429 Too Many Requests
})
```

### Validation Errors

```go
err := lgerr.Validation("form validation failed", "").
    WithValidationError("email", "must be valid email", "invalid@").
    WithValidationError("age", "must be at least 18", 15).
    WithTitle("Validation Failed").
    WithDetail("Please correct the errors and try again")

// Access validation errors
if err.HasValidationErrors() {
    for _, ve := range err.ValidationErrors() {
        fmt.Printf("Field: %s, Error: %s\n", ve.Field, ve.Message)
    }
}
```

### Wrapping Errors

```go
dbErr := database.Query(...)
if dbErr != nil {
    return lgerr.Database("failed to fetch user").
        Wrap(dbErr).
        WithContext("user_id", userID)
}
```

## Fiber Integration

### Setup

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/aeternitas-infinita/logbundle-go"
    "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgfiber"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: lgfiber.ErrorHandler,
    })

    // Add middleware (recommended)
    app.Use(lgfiber.RecoverMiddleware())              // Panic recovery
    app.Use(lgfiber.BreadcrumbsMiddleware())          // Request breadcrumbs
    app.Use(lgfiber.ContextEnrichmentMiddleware())    // Request context
    app.Use(lgfiber.PerformanceMiddleware())          // Performance tracking

    app.Listen(":3000")
}
```

### Using in Handlers

```go
func getUserHandler(c *fiber.Ctx) error {
    userID := c.Params("id")

    user, err := database.FindUser(userID)
    if err != nil {
        // Return lgerr.Error - automatically logged and sent to Sentry
        return lgerr.NotFound("User", userID).
            WithTitle("User Not Found").
            WithDetail("The requested user does not exist")
    }

    return c.JSON(user)
}
```

### Manual Error Handling

For goroutines or background tasks where you can't return an error:

```go
func handler(c *fiber.Ctx) error {
    // Async operation
    go func() {
        if err := doBackgroundTask(); err != nil {
            lgErr := lgerr.Internal("background task failed").Wrap(err)

            // With Fiber context (includes request data)
            lgfiber.HandleErrorWithFiber(c, lgErr)

            // Or without Fiber context
            lgfiber.HandleError(c.UserContext(), lgErr)
        }
    }()

    return c.JSON(fiber.Map{"status": "processing"})
}
```

### Middleware Details

#### RecoverMiddleware
Recovers from panics and prevents application crashes:
- Captures panic value and stack trace
- Sends panic details to Sentry (if enabled)
- Logs comprehensive panic information
- Returns 500 error to client
- **IMPORTANT**: Place this first in your middleware chain

#### BreadcrumbsMiddleware
Adds HTTP request breadcrumbs to Sentry for request flow tracking:
- Request start/end events
- Duration and status code
- Request details (URL, method, path)

#### ContextEnrichmentMiddleware
Enriches Sentry scope with request data:
- HTTP method, route, host
- Query and route parameters
- User identification (if available)
- Request headers and metadata

#### PerformanceMiddleware
Creates Sentry performance transactions:
- Request duration tracking
- Distributed tracing support
- Transaction status based on HTTP status

## Sentry Integration

### Enable/Disable Sentry

```go
import (
    "github.com/aeternitas-infinita/logbundle-go"
    "github.com/getsentry/sentry-go"
)

func main() {
    // Initialize Sentry
    sentry.Init(sentry.ClientOptions{
        Dsn: "your-dsn-here",
        Environment: "production",
    })

    // Enable Sentry integration
    logbundle.SetSentryEnabled(true)

    // Check if enabled
    if logbundle.IsSentryEnabled() {
        // ...
    }
}
```

### Control Sentry Reporting

#### Filter by HTTP Status Code

```go
// Configure minimum HTTP status to send to Sentry
logbundle.SetSentryMinHTTPStatus(500)  // Only 5xx errors (default)
logbundle.SetSentryMinHTTPStatus(400)  // 4xx and 5xx errors
logbundle.SetSentryMinHTTPStatus(0)    // All errors

// Get current setting
minStatus := logbundle.GetSentryMinHTTPStatus()
```

#### Skip Specific Errors

```go
// Skip Sentry for specific errors
err := lgerr.NotFound("Resource", id).
    IgnoreSentry()  // Won't be sent to Sentry

// Unhandled errors and panics are always sent to Sentry (if enabled)
// Generic errors are converted to lgerr.Internal automatically
```

### Custom Sentry Context

```go
// In your Fiber handlers
lgfiber.SetTag(c, "feature", "user-management")
lgfiber.SetContext(c, "custom_data", map[string]any{
    "operation": "user_update",
    "batch_id":  batchID,
})

lgfiber.AddBreadcrumb(c, "database", "Query executed", sentry.LevelInfo, map[string]any{
    "query": "SELECT * FROM users",
})
```

## Best Practices

### 1. Use Appropriate Error Types

Choose error types that match the situation:
```go
// Good
return lgerr.NotFound("User", userID)
return lgerr.Validation("email", "invalid format")

// Avoid generic errors
// Bad: lgerr.New("user not found")  // Should use NotFound
```

### 2. Add Context to Errors

Always add relevant context:
```go
return lgerr.Database("failed to insert record").
    WithContext("table", "users").
    WithContext("operation", "insert").
    Wrap(dbErr)
```

### 3. Use Titles and Details for Client-Facing Errors

Separate internal messages from public-facing ones:
```go
return lgerr.BadInput("invalid email format: missing @ symbol").
    WithTitle("Invalid Email").
    WithDetail("Please provide a valid email address")
```

### 4. Handle Validation Errors Properly

Group validation errors together:
```go
if len(errors) > 0 {
    err := lgerr.Validation("form validation failed", "").
        WithTitle("Validation Failed")

    for field, msg := range errors {
        err.WithValidationError(field, msg, formData[field])
    }

    return err
}
```

### 5. Configure Log Levels Appropriately

```bash
# Development
export log_level=debug

# Production
export log_level=warn
```

### 6. Use Context-Aware Logging

Always use context-aware functions when context is available:
```go
// Good
logbundle.InfoCtx(ctx, "message")

// Avoid
logbundle.Info("message")  // Loses context
```

### 7. Don't Send Sensitive Errors to Sentry

```go
err := lgerr.Internal("authentication failed").
    WithContext("reason", "invalid credentials").
    IgnoreSentry()  // Don't send auth failures to Sentry
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `log_level` | `warn` | Minimum log level (debug, info, warn, error) |

## Performance Considerations

1. **Source Information**: The global `Log` variable includes source file/line info. For high-throughput scenarios, consider creating a custom logger without source info.

2. **Pre-allocated Slices**: Maps and slices are pre-allocated where possible to reduce allocations.

3. **Stack Trace Filtering**: Intelligent filtering skips internal frames, reducing noise and improving readability.

4. **Sentry Batching**: Sentry SDK handles batching automatically. Consider adjusting `MaxErrorDepth` and `SampleRate` for high-volume applications.

## Thread Safety

- All error operations are safe for concurrent use after creation
- HTTP status map modifications are protected by `sync.RWMutex`
- Sentry enable/disable is thread-safe
- Logger instances are safe for concurrent use

## Examples

See the [examples](./examples) directory for more detailed examples:
- Basic logging setup
- Fiber application with error handling
- Custom error types
- Sentry integration
- Validation error handling

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: https://github.com/aeternitas-infinita/logbundle-go/issues
- Documentation: [GoDoc](https://pkg.go.dev/github.com/aeternitas-infinita/logbundle-go)
