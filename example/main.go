package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aeternitas-infinita/logbundle-go"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgfiber"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Choose example to run
	runBasicLoggingExample()
	// runFiberWithSentryExample()
}

// Example 1: Basic Logging
func runBasicLoggingExample() {
	fmt.Println("=== Basic Logging Example ===")

	// Initialize logger with source info enabled
	logbundle.InitLog(logbundle.LoggerConfig{
		Level:         slog.LevelDebug,
		SentryEnabled: false,
		AddSource:     true,
	})

	// Basic logging at different levels
	logbundle.Debug("This is a debug message")
	logbundle.Info("This is an info message", slog.String("key", "value"))
	logbundle.Warn("This is a warning message", slog.Int("code", 100))
	logbundle.Error("This is an error message", slog.Int("code", 500))

	// Structured logging with multiple fields
	logbundle.Info("User logged in",
		slog.String("user_id", "12345"),
		slog.String("username", "john_doe"),
		slog.String("ip", "192.168.1.1"),
		slog.Duration("session_duration", 2*time.Hour),
	)

	// Logging errors
	err := errors.New("database connection failed")
	logbundle.Error("Failed to connect to database",
		logbundle.ErrAttr(err),
		slog.String("host", "localhost"),
		slog.Int("port", 5432),
	)

	// Context-aware logging
	ctx := context.Background()
	logbundle.InfoCtx(ctx, "Processing background job",
		slog.String("job_id", "job-123"),
		slog.String("type", "email"))

	// Call from another function
	demonstrateStructuredLogging()
}

// Example 2: Fiber Application with Full Sentry Integration
func runFiberWithSentryExample() {
	fmt.Println("=== Fiber with Sentry Example ===")

	// Set environment variables for this example
	os.Setenv("SENTRY_DSN", "https://your-dsn@sentry.io/project-id")
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("SENTRY_ENABLE_PERFORMANCE", "true")
	os.Setenv("SENTRY_DEBUG", "false")

	// Initialize logger with Sentry enabled
	logbundle.InitLog(logbundle.LoggerConfig{
		Level:         slog.LevelDebug,
		SentryEnabled: true,
		AddSource:     true,
	})

	// Initialize Sentry with full configuration
	// Config embeds sentry.ClientOptions - all Sentry fields available directly
	if err := lgsentry.Init(&lgsentry.Config{
		ClientOptions: sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			Environment:      os.Getenv("ENVIRONMENT"),
			Debug:            false,
			ServerName:       "example-server",
			AttachStacktrace: true,
			SampleRate:       1.0,
			EnableTracing:    true,
			TracesSampleRate: 1.0, // 100% for demo purposes
			MaxBreadcrumbs:   100,
		},
		// Custom logbundle fields
		FilterLevels: []slog.Level{slog.LevelWarn, slog.LevelError},
	}); err != nil {
		logbundle.Error("Failed to initialize Sentry", logbundle.ErrAttr(err))
		return
	}
	defer lgsentry.Flush(2 * time.Second)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: lgfiber.ErrorHandler,
	})

	// Setup middleware in correct order
	setupMiddleware(app)

	// Setup routes with examples
	setupRoutes(app)

	logbundle.Info("Server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		logbundle.Error("Server failed to start", logbundle.ErrAttr(err))
	}
}

func setupMiddleware(app *fiber.App) {
	// 1. Sentry base middleware (MUST BE FIRST)
	app.Use(sentryfiber.New(sentryfiber.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         3 * time.Second,
	}))

	// 2. Performance monitoring
	app.Use(lgfiber.PerformanceMiddleware())

	// 3. Context enrichment
	app.Use(lgfiber.ContextEnrichmentMiddleware())

	// 4. Breadcrumbs
	app.Use(lgfiber.BreadcrumbsMiddleware())

	// 5. Panic recovery
	app.Use(lgfiber.RecoverMiddleware)

	// 6. Application middleware
	app.Use(cors.New())

	// 7. Custom middleware - add business context
	app.Use(func(c *fiber.Ctx) error {
		// Add custom tags for all requests
		lgfiber.SetTag(c, "app_version", "1.0.0")

		// Simulate tenant ID from header
		if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
			lgfiber.SetTag(c, "tenant_id", tenantID)
		}

		return c.Next()
	})
}

