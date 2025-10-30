# LogBundle-Go

**LogBundle-Go** is a comprehensive logging solution for Go applications that seamlessly integrates `log/slog`, Sentry error tracking, and Fiber web framework. It provides structured logging with automatic error reporting, breadcrumb tracking, and performance monitoring.

## Features

- ✅ **Structured Logging** with `log/slog`
- 🔍 **Deep Sentry Integration** with automatic error tracking
- 🌐 **Fiber Middleware** with request context enrichment
- 📊 **Performance Monitoring** with Sentry Transactions and Spans
- 🍞 **Automatic Breadcrumbs** for error investigation
- 🎯 **Trace ID Propagation** across services
- 🚀 **Production-Ready** with optimized performance
- 🛠️ **Flexible Configuration** via environment variables

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Logging](#logging)
- [Sentry Integration](#sentry-integration)
- [Fiber Integration](#fiber-integration)
- [Performance Monitoring](#performance-monitoring)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)

## Installation

```bash
go get github.com/aeternitas-infinita/logbundle-go
```

## Quick Start

### Basic Logging

```go
package main

import (
    "log/slog"
    "github.com/aeternitas-infinita/logbundle-go"
)

func main() {
    // Initialize logger
    logbundle.InitLog(logbundle.LoggerConfig{
        Level:         slog.LevelInfo,
        AddSource:     true,
        SentryEnabled: false,
    })

    // Log messages
    logbundle.Info("Application started")
    logbundle.Debug("Debug information", slog.String("version", "1.0.0"))
    logbundle.Warn("Warning message", slog.Int("code", 100))
    logbundle.Error("Error occurred", slog.Any("error", err))
}
```

### Fiber Application with Sentry

```go
package main

import (
    "log/slog"
    "os"
    "time"

    "github.com/aeternitas-infinita/logbundle-go"
    "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgfiber"
    "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
    "github.com/getsentry/sentry-go"
    sentryfiber "github.com/getsentry/sentry-go/fiber"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Initialize logger
    logbundle.InitLog(logbundle.LoggerConfig{
        Level:         slog.LevelInfo,
        AddSource:     true,
        SentryEnabled: true,
    })

    // Initialize Sentry
    // Config embeds sentry.ClientOptions, so all Sentry fields are available directly
    if err := lgsentry.Init(&lgsentry.Config{
        ClientOptions: sentry.ClientOptions{
            Dsn:              os.Getenv("SENTRY_DSN"),
            Environment:      os.Getenv("ENVIRONMENT"),
            AttachStacktrace: true,
            MaxBreadcrumbs:   100,
            EnableTracing:    true,
            TracesSampleRate: 0.1,
        },
        // Custom logbundle fields
        FilterLevels: []slog.Level{slog.LevelWarn, slog.LevelError},
    }); err != nil {
        panic(err)
    }
    defer lgsentry.Flush(2 * time.Second)

    // Create Fiber app
    app := fiber.New(fiber.Config{
        ErrorHandler: lgfiber.ErrorHandler,
    })

    // Add Sentry middleware (order matters!)
    app.Use(sentryfiber.New(sentryfiber.Options{
        Repanic:         true,
        WaitForDelivery: false,
        Timeout:         3 * time.Second,
    }))
    app.Use(lgfiber.PerformanceMiddleware())       // Creates transaction
    app.Use(lgfiber.TraceIDMiddleware())           // Extracts Sentry trace_id
    app.Use(lgfiber.ContextEnrichmentMiddleware())
    app.Use(lgfiber.BreadcrumbsMiddleware())
    app.Use(lgfiber.RecoverMiddleware)

    // Routes
    app.Get("/", func(c *fiber.Ctx) error {
        logbundle.InfoCtx(c.UserContext(), "Request received")
        return c.SendString("Hello, World!")
    })

    app.Listen(":3000")
}
```

## Configuration

### Environment Variables

```bash
# Logging
LOG_LEVEL=info                          # debug, info, warn, error

# Sentry
SENTRY_ENABLED=true
SENTRY_DSN=https://xxx@sentry.io/xxx
SENTRY_DEBUG=false
ENVIRONMENT=production
SERVER_NAME=api-server-01

# Sentry Performance
SENTRY_ENABLE_PERFORMANCE=true
SENTRY_TRACES_SAMPLE_RATE=0.1          # 10% of requests
SENTRY_SAMPLE_RATE=1.0                  # 100% of errors

# Sentry Privacy
SENTRY_SEND_DEFAULT_PII=false
```

### Logger Configuration

```go
type LoggerConfig struct {
    Level         slog.Level  // Minimum log level
    SentryEnabled bool        // Enable Sentry integration
    AddSource     bool        // Add source file/line to logs
}
```

### Sentry Configuration

```go
// Config embeds sentry.ClientOptions with additional logbundle-specific settings.
// All sentry.ClientOptions fields are available directly on this struct.
type Config struct {
    sentry.ClientOptions  // Embedded - all Sentry config fields available

    // FilterLevels specifies which slog levels should be sent to Sentry.
    // For example: []slog.Level{slog.LevelWarn, slog.LevelError}
    FilterLevels []slog.Level
}
```

**Key sentry.ClientOptions fields** (all available on Config):
- `Dsn` - Sentry DSN
- `Environment` - Environment name (production, staging, etc.)
- `Debug` - Enable debug mode
- `AttachStacktrace` - Attach stacktraces to messages
- `SampleRate` - Sample rate for errors (0.0-1.0)
- `EnableTracing` - Enable performance tracing
- `TracesSampleRate` - Sample rate for traces (0.0-1.0)
- `SendDefaultPII` - Send personally identifiable information
- `MaxBreadcrumbs` - Maximum number of breadcrumbs
- `BeforeSend` - Callback before sending events
- `BeforeBreadcrumb` - Callback before adding breadcrumbs
- `ServerName` - Server/instance identifier
- `Release` - Release version
- And many more - see [Sentry Go SDK docs](https://docs.sentry.io/platforms/go/)

## Logging

### Context-Aware Logging

Always use context-aware logging in HTTP handlers to include request context:

```go
app.Get("/user/:id", func(c *fiber.Ctx) error {
    userID := c.Params("id")

    // Use context-aware logging
    logbundle.InfoCtx(c.UserContext(), "Fetching user",
        slog.String("user_id", userID))

    user, err := getUserByID(userID)
    if err != nil {
        logbundle.ErrorCtx(c.UserContext(), "Failed to fetch user",
            slog.String("user_id", userID),
            logbundle.ErrAttr(err))
        return err
    }

    return c.JSON(user)
})
```

### Request Tracing with trace_id

The `TraceIDMiddleware` extracts Sentry's transaction `trace_id` and adds it to all your logs. This allows you to correlate logs with Sentry transactions and spans:

```go
// In your middleware setup (order matters!)
app.Use(sentryfiber.New(...))              // 1. Base Sentry
app.Use(lgfiber.PerformanceMiddleware())   // 2. Creates transaction
app.Use(lgfiber.TraceIDMiddleware())       // 3. Extracts trace_id

// In your handlers - trace_id is automatically added to logs
app.Get("/user/:id", func(c *fiber.Ctx) error {
    // This log will include log_trace_id=<sentry-trace-id>
    logbundle.InfoCtx(c.UserContext(), "Fetching user")

    // All subsequent logs will have the same trace_id
    user, err := fetchUser(c.UserContext())
    if err != nil {
        // This error log will have the same trace_id
        logbundle.ErrorCtx(c.UserContext(), "Failed to fetch user")
        return err
    }

    return c.JSON(user)
})
```

**Log Output Example:**
```
2025/01/30 12:34:56 [INFO] Fetching user log_trace_id=cd007055f39ac6383d750a071bc719aa
```

**Sentry Transaction:**
```
Trace ID: cd007055f39ac6383d750a071bc719aa
Span ID: cec9e91ab008c7c7
Parent Span ID: ef7bd4ef977face7
```

**Benefits:**

- Uses Sentry's actual trace_id (not a random UUID)
- Correlate logs with Sentry transactions/spans
- Track all logs from a single request
- Automatically added to Sentry events as a tag
- Makes debugging easier by following the complete request flow across logs and Sentry

### Log Levels

```go
logbundle.Debug("Debug message", slog.String("key", "value"))
logbundle.Info("Info message", slog.Int("count", 42))
logbundle.Warn("Warning message", slog.Bool("important", true))
logbundle.Error("Error message", logbundle.ErrAttr(err))

// Context-aware variants
logbundle.DebugCtx(ctx, "Debug with context")
logbundle.InfoCtx(ctx, "Info with context")
logbundle.WarnCtx(ctx, "Warn with context")
logbundle.ErrorCtx(ctx, "Error with context")
```

### Minimal Logging (without source)

For high-performance scenarios where source information is not needed:

```go
logbundle.InitLogMin(logbundle.LoggerConfig{
    Level:         slog.LevelInfo,
    AddSource:     false,
    SentryEnabled: false,
})

logbundle.InfoMin("Fast log message")
```

## Sentry Integration

### Automatic Error Tracking

All `Warn` and `Error` level logs are automatically sent to Sentry when enabled:

```go
// This will be sent to Sentry
logbundle.Error("Payment failed",
    slog.String("payment_id", "123"),
    slog.Float64("amount", 99.99),
    logbundle.ErrAttr(err))
```

### Manual Breadcrumbs

Add custom breadcrumbs for better error context:

```go
app.Post("/payment", func(c *fiber.Ctx) error {
    lgfiber.AddBreadcrumb(c, "payment", "Payment initiated",
        sentry.LevelInfo,
        map[string]any{
            "amount":   100.50,
            "currency": "USD",
        })

    // Process payment...

    lgfiber.AddBreadcrumb(c, "payment", "Payment validated",
        sentry.LevelInfo,
        map[string]any{"status": "valid"})

    return c.SendStatus(200)
})
```

### Custom Tags and Context

Enrich Sentry events with custom data:

```go
app.Use(func(c *fiber.Ctx) error {
    // Add custom tags
    lgfiber.SetTag(c, "tenant_id", tenantID)
    lgfiber.SetTag(c, "feature", "checkout")

    // Add custom context
    lgfiber.SetContext(c, "business_data", map[string]any{
        "cart_total": 299.99,
        "item_count": 3,
    })

    return c.Next()
})
```

### Custom Error Handling

Use the built-in error types for better categorization:

```go
import "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"

func getUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, erri.New().
            Type(erri.ErriStruct.DATABASE).
            Message("User not found").
            Details("Database query failed").
            Property("user_id").
            Value(id).
            SystemError(err).
            Build()
    }
    return user, nil
}
```

## Fiber Integration

### Middleware Order

**Order is critical!** Follow this sequence:

```go
app := fiber.New(fiber.Config{
    ErrorHandler: lgfiber.ErrorHandler,
})

// 1. Sentry base middleware (MUST BE FIRST)
app.Use(sentryfiber.New(sentryfiber.Options{
    Repanic:         true,
    WaitForDelivery: false,
    Timeout:         3 * time.Second,
}))

// 2. Performance monitoring (creates transaction with trace_id)
if os.Getenv("SENTRY_ENABLE_PERFORMANCE") == "true" {
    app.Use(lgfiber.PerformanceMiddleware())
}

// 3. Trace ID injection (extracts Sentry trace_id for logs)
// MUST be after PerformanceMiddleware to get the transaction trace_id
app.Use(lgfiber.TraceIDMiddleware())

// 4. Context enrichment (tags, request data, user info)
app.Use(lgfiber.ContextEnrichmentMiddleware())

// 5. Breadcrumbs (automatic request tracking)
app.Use(lgfiber.BreadcrumbsMiddleware())

// 6. Panic recovery (MUST BE AFTER SENTRY)
app.Use(lgfiber.RecoverMiddleware)

// 7. Your application middleware
app.Use(cors.New())
app.Use(logger.New())
```

### Error Handler

The global error handler automatically:
- Captures 5xx errors to Sentry
- Logs 4xx errors as warnings
- Includes full request context
- Adds proper fingerprinting

```go
// Errors are automatically handled
app.Get("/api/data", func(c *fiber.Ctx) error {
    if err := validateRequest(c); err != nil {
        return fiber.NewError(400, "Invalid request") // Logged as warning
    }

    data, err := fetchData()
    if err != nil {
        return err // Captured to Sentry as 500 error
    }

    return c.JSON(data)
})
```

## Performance Monitoring

### Automatic Transaction Tracking

When `EnablePerformance` is true, every request is tracked as a transaction:

```go
lgsentry.Init(&lgsentry.Config{
    EnablePerformance: true,
    TracesSampleRate:  0.1, // Sample 10% of requests
    // ...
})
```

### Custom Spans

Track specific operations within a request:

```go
app.Get("/report", func(c *fiber.Ctx) error {
    // Track database query
    span := lgfiber.StartSpan(c, "database.query", "Fetch report data")
    data, err := db.Query("SELECT * FROM reports")
    span.Finish()

    if err != nil {
        return err
    }

    // Track processing
    processSpan := lgfiber.StartSpan(c, "processing", "Generate report")
    report := generateReport(data)
    processSpan.Finish()

    return c.JSON(report)
})
```

### Nested Spans

```go
func processOrder(c *fiber.Ctx, order *Order) error {
    span := lgfiber.StartSpan(c, "business.process_order", "Process order")
    defer span.Finish()

    // Validate
    validateSpan := lgfiber.StartSpan(c, "validation", "Validate order")
    if err := validateOrder(order); err != nil {
        validateSpan.Finish()
        return err
    }
    validateSpan.Finish()

    // Payment
    paymentSpan := lgfiber.StartSpan(c, "payment.charge", "Charge payment")
    if err := chargePayment(order); err != nil {
        paymentSpan.Finish()
        return err
    }
    paymentSpan.Finish()

    return nil
}
```

## Best Practices

### 1. Always Use Context-Aware Logging in HTTP Handlers

```go
// ✅ Good - includes request context
logbundle.InfoCtx(c.UserContext(), "User logged in", slog.String("user_id", id))

// ❌ Bad - loses request context
logbundle.Info("User logged in", slog.String("user_id", id))
```

### 2. Use Trace IDs for Request Tracking

```go
// In middleware or early in request lifecycle
logbundle.LogTraceIDToFHCtx(c.RequestCtx())

// All subsequent logs will include the log_trace_id
logbundle.InfoCtx(c.UserContext(), "Processing request")
// Output: log_trace_id=abc123-def456 message="Processing request"
```

### 3. Structure Your Logs

```go
// ✅ Good - structured and searchable
logbundle.Info("Order created",
    slog.String("order_id", order.ID),
    slog.String("customer_id", order.CustomerID),
    slog.Float64("total", order.Total),
    slog.Int("items", len(order.Items)))

// ❌ Bad - unstructured
logbundle.Info(fmt.Sprintf("Order %s created for customer %s", order.ID, order.CustomerID))
```

### 4. Use Appropriate Log Levels

```go
// Debug - detailed information for debugging
logbundle.Debug("Cache hit", slog.String("key", key))

// Info - general informational messages
logbundle.Info("Request processed", slog.Int("duration_ms", 45))

// Warn - warning messages (still sent to Sentry)
logbundle.Warn("Rate limit approaching", slog.Int("remaining", 10))

// Error - error messages (sent to Sentry)
logbundle.Error("Failed to process payment", logbundle.ErrAttr(err))
```

### 5. Sample Traces in Production

Don't track 100% of requests in production:

```go
lgsentry.Init(&lgsentry.Config{
    EnablePerformance: true,
    TracesSampleRate:  0.05, // Only 5% in production
    SampleRate:        1.0,  // But capture all errors
})
```

### 6. Add Business Context

```go
app.Use(func(c *fiber.Ctx) error {
    // Extract and add business context
    if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
        lgfiber.SetTag(c, "tenant_id", tenantID)
    }

    if userID := getUserFromSession(c); userID != "" {
        lgfiber.SetTag(c, "user_id", userID)
    }

    return c.Next()
})
```

### 7. Disable WaitForDelivery in Production

```go
// ❌ Bad - blocks requests
sentryfiber.New(sentryfiber.Options{
    WaitForDelivery: true,
    Timeout:         20 * time.Second,
})

// ✅ Good - Sentry buffers events
sentryfiber.New(sentryfiber.Options{
    WaitForDelivery: false,
    Timeout:         3 * time.Second,
})
```

### 8. Use Custom Error Types

```go
import "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"

// Better error grouping in Sentry
return erri.New().
    Type(erri.ErriStruct.VALIDATION).
    Message("Invalid email format").
    Property("email").
    Value(email).
    Build()
```

### 9. Graceful Shutdown

Always flush Sentry events before shutdown:

```go
func main() {
    // Initialize Sentry
    lgsentry.Init(&lgsentry.Config{...})
    defer lgsentry.Flush(2 * time.Second) // Flush on exit

    // ... application code
}
```

### 10. Don't Log Sensitive Data

```go
// ❌ Bad - logs sensitive data
logbundle.Info("User authenticated",
    slog.String("password", password),
    slog.String("credit_card", card))

// ✅ Good - no sensitive data
logbundle.Info("User authenticated",
    slog.String("user_id", userID),
    slog.Bool("card_verified", true))
```

## API Reference

### Core Functions

```go
// Logger initialization
func InitLog(cfg LoggerConfig)
func InitLogMin(cfg LoggerConfig)

// Logging
func Debug(msg string, args ...any)
func Info(msg string, args ...any)
func Warn(msg string, args ...any)
func Error(msg string, args ...any)

// Context-aware logging
func DebugCtx(ctx context.Context, msg string, args ...any)
func InfoCtx(ctx context.Context, msg string, args ...any)
func WarnCtx(ctx context.Context, msg string, args ...any)
func ErrorCtx(ctx context.Context, msg string, args ...any)

// Utilities
func ErrAttr(err error) slog.Attr
func GetLogTraceID(ctx any) string
func LogTraceIDToFHCtx(ctx *fasthttp.RequestCtx)
```

### Sentry Functions

```go
// Initialization
func Init(config *Config) error
func Flush(timeout time.Duration)
```

### Fiber Functions

```go
// Middleware
func TraceIDMiddleware() fiber.Handler
func PerformanceMiddleware() fiber.Handler
func ContextEnrichmentMiddleware() fiber.Handler
func BreadcrumbsMiddleware() fiber.Handler
func RecoverMiddleware(c *fiber.Ctx) error

// Error handling
func ErrorHandler(c *fiber.Ctx, err error) error

// Utilities
func StartSpan(c *fiber.Ctx, operation, description string) *sentry.Span
func AddBreadcrumb(c *fiber.Ctx, category, message string, level sentry.Level, data map[string]any)
func SetTag(c *fiber.Ctx, key, value string)
func SetContext(c *fiber.Ctx, key string, value map[string]any)
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
