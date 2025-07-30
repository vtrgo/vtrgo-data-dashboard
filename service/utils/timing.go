// file: service/utils/timing.go
package utils

import (
	"strconv"
	"time"
	"vtarchitect/config"
)

// GetPollInterval retrieves the polling interval from the configuration.
func GetPollInterval(cfg *config.Config) time.Duration {
	pollInterval := cfg.Values["PLC_POLL_MS"]
	pollIntervalMs, err := strconv.Atoi(pollInterval)
	if err != nil || pollIntervalMs <= 0 {
		pollIntervalMs = 1000 // default to 1 second
	}
	return time.Duration(pollIntervalMs) * time.Millisecond
}

// GetFullWriteInterval retrieves the full-state write interval from the configuration (in minutes).
func GetFullWriteInterval(cfg *config.Config) time.Duration {
	intervalStr := cfg.Values["FULL_WRITE_MINUTES"]
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil || intervalMin <= 0 {
		intervalMin = 60 // default to 60 minutes
	}
	return time.Duration(intervalMin) * time.Minute
}
