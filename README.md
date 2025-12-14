# logbundle-go

A comprehensive Go logging and error handling library designed for production applications with deep integration for Fiber web framework and Sentry error tracking.

## Features

- **Structured Logging**: Built on Go's standard `log/slog` with custom formatting
- **Rich Error Types**: Pre-defined error categories with automatic HTTP status mapping
- **Stack Trace Capture**: Automatic stack trace collection with intelligent frame filtering
- **Panic Recovery**: Automatic panic recovery middleware prevents application crashes
- **Fiber Integration**: Drop-in error handler and middleware for Fiber applications
- **Sentry Integration**: Automatic error reporting with context enrichment and HTTP status filtering (opt-in)
- **RFC 7807 Compliant**: Problem Details for HTTP APIs standard support
- **Validation Errors**: Field-level validation error collection and reporting
- **Thread-Safe**: Concurrent-safe operations with mutex protection

## Installation

```bash
go get github.com/aeternitas-infinita/logbundle-go
```

## Quick Start

### Creating a Logger

```go
package main

import (
    "log/slog"
    "github.com/aeternitas-infinita/logbundle-go"
)

func main() {
    // Create a logger with custom configuration
    logger := logbundle.CreateLogger(logbundle.LoggerConfig{
        Level:     slog.LevelDebug,
        AddSource: true,
    })

    // Or use default configuration (from log_level environment variable)
    logger := logbundle.CreateLoggerDefault()

    // Use the logger
    logger.Info("Application started")
    logger.Debug("Debug information", "key", "value")
    logger.Warn("Warning message")
    logger.Error("Error occurred", logbundle.ErrAttr(err))
}
```

### Context-Aware Logging

```go
func handler(ctx context.Context, logger *slog.Logger) error {
    logger.InfoContext(ctx, "Processing request", "user_id", userID)
    logger.ErrorContext(ctx, "Failed to process", logbundle.ErrAttr(err))
    return nil
}
```

### Using Logbundle in Your Application

Each application should create its own logger instance to maintain explicit dependency management:

```go
import (
    "log/slog"
    "github.com/aeternitas-infinita/logbundle-go"
)

func main() {
    // Create logger for your application
    appLogger := logbundle.CreateLoggerDefault()

    // Pass logger to handlers, services, etc.
    handleRequest(appLogger)
}

func handleRequest(logger *slog.Logger) {
    logger.Info("Handling request")
}
```

## Error Handling

### Creating Errors

