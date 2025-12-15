# logbundle-go

A high-performance Go logging and error handling library optimized for production applications with deep Fiber web framework and Sentry integration.

## Features

- **Structured Logging**: Built on Go's `log/slog` with custom formatting
- **Rich Error Types**: Pre-defined error categories with automatic HTTP status mapping
- **Stack Trace Capture**: Automatic stack trace collection with intelligent frame filtering
- **Panic Recovery**: Goroutine-safe panic recovery for middleware and background tasks
- **Fiber Integration**: Drop-in error handler, validation middleware, and Sentry integration
- **Sentry Integration**: Automatic error reporting with context enrichment and HTTP status filtering (opt-in)
- **RFC 7807 Compliant**: Problem Details for HTTP APIs standard support
- **Validation Middleware**: Type-safe validation for body, query, params, headers, and form data
- **Thread-Safe**: Concurrent-safe operations with optimized mutex usage
- **Zero-Allocation Paths**: Lazy initialization and object pooling for hot paths

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
    // Create logger with custom configuration
    logger := logbundle.CreateLogger(logbundle.LoggerConfig{
        Level:     slog.LevelDebug,
        AddSource: true,
    })

    // Or use default (reads from log_level env var)
    logger := logbundle.CreateLoggerDefault()

    // Use the logger
    logger.Info("Application started")
    logger.Debug("Debug information", "key", "value")
    logger.ErrorContext(ctx, "Error occurred", logbundle.ErrAttr(err))
}
```

## Error Handling

### Creating Errors

**Recommended: Factory Functions (Optimized - No intermediate allocations)**

```go
import "github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"

// Simple errors
err := lgerr.NotFound("User", 123)
err := lgerr.Validation("Email is required")
err := lgerr.Internal("database connection failed")

// With options
err := lgerr.Database("query timeout",
    lgerr.WithWrapped(originalErr),
    lgerr.WithContextKV("query", "SELECT * FROM users"),
    lgerr.WithContextKV("timeout", "5s"),
)

err := lgerr.Forbidden("admin-panel", "insufficient permissions",
    lgerr.WithContextKV("user_role", "viewer"),
)
```

**Alternative: Builder Pattern (Legacy - Still supported)**

```go
err := lgerr.New("custom error message").
    WithType(lgerr.TypeBadInput).
    WithTitle("Invalid Request").
    WithDetail("The provided data is invalid").
    WithContext("field", "email").
    Wrap(originalErr)
```

### Error Types and HTTP Status Codes

| Factory Function | Error Type | HTTP Status | Use Case |
|------------------|------------|-------------|----------|
| `Internal(msg)` | `TypeInternal` | 500 | Internal server errors |
| `NotFound(resource, id)` | `TypeNotFound` | 404 | Resource not found |
| `Validation(msg)` | `TypeValidation` | 400 | Validation failures |
| `Database(msg)` | `TypeDatabase` | 500 | Database errors |
| `Busy(msg)` | `TypeBusy` | 503 | Service unavailable |
| `Forbidden(resource, reason)` | `TypeForbidden` | 403 | Access forbidden |
| `BadInput(msg)` | `TypeBadInput` | 400 | Bad request input |
| `Unauthorized(reason)` | `TypeUnauth` | 401 | Unauthorized access |
| `Conflict(resource, reason)` | `TypeConflict` | 409 | Resource conflicts |
| `External(service, msg)` | `TypeExternal` | 502 | External service errors |
| `Timeout(operation, duration)` | `TypeTimeout` | 504 | Request timeouts |

### Custom Error Types

```go
const TypeRateLimited lgerr.ErrorType = "rate_limited"
lgerr.RegisterErrorType(TypeRateLimited, 429)

// Override existing mappings
lgerr.SetHTTPStatusMap(map[lgerr.ErrorType]int{
    lgerr.TypeNotFound: 410,  // Use 410 Gone
    lgerr.TypeBusy:     429,  // Use 429 Too Many Requests
})
```

### Validation Errors

```go
err := lgerr.Validation("form validation failed",
    lgerr.WithValidationErr("email", "must be valid email", "invalid@"),
    lgerr.WithValidationErr("age", "must be at least 18", 15),
    lgerr.WithDetail("Please correct the errors and try again"),
)