func setupRoutes(app *fiber.App) {
	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		logbundle.InfoCtx(c.UserContext(), "Root endpoint accessed")
		return c.JSON(fiber.Map{
			"message": "LogBundle-Go Example API",
			"version": "1.0.0",
		})
	})

	// Example: Successful request with logging
	app.Get("/users/:id", func(c *fiber.Ctx) error {
		userID := c.Params("id")

		logbundle.InfoCtx(c.UserContext(), "Fetching user",
			slog.String("user_id", userID))

		// Add custom breadcrumb
		lgfiber.AddBreadcrumb(c, "user", "User fetch initiated",
			sentry.LevelInfo,
			map[string]any{"user_id": userID})

		// Simulate database query with span
		span := lgfiber.StartSpan(c, "database.query", "SELECT user by ID")
		time.Sleep(50 * time.Millisecond) // Simulate query time
		span.Finish()

		user := fiber.Map{
			"id":   userID,
			"name": "John Doe",
			"email": "john@example.com",
		}

		lgfiber.AddBreadcrumb(c, "user", "User fetched successfully",
			sentry.LevelInfo,
			map[string]any{"user_id": userID})

		return c.JSON(user)
	})

	// Example: Error handling
	app.Get("/error", func(c *fiber.Ctx) error {
		logbundle.WarnCtx(c.UserContext(), "Error endpoint accessed - this will fail")

		// This error will be captured by Sentry
		return errors.New("intentional error for testing")
	})

	// Example: Custom error with erri
	app.Get("/custom-error", func(c *fiber.Ctx) error {
		err := erri.New().
			Type(erri.ErriStruct.VALIDATION).
			Message("Invalid user ID").
			Details("User ID must be numeric").
			Property("user_id").
			Value("abc123").
			Build()

		logbundle.ErrorCtx(c.UserContext(), "Validation failed",
			logbundle.ErrAttr(err))

		return err
	})

	// Example: Panic (will be recovered by middleware)
	app.Get("/panic", func(c *fiber.Ctx) error {
		logbundle.WarnCtx(c.UserContext(), "Panic endpoint accessed - will panic")
		panic("intentional panic for testing")
	})

	// Example: Complex operation with multiple spans
	app.Post("/order", func(c *fiber.Ctx) error {
		orderID := "order-" + time.Now().Format("20060102150405")

		logbundle.InfoCtx(c.UserContext(), "Processing order",
			slog.String("order_id", orderID))

		// Main operation span
		span := lgfiber.StartSpan(c, "business.process_order", "Process order")

		// Step 1: Validate
		validateSpan := lgfiber.StartSpan(c, "validation", "Validate order")
		time.Sleep(20 * time.Millisecond)
		validateSpan.Finish()
		lgfiber.AddBreadcrumb(c, "order", "Order validated", sentry.LevelInfo, nil)

		// Step 2: Check inventory
		inventorySpan := lgfiber.StartSpan(c, "inventory.check", "Check inventory")
		time.Sleep(30 * time.Millisecond)
		inventorySpan.Finish()
		lgfiber.AddBreadcrumb(c, "order", "Inventory checked", sentry.LevelInfo, nil)

		// Step 3: Process payment
		paymentSpan := lgfiber.StartSpan(c, "payment.charge", "Charge payment")
		time.Sleep(100 * time.Millisecond)
		paymentSpan.Finish()
		lgfiber.AddBreadcrumb(c, "order", "Payment processed", sentry.LevelInfo, nil)

		span.Finish()

		logbundle.InfoCtx(c.UserContext(), "Order processed successfully",
			slog.String("order_id", orderID),
			slog.Duration("total_time", 150*time.Millisecond))

		return c.JSON(fiber.Map{
			"order_id": orderID,
			"status":   "completed",
		})
	})

	// Example: Demonstrating different log levels
	app.Get("/logs", func(c *fiber.Ctx) error {
		logbundle.DebugCtx(c.UserContext(), "Debug log - detailed info",
			slog.String("cache_key", "users:123"))

		logbundle.InfoCtx(c.UserContext(), "Info log - normal operation",
			slog.Int("records_processed", 42))

		logbundle.WarnCtx(c.UserContext(), "Warn log - potential issue",
			slog.Int("retry_count", 3))

		// This will be sent to Sentry
		logbundle.ErrorCtx(c.UserContext(), "Error log - something went wrong",
			slog.String("operation", "data_sync"),
			logbundle.ErrAttr(errors.New("sync timeout")))

		return c.SendString("Check logs and Sentry")
	})

	// Example: Health check (no logging clutter)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
}

func demonstrateStructuredLogging() {
	fmt.Println("\n=== Structured Logging Examples ===")

	// Business event logging
	logbundle.Info("Order created",
		slog.String("order_id", "ORD-12345"),
		slog.String("customer_id", "CUST-789"),
		slog.Float64("total", 299.99),
		slog.Int("items", 5),
		slog.String("status", "pending"))

	// Performance logging
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(start)

	logbundle.Info("API call completed",
		slog.String("endpoint", "/api/v1/users"),
		slog.Int("status_code", 200),
		slog.Duration("duration", duration),
		slog.Int("response_size", 1024))

	// Error with context
	err := simulateDatabaseError()
	if err != nil {
		logbundle.Error("Database operation failed",
			logbundle.ErrAttr(err),
			slog.String("operation", "INSERT"),
			slog.String("table", "users"),
			slog.String("query_id", "q-456"))
	}

	// Using custom error type
	customErr := erri.New().
		Type(erri.ErriStruct.DATABASE).
		Message("Failed to create user").
		Details("Duplicate email address").
		Property("email").
		Value("john@example.com").
		SystemError(errors.New("UNIQUE constraint failed")).
		Build()

	logbundle.Error("User creation failed",
		logbundle.ErrAttr(customErr),
		slog.String("attempt", "1"))
}

func simulateDatabaseError() error {
	return errors.New("connection timeout after 30s")
}

// Example: Using trace IDs
func demonstrateTraceID() {
	ctx, cancel := logbundle.CtxWithLogTraceID(context.Background(), 5*time.Second)
	defer cancel()

	traceID := logbundle.GetLogTraceID(ctx)
	fmt.Printf("Trace ID: %s\n", traceID)

	// All logs with this context will have the same log_trace_id
	logbundle.InfoCtx(ctx, "Step 1 - Starting process")
	logbundle.InfoCtx(ctx, "Step 2 - Processing data")
	logbundle.InfoCtx(ctx, "Step 3 - Completing process")
}