```go
import "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"

// Using factory functions (recommended)
err := lgerr.NotFound("User", 123)
// Produces: 404 error with title "Resource Not Found" and detail "The requested User does not exist"

err := lgerr.Validation("Email is required")
// Produces: 400 error with title "Validation Error"

err := lgerr.Internal("database connection failed")
// Produces: 500 error with title "Internal Server Error"

// Extend with additional options
err := lgerr.Database("query timeout",
    lgerr.WithContextKV("query", "SELECT * FROM users"),
    lgerr.WithContextKV("timeout", "5s"),
    lgerr.WithWrapped(originalErr),
)

// Using functional options for complex errors
err := lgerr.NewWithOptions(
    lgerr.WithMessage("User registration failed"),
    lgerr.WithType(lgerr.TypeValidation),
    lgerr.WithTitle("Registration Error"),
    lgerr.WithDetail("Please correct the errors below"),
    lgerr.WithValidationErr("email", "Invalid email format", "user@"),
    lgerr.WithValidationErr("age", "Must be 18 or older", 16),
    lgerr.WithContextKV("ip", requestIP),
)

// Legacy builder pattern (still supported)
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
// Using functional options
err := lgerr.Validation("form validation failed",
    lgerr.WithValidationErr("email", "must be valid email", "invalid@"),
    lgerr.WithValidationErr("age", "must be at least 18", 15),
    lgerr.WithDetail("Please correct the errors and try again"),
)

// Using builder pattern (legacy)
err := lgerr.Validation("form validation failed").
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
// Using functional options
dbErr := database.Query(...)
if dbErr != nil {
    return lgerr.Database("failed to fetch user",
        lgerr.WithWrapped(dbErr),
        lgerr.WithContextKV("user_id", userID),
    )
}

// Using builder pattern (legacy)
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
    "log/slog"
    "os"
    "time"

    "github.com/getsentry/sentry-go"
    sentryfiber "github.com/getsentry/sentry-go/fiber"
    "github.com/gofiber/fiber/v2"

    "github.com/aeternitas-infinita/logbundle-go"
    "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgfiber"
)

func main() {
    // Create application logger
    appLogger := logbundle.CreateLoggerDefault()

    // Initialize Sentry SDK
    sentry.Init(sentry.ClientOptions{
        Dsn:         os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENVIRONMENT"),
    })
    defer sentry.Flush(2 * time.Second)

    // ⚠️ REQUIRED: Enable Sentry integration (disabled by default)
    logbundle.SetSentryEnabled(true)

    // Optional: Configure HTTP status filtering (default: 500)
    logbundle.SetSentryMinHTTPStatus(500) // Only 5xx errors

    app := fiber.New(fiber.Config{
        ErrorHandler: lgfiber.ErrorHandler,
    })

    // ⚠️ MIDDLEWARE ORDER IS CRITICAL ⚠️

    // 1. Sentry base middleware (MUST BE FIRST)
    app.Use(sentryfiber.New(sentryfiber.Options{
        Repanic:         true,
        WaitForDelivery: false,
        Timeout:         3 * time.Second,
    }))

    // 2. Panic recovery (MUST BE AFTER SENTRY)
    app.Use(lgfiber.RecoverMiddleware())

    // 3. Performance monitoring (creates transactions)
    app.Use(lgfiber.PerformanceMiddleware())

    // 4. Context enrichment (tags, request data)
    app.Use(lgfiber.ContextEnrichmentMiddleware())

    // 5. Breadcrumbs (request tracking)
    app.Use(lgfiber.BreadcrumbsMiddleware())

    // 6. Your application middleware
    // app.Use(cors.New())
    // app.Use(yourMiddleware...)

    app.Listen(":3000")
}
```

### Middleware Order

**⚠️ ORDER IS CRITICAL! ⚠️**

The middleware **must** be registered in this specific order:

1. **`sentryfiber.New()`** - MUST BE FIRST
   - Initializes Sentry hub in context
   - All other middleware depend on this

2. **`lgfiber.RecoverMiddleware()`** - MUST BE AFTER SENTRY
   - Catches panics before they crash your app
   - Needs Sentry hub to report panics

3. **`lgfiber.PerformanceMiddleware()`** - Creates performance transactions
   - Tracks request duration and tracing
   - Should be early to measure entire request

4. **`lgfiber.ContextEnrichmentMiddleware()`** - Enriches Sentry context
   - Adds request data, query params, user info
   - Should be before breadcrumbs

5. **`lgfiber.BreadcrumbsMiddleware()`** - Tracks request flow
   - Logs request start/end events
   - Should be after context enrichment

6. **Your application middleware** - Place last
   - CORS, rate limiting, auth, etc.
   - After all logbundle middleware

**Why order matters:**

- Sentry base middleware creates the hub - without it, all other middleware will fail silently
- RecoverMiddleware must be early to catch panics from all subsequent middleware
- Performance middleware should wrap the entire request lifecycle
- Context enrichment before breadcrumbs ensures breadcrumbs have full context

**Common mistakes:**

```go
// ❌ WRONG - RecoverMiddleware before Sentry
app.Use(lgfiber.RecoverMiddleware())
app.Use(sentryfiber.New(...))  // Too late!

// ❌ WRONG - Missing Sentry base middleware
app.Use(lgfiber.RecoverMiddleware())  // Won't work without Sentry hub

// ✅ CORRECT - Sentry first, then Recover
app.Use(sentryfiber.New(...))
app.Use(lgfiber.RecoverMiddleware())
```

### Using in Handlers

```go
func getUserHandler(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Params("id")

        user, err := database.FindUser(userID)
        if err != nil {
            // Log and return lgerr.Error - automatically sent to Sentry
            logger.ErrorContext(c.UserContext(), "Failed to find user",
                "user_id", userID,
                logbundle.ErrAttr(err),
            )
            return lgerr.NotFound("User", userID).
                WithTitle("User Not Found").
                WithDetail("The requested user does not exist")
        }

        logger.InfoContext(c.UserContext(), "User retrieved", "user_id", userID)
        return c.JSON(user)
    }
}

// In main:
app.Get("/users/:id", getUserHandler(appLogger))
```