// Access validation errors
if err.HasValidationErrors() {
    for _, ve := range err.ValidationErrors() {
        fmt.Printf("Field: %s, Error: %s\n", ve.Field, ve.Message)
    }
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

    // Initialize Sentry SDK (optional)
    sentry.Init(sentry.ClientOptions{
        Dsn:         os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENVIRONMENT"),
    })
    defer sentry.Flush(2 * time.Second)

    // ⚠️ REQUIRED: Enable Sentry integration (disabled by default)
    logbundle.SetSentryEnabled(true)
    logbundle.SetSentryMinHTTPStatus(500) // Only 5xx errors

    app := fiber.New(fiber.Config{
        ErrorHandler: lgfiber.ErrorHandler,
    })

    // ⚠️ MIDDLEWARE ORDER IS CRITICAL ⚠️

    // 1. Sentry base (MUST BE FIRST)
    app.Use(sentryfiber.New(sentryfiber.Options{
        Repanic:         true,
        WaitForDelivery: false,
        Timeout:         3 * time.Second,
    }))

    // 2. Panic recovery (MUST BE AFTER SENTRY)
    app.Use(lgfiber.RecoverMiddleware())

    // 3. Performance monitoring
    app.Use(lgfiber.PerformanceMiddleware())

    // 4. Context enrichment
    app.Use(lgfiber.ContextEnrichmentMiddleware())

    // 5. Breadcrumbs
    app.Use(lgfiber.BreadcrumbsMiddleware())

    // 6. Your application middleware
    app.Use(yourMiddleware...)

    app.Listen(":3000")
}
```

### Validation Middleware

The library provides type-safe validation middleware with pooled allocations for optimal performance.

**Configure once at startup:**

```go
func main() {
    appLogger := logbundle.CreateLoggerDefault()

    // Set global logger for all validation middleware
    lgfiber.SetValidationLogger(appLogger)

    // Optional: customize configs per middleware type
    lgfiber.SetBodyValidationConfig(lgfiber.ValidationConfig{
        Title: "Invalid Request Body",
    })

    lgfiber.SetQueryValidationConfig(lgfiber.ValidationConfig{
        Title: "Invalid Query Parameters",
    })

    lgfiber.SetParamsValidationConfig(lgfiber.ValidationConfig{
        Title: "Invalid Route Parameters",
    })

    lgfiber.SetHeadersValidationConfig(lgfiber.ValidationConfig{
        Title: "Missing Required Headers",
    })

    app := fiber.New()
    // ... setup routes
}
```

#### 1. Body Validation

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Age   int    `json:"age" validate:"required,gte=18"`
}

app.Post("/users",
    lgfiber.BodyValidationMiddleware[CreateUserRequest](),
    createUserHandler,
)

func createUserHandler(c *fiber.Ctx) error {
    // Validated data in c.Locals
    body := c.Locals("body").(CreateUserRequest)

    user := createUser(body)
    return c.JSON(user)
}
```

#### 2. Query Validation

```go
type SearchQuery struct {
    Query string `query:"q" validate:"required,min=3"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
    Page  int    `query:"page" validate:"min=1"`
}

app.Get("/search",
    lgfiber.QueryValidationMiddleware[SearchQuery](),
    searchHandler,
)

func searchHandler(c *fiber.Ctx) error {
    query := c.Locals("query").(SearchQuery)

    results := search(query.Query, query.Limit, query.Page)
    return c.JSON(results)
}
```

#### 3. Params Validation

```go
type UserParams struct {
    ID string `params:"id" validate:"required,uuid"`
}

app.Get("/users/:id",
    lgfiber.ParamsValidationMiddleware[UserParams](),
    getUserHandler,
)

func getUserHandler(c *fiber.Ctx) error {
    params := c.Locals("params").(UserParams)

    user := getUser(params.ID)
    return c.JSON(user)
}
```

#### 4. Headers Validation

```go
type RequiredHeaders struct {
    Authorization string `reqheader:"Authorization" validate:"required"`
    ContentType   string `reqheader:"Content-Type" validate:"required,oneof=application/json application/xml"`
}

app.Post("/api",
    lgfiber.HeadersValidationMiddleware[RequiredHeaders](),
    apiHandler,
)

