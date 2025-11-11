package config

import (
	"sync"
)

var (
	// sentryEnabled controls whether Sentry integration is active
	// Default: false (disabled)
	sentryEnabled   bool = false
	sentryEnabledMu sync.RWMutex

	// sentryMinHTTPStatus defines the minimum HTTP status code to send to Sentry
	// Default: 500 (only server errors)
	// Set to 400 to include client errors, or 0 to send all errors
	sentryMinHTTPStatus   int = 500
	sentryMinHTTPStatusMu sync.RWMutex
)

// IsSentryEnabled returns whether Sentry integration is currently enabled
func IsSentryEnabled() bool {
	sentryEnabledMu.RLock()
	defer sentryEnabledMu.RUnlock()
	return sentryEnabled
}

// SetSentryEnabled enables or disables Sentry integration globally
// When disabled, no events will be sent to Sentry from any part of the library
func SetSentryEnabled(enabled bool) {
	sentryEnabledMu.Lock()
	defer sentryEnabledMu.Unlock()
	sentryEnabled = enabled
}

// GetSentryMinHTTPStatus returns the minimum HTTP status code to send to Sentry
func GetSentryMinHTTPStatus() int {
	sentryMinHTTPStatusMu.RLock()
	defer sentryMinHTTPStatusMu.RUnlock()
	return sentryMinHTTPStatus
}

// SetSentryMinHTTPStatus sets the minimum HTTP status code to send to Sentry
// Examples:
//   - 500: Only server errors (5xx) - default
//   - 400: Client and server errors (4xx and 5xx)
//   - 0: All errors regardless of status code
func SetSentryMinHTTPStatus(minStatus int) {
	sentryMinHTTPStatusMu.Lock()
	defer sentryMinHTTPStatusMu.Unlock()
	sentryMinHTTPStatus = minStatus
}