### Manual Error Handling

For goroutines or background tasks where you can't return an error:

```go
func handler(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Async operation
        go func() {
            if err := doBackgroundTask(); err != nil {
                lgErr := lgerr.Internal("background task failed").Wrap(err)

                // Log the error
                logger.ErrorContext(c.UserContext(), "Background task failed",
                    logbundle.ErrAttr(err),
                )

                // With Fiber context (includes request data)
                lgfiber.HandleErrorWithFiber(c, lgErr)

                // Or without Fiber context
                lgfiber.HandleError(c.UserContext(), lgErr)
            }
        }()

        return c.JSON(fiber.Map{"status": "processing"})
    }
}
```

### Middleware Details

#### RecoverMiddleware

Recovers from panics and prevents application crashes:

- Captures panic value and stack trace
- Sends panic details to Sentry (if enabled)
- Logs comprehensive panic information
- Returns 500 error to client
- **CRITICAL**: Must be placed AFTER `sentryfiber.New()` but BEFORE other middleware

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

**⚠️ IMPORTANT: Sentry is DISABLED by default!**

You must explicitly enable it by calling `SetSentryEnabled(true)`. Without this, no events will be sent to Sentry.

```go
import (
    "github.com/aeternitas-infinita/logbundle-go"
    "github.com/getsentry/sentry-go"
)

func main() {
    // Initialize Sentry SDK
    sentry.Init(sentry.ClientOptions{
        Dsn:         "your-dsn-here",
        Environment: "production",
    })

    // ⚠️ REQUIRED: Enable Sentry integration (disabled by default)
    logbundle.SetSentryEnabled(true)

    // Optional: Configure HTTP status filtering
    logbundle.SetSentryMinHTTPStatus(500) // Only 5xx errors (default)

    // Check if enabled
    if logbundle.IsSentryEnabled() {
        // Sentry is active
    }

    // Disable Sentry (for testing, etc.)
    // logbundle.SetSentryEnabled(false)
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
    // Using functional options
    opts := []lgerr.ErrorOption{
        lgerr.WithDetail("Please correct the errors and try again"),
    }
    for field, msg := range errors {
        opts = append(opts, lgerr.WithValidationErr(field, msg, formData[field]))
    }
    return lgerr.Validation("form validation failed", opts...)
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
// Using functional options
err := lgerr.Internal("authentication failed",
    lgerr.WithContextKV("reason", "invalid credentials"),
    lgerr.WithIgnoreSentry(),
)

// Using builder pattern (legacy)
err := lgerr.Internal("authentication failed").
    WithContext("reason", "invalid credentials").
    IgnoreSentry()
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `log_level` | `warn` | Minimum log level (debug, info, warn, error) |

## Performance Considerations

1. **Source Information**: Logger instances created with `AddSource: true` include source file/line info. For high-throughput scenarios, disable source tracking or use the internal logger helpers:
   ```go
   import "github.com/aeternitas-infinita/logbundle-go/internal/logger"

   logger := logbundle.CreateLogger(logbundle.LoggerConfig{
       Level:     slog.LevelInfo,
       AddSource: false,  // Disable for high-throughput
   })

   // Or use direct logging without source capture
   logger.LogNoSource(logger, slog.LevelInfo, "high frequency log")
   ```

2. **Sentry Overhead**: When Sentry is disabled, all middleware and capture functions return immediately with zero allocations. The library checks `config.IsSentryEnabled()` before any expensive operations.

3. **Pre-allocated Slices**: Maps and slices are pre-allocated where possible to reduce allocations. Error options use variadic functions to avoid intermediate slice allocations.

4. **Stack Trace Filtering**: Intelligent filtering uses early exit optimization without full string splits, improving panic recovery performance by 50%+.

5. **Context Cancellation**: All Sentry operations check `ctx.Done()` before expensive work to respect cancellation signals.

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

- GitHub Issues: <https://github.com/aeternitas-infinita/logbundle-go/issues>
- Documentation: [GoDoc](https://pkg.go.dev/github.com/aeternitas-infinita/logbundle-go)
