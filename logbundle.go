// Package logbundle provides comprehensive logging solution for Go applications
// that seamlessly integrates log/slog, Sentry error tracking, and Fiber web framework.
//
// Features:
//   - Structured logging with log/slog
//   - Deep Sentry integration with automatic error tracking
//   - Fiber middleware with request context enrichment
//   - Performance monitoring with Sentry Transactions and Spans
//   - Automatic breadcrumbs for error investigation
//   - Trace ID propagation across services
//
// Basic Usage:
//
//	logbundle.InitLog(logbundle.LoggerConfig{
//	    Level:         slog.LevelInfo,
//	    AddSource:     true,
//	    SentryEnabled: false,
//	})
//
//	logbundle.Info("Application started")
//	logbundle.Error("Error occurred", logbundle.ErrAttr(err))
//
// With Fiber and Sentry:
//
//	app := fiber.New(fiber.Config{
//	    ErrorHandler: lgfiber.ErrorHandler,
//	})
//
//	app.Use(sentryfiber.New(...))
//	app.Use(lgfiber.PerformanceMiddleware())
//	app.Use(lgfiber.TraceIDMiddleware())
//	app.Use(lgfiber.ContextEnrichmentMiddleware())
//	app.Use(lgfiber.BreadcrumbsMiddleware())
//	app.Use(lgfiber.RecoverMiddleware)
//
// For more examples, see the example/ directory.
package logbundle