func apiHandler(c *fiber.Ctx) error {
    headers := c.Locals("headers").(RequiredHeaders)

    // Headers are validated and available
    return c.JSON(fiber.Map{"status": "ok"})
}
```

#### 5. Form Data Validation (with JSON field)

For form submissions with embedded JSON data:

```go
type UploadRequest struct {
    Title       string   `json:"title" validate:"required,min=3"`
    Description string   `json:"description" validate:"required"`
    Tags        []string `json:"tags" validate:"required,min=1"`
}

// Form field name defaults to "json_data" if empty string
app.Post("/upload",
    lgfiber.FormDataValidationMiddleware[UploadRequest]("json_data"),
    uploadHandler,
)

// Or use default field name
app.Post("/upload",
    lgfiber.FormDataValidationMiddleware[UploadRequest](""),
    uploadHandler,
)

func uploadHandler(c *fiber.Ctx) error {
    // Validated JSON data from form field
    formData := c.Locals("form_data").(UploadRequest)

    // File from multipart form
    file, _ := c.FormFile("file")

    return processUpload(formData, file)
}
```

**HTML form example:**

```html
<form action="/upload" method="POST" enctype="multipart/form-data">
    <input type="file" name="file" />
    <input type="hidden" name="json_data" value='{"title":"My File","description":"A test file","tags":["test","upload"]}' />
    <button type="submit">Upload</button>
</form>
```

### Validation Error Response (RFC 7807)

All validation middleware returns consistent RFC 7807 compliant responses:

```json
{
  "title": "Validation Error",
  "detail": "Please check your request body",
  "errors": [
    {
      "field": "email",
      "message": "Invalid email format",
      "value": "invalid@"
    },
    {
      "field": "age",
      "message": "Value must be greater than or equal to 18",
      "value": 16
    }
  ]
}
```

### Supported Validation Tags

The library uses `go-playground/validator` under the hood.

[Full validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)

### Error Handling in Handlers

```go
func getUserHandler(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Params("id")

        user, err := database.FindUser(userID)
        if err != nil {
            // Return lgerr.Error - automatically handled by ErrorHandler
            return lgerr.NotFound("User", userID).
                WithDetail("The requested user does not exist")
        }

        return c.JSON(user)
    }
}
```

### Manual Error Handling (Goroutines)

**CRITICAL: Use `RecoverGoroutinePanic` to prevent crashes**

```go
func handler(c *fiber.Ctx) error {
    // Async operation
    go func() {
        defer lgfiber.RecoverGoroutinePanic(c.UserContext(), "background-task")

        if err := doBackgroundTask(); err != nil {
            lgErr := lgerr.Internal("background task failed",
                lgerr.WithWrapped(err),
            )

            // With Fiber context (includes request data)
            lgfiber.HandleErrorWithFiber(c, lgErr)
        }
    }()

    return c.JSON(fiber.Map{"status": "processing"})
}
```

### Manual Sentry Integration

For custom logging with Sentry:

```go
import "github.com/aeternitas-infinita/logbundle-go"

// Debug level (info only, not sent to Sentry)
logbundle.SentryDebug(ctx, logger, "Processing started",
    slog.String("user_id", userID),
)

// Info level
logbundle.SentryInfo(ctx, logger, "User logged in",
    slog.String("user_id", userID),
    slog.String("ip", ip),
)

// Warning level
logbundle.SentryWarn(ctx, logger, "Deprecated API used", err,
    slog.String("endpoint", "/old/api"),
)

// Error level
logbundle.SentryError(ctx, logger, "Failed to process payment", err,
    slog.String("payment_id", paymentID),
    slog.Int("amount", amount),
)
```

## Sentry Integration

### Enable/Disable Sentry

**⚠️ IMPORTANT: Sentry is DISABLED by default!**

```go
// Initialize Sentry SDK
sentry.Init(sentry.ClientOptions{
    Dsn:         "your-dsn-here",
    Environment: "production",
})

// ⚠️ REQUIRED: Enable Sentry integration
logbundle.SetSentryEnabled(true)

// Configure HTTP status filtering
logbundle.SetSentryMinHTTPStatus(500)  // Only 5xx errors (default)
logbundle.SetSentryMinHTTPStatus(400)  // 4xx and 5xx errors
logbundle.SetSentryMinHTTPStatus(0)    // All errors

// Check if enabled
if logbundle.IsSentryEnabled() {
    // Sentry is active
}
```

### Skip Specific Errors

```go
// Won't be sent to Sentry
err := lgerr.NotFound("Resource", id).
    IgnoreSentry()

