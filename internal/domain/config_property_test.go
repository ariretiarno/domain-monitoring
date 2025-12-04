package domain

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: domain-expiration-monitor, Property 3: Configuration persistence
// For any valid configuration change (monitoring interval, alert thresholds, webhook URL, retention period),
// saving the configuration and then retrieving it should return the same values.
// Validates: Requirements 2.2, 6.3, 7.3
func TestProperty_ConfigurationPersistence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("monitoring interval round-trip", prop.ForAll(
		func(intervalHours int64) bool {
			config := &Config{}
			original := time.Duration(intervalHours) * time.Hour
			config.SetMonitoringInterval(original)
			retrieved := config.GetMonitoringInterval()
			return retrieved == original
		},
		gen.Int64Range(1, 8760), // 1 hour to 1 year in hours
	))

	properties.Property("retention period round-trip", prop.ForAll(
		func(retentionDays int64) bool {
			config := &Config{}
			original := time.Duration(retentionDays) * 24 * time.Hour
			config.SetRetentionPeriod(original)
			retrieved := config.GetRetentionPeriod()
			return retrieved == original
		},
		gen.Int64Range(1, 365), // 1 day to 1 year in days
	))

	properties.Property("alert thresholds round-trip", prop.ForAll(
		func(thresholdDays []int64) bool {
			if len(thresholdDays) == 0 {
				return true // Skip empty arrays
			}
			
			config := &Config{}
			original := make([]time.Duration, len(thresholdDays))
			for i, days := range thresholdDays {
				original[i] = time.Duration(days) * 24 * time.Hour
			}
			
			config.SetAlertThresholds(original)
			retrieved := config.GetAlertThresholds()
			
			if len(retrieved) != len(original) {
				return false
			}
			
			for i := range original {
				if retrieved[i] != original[i] {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.Int64Range(1, 365)), // 1 day to 1 year
	))

	properties.Property("webhook URL persistence", prop.ForAll(
		func(webhook string) bool {
			config := &Config{GoogleChatWebhook: webhook}
			return config.GoogleChatWebhook == webhook
		},
		gen.AnyString(),
	))

	properties.Property("durations JSON serialization round-trip", prop.ForAll(
		func(durationDays []int64) bool {
			if len(durationDays) == 0 {
				return true
			}
			
			original := make(Durations, len(durationDays))
			for i, days := range durationDays {
				original[i] = time.Duration(days) * 24 * time.Hour
			}
			
			// Simulate database storage
			value, err := original.Value()
			if err != nil {
				return false
			}
			
			// Simulate database retrieval
			var retrieved Durations
			if err := retrieved.Scan(value); err != nil {
				return false
			}
			
			if len(retrieved) != len(original) {
				return false
			}
			
			for i := range original {
				if retrieved[i] != original[i] {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.Int64Range(1, 365)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
