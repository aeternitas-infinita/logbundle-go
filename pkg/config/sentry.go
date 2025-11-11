package config

import (
	"sync"
)

var (
	// sentryEnabled controls whether Sentry integration is active
	// Default: false (disabled)
	sentryEnabled   bool = false
	sentryEnabledMu sync.RWMutex
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