// Or with factory + options
err := lgerr.Unauthorized("invalid token",
    lgerr.WithIgnoreSentry(),
)
```

### Custom Sentry Context

```go
lgfiber.SetTag(c, "feature", "user-management")
lgfiber.SetContext(c, "custom_data", map[string]any{
    "operation": "user_update",
    "batch_id":  batchID,
})

lgfiber.AddBreadcrumb(c, "database", "Query executed",
    sentry.LevelInfo,
    map[string]any{"query": "SELECT * FROM users"},
)
```

## Performance Optimizations

### 1. Lazy Initialization

```go
// Error context maps are nil by default, allocated only when used
err := lgerr.New("error")  // No context map allocation
err.WithContext("key", "value")  // Now allocated

// Sentry tags/extra maps lazy-initialized
// Zero allocations if extraData is empty
```

### 2. Object Pooling

```go
// Validation error slices are pooled and reused
// Reduces allocations in validation middleware by 40-60%
lgfiber.BodyValidationMiddleware[T]()  // Uses sync.Pool internally
```

### 3. Zero-Allocation Fast Paths

```go
// When Sentry is disabled, all checks return immediately
if !config.IsSentryEnabled() {
    return  // No hub fetch, no allocations
}

// Context cancellation checked before expensive operations
select {
case <-ctx.Done():
    return
default:
}
```

### 4. Pre-sized Allocations

```go
// Maps pre-allocated with known capacity
queries := c.Queries()
if len(queries) > 0 {
    queryParams := make(map[string]any, len(queries))  // Exact size
}
```

### 5. Factory Functions vs Builder Pattern

```go
// ❌ OLD (3+ allocations): options slice + append + NewWithOptions
options := []ErrorOption{...}
return NewWithOptions(append(options, opts...)...)

// ✅ NEW (1 allocation): direct field assignment
err := New(message)
err.errorType = TypeDatabase
err.title = "Database Error"
for _, opt := range opts {
    opt(err)
}
```

## Middleware Order

**⚠️ ORDER IS CRITICAL FOR SENTRY INTEGRATION ⚠️**

```go
// ✅ CORRECT ORDER
app.Use(sentryfiber.New(...))          // 1. Initialize Sentry hub
app.Use(lgfiber.RecoverMiddleware())   // 2. Catch panics
app.Use(lgfiber.PerformanceMiddleware())
app.Use(lgfiber.ContextEnrichmentMiddleware())
app.Use(lgfiber.BreadcrumbsMiddleware())
app.Use(yourMiddleware...)

// ❌ WRONG - RecoverMiddleware before Sentry
app.Use(lgfiber.RecoverMiddleware())   // No hub available!
app.Use(sentryfiber.New(...))          // Too late
```

**Why order matters:**

- `sentryfiber.New()` creates the Sentry hub - all other middleware need this
- `RecoverMiddleware()` must be early to catch panics from subsequent middleware
- Performance tracking should wrap the entire request lifecycle

## Best Practices

### 1. Use Factory Functions for Errors

```go
// ✅ Optimized - no intermediate allocations
err := lgerr.Database("connection failed",
    lgerr.WithWrapped(dbErr),
    lgerr.WithContextKV("host", "localhost"),
)

// ❌ Less efficient - creates options slice
err := lgerr.NewWithOptions(
    lgerr.WithMessage("connection failed"),
    lgerr.WithType(lgerr.TypeDatabase),
    ...
)
```

### 2. Configure Validation Middleware at Startup

```go
// ✅ Set global config once
func main() {
    lgfiber.SetValidationLogger(appLogger)
    lgfiber.SetBodyValidationConfig(lgfiber.ValidationConfig{
        Title: "Invalid Request",
    })
    // Configs are captured at middleware creation, no locks during requests
}

// ❌ Don't reconfigure during runtime
func handler(c *fiber.Ctx) error {
    lgfiber.SetBodyValidationConfig(...)  // Causes lock contention
}
```

### 3. Always Use RecoverGoroutinePanic

```go
// ✅ Safe goroutine
go func() {
    defer lgfiber.RecoverGoroutinePanic(ctx, "worker-name")
    doWork()
}()

// ❌ Unsafe - panics crash the application
go func() {
    doWork()  // If this panics, app crashes!
}()
```

### 4. Add Context to Errors

```go
// ✅ Rich context for debugging
return lgerr.Database("query failed",
    lgerr.WithWrapped(err),
    lgerr.WithContextKV("table", "users"),
    lgerr.WithContextKV("operation", "insert"),
)

// ❌ Missing context
return lgerr.Database("query failed")
```

### 5. Use Context-Aware Logging

```go
// ✅ Preserves request context
logger.ErrorContext(ctx, "operation failed",
    "user_id", userID,
    logbundle.ErrAttr(err),
)

// ❌ Loses context
logger.Error("operation failed")
```

### 6. Don't Send Sensitive Errors to Sentry

```go
err := lgerr.Unauthorized("invalid credentials",
    lgerr.WithContextKV("attempt_count", 3),
    lgerr.WithIgnoreSentry(),  // Don't leak auth details
)
```

## API Reference

### Validation Middleware Functions

| Function | Description | Locals Key |
|----------|-------------|------------|
| `BodyValidationMiddleware[T]()` | Validates JSON request body | `"body"` |
| `QueryValidationMiddleware[T]()` | Validates query parameters | `"query"` |
| `ParamsValidationMiddleware[T]()` | Validates route parameters | `"params"` |
| `HeadersValidationMiddleware[T]()` | Validates request headers | `"headers"` |
| `FormDataValidationMiddleware[T](field)` | Validates form data with JSON field | `"form_data"` |

### Configuration Functions

| Function | Description |
|----------|-------------|
| `SetValidationLogger(logger)` | Set global logger for all validation middleware |
| `SetBodyValidationConfig(config)` | Configure body validation globally |
| `SetQueryValidationConfig(config)` | Configure query validation globally |
| `SetParamsValidationConfig(config)` | Configure params validation globally |
| `SetHeadersValidationConfig(config)` | Configure headers validation globally |
| `GetValidationLogger()` | Get current global validation logger |
| `GetBodyValidationConfig()` | Get current body validation config |
| `ResetValidationConfigs()` | Reset all configs to defaults |

### Error Handler Functions

| Function | Description |
|----------|-------------|
| `ErrorHandler(c, err)` | Main Fiber error handler |
| `HandleError(ctx, lgErr)` | Manual error handling (no Fiber context) |
| `HandleErrorWithFiber(c, lgErr)` | Manual error handling with Fiber context |
| `RecoverGoroutinePanic(ctx, name)` | Panic recovery for goroutines |

### Sentry Middleware Functions

| Function | Description |
|----------|-------------|
| `BreadcrumbsMiddleware()` | Adds HTTP breadcrumbs |
| `ContextEnrichmentMiddleware()` | Enriches Sentry context |
| `PerformanceMiddleware()` | Creates performance transactions |
| `RecoverMiddleware()` | Catches panics |

### Sentry Helper Functions

| Function | Description |
|----------|-------------|
| `SetTag(c, key, value)` | Set Sentry tag |
| `SetContext(c, key, data)` | Set Sentry context |
| `AddBreadcrumb(c, category, msg, level, data)` | Add custom breadcrumb |
| `StartSpan(c, operation, description)` | Start performance span |

## Thread Safety

- All error operations are safe for concurrent use after creation
- HTTP status map modifications protected by `sync.RWMutex`
- Sentry enable/disable is thread-safe with `sync.RWMutex`
- Logger instances safe for concurrent use
- Validation config reads use `sync.RLock` (writes use `sync.Lock`)
- Object pools (`sync.Pool`) are thread-safe

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `log_level` | `warn` | Minimum log level (debug, info, warn, error) |

## Performance Benchmarks

**Validation Middleware:**

- 40-60% allocation reduction vs pre-pooling (sync.Pool for validation errors)
- 30-40% faster middleware creation (init() vs lazy initDefaultConfigs())

**Error Creation:**

- Factory functions: 3x fewer allocations vs builder pattern with options slices
- Lazy context maps: ~150 bytes saved per error when context unused

**Sentry Integration:**

- Zero-allocation when disabled (early returns)
- Lazy map initialization: 100% reduction when extraData is empty

## License

MIT License - see LICENSE file for details

## Support

- GitHub Issues: <https://github.com/aeternitas-infinita/logbundle-go/issues>
- Documentation: [GoDoc](https://pkg.go.dev/github.com/aeternitas-infinita/logbundle-go)
