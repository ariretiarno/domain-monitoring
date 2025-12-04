package alert

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/domain-expiration-monitor/dem/internal/repository"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: domain-expiration-monitor, Property 8: Alert threshold triggering
// For any domain with an expiration date and any configured alert threshold,
// when the time until expiration is less than or equal to the threshold, an alert should be generated.
// Validates: Requirements 5.1
func TestProperty_AlertThresholdTriggering(t *testing.T) {
	dbPath := "test_alert_threshold.db"
	defer os.Remove(dbPath)

	db, err := repository.NewDB(dbPath, "sqlite3")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	alertRepo := repository.NewAlertRepository(db)
	configRepo := repository.NewConfigRepository(db)
	service := NewService(alertRepo, configRepo)

	properties := gopter.NewProperties(nil)

	properties.Property("alert generated when within threshold", prop.ForAll(
		func(daysUntilExpiration uint8, thresholdDays uint8) bool {
			if daysUntilExpiration == 0 || thresholdDays == 0 {
				return true
			}

			// Create domain expiring in daysUntilExpiration days
			d := &domain.Domain{
				ID:             time.Now().Format("20060102150405.000000"),
				Name:           "test" + time.Now().Format("150405.000000") + ".com",
				ExpirationDate: time.Now().Add(time.Duration(daysUntilExpiration) * 24 * time.Hour),
				Nameservers:    domain.Strings{"ns1.example.com"},
				Registrant:     "Test",
				Registrar:      "Test",
				LastChecked:    time.Now(),
				NextCheck:      time.Now().Add(24 * time.Hour),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			// Set threshold
			threshold := time.Duration(thresholdDays) * 24 * time.Hour
			config, _ := configRepo.Get()
			config.SetAlertThresholds([]time.Duration{threshold})
			configRepo.Update(config)

			// Evaluate alerts
			err := service.EvaluateAlerts(d)
			if err != nil && !strings.Contains(err.Error(), "no webhook") {
				return false
			}

			// Check if alert was created when it should be
			alerts, _ := alertRepo.GetByDomainID(d.ID)

			if daysUntilExpiration <= thresholdDays {
				// Should have created an alert
				return len(alerts) > 0
			} else {
				// Should not have created an alert
				return len(alerts) == 0
			}
		},
		gen.UInt8Range(1, 100),
		gen.UInt8Range(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: domain-expiration-monitor, Property 9: Alert message completeness
// For any generated alert, the formatted message should contain the domain name,
// expiration date, and time remaining until expiration.
// Validates: Requirements 5.2, 5.3, 5.4
func TestProperty_AlertMessageCompleteness(t *testing.T) {
	service := NewService(nil, nil)
	properties := gopter.NewProperties(nil)

	properties.Property("alert message contains all required fields", prop.ForAll(
		func(domainName string, daysUntilExpiration uint8) bool {
			if daysUntilExpiration == 0 {
				return true
			}

			alert := &domain.Alert{
				DomainName:     domainName + ".com",
				ExpirationDate: time.Now().Add(time.Duration(daysUntilExpiration) * 24 * time.Hour),
				SentAt:         time.Now(),
			}
			alert.SetThreshold(30 * 24 * time.Hour)

			message := service.FormatAlertMessage(alert)

			// Check that message contains all required fields
			return strings.Contains(message, alert.DomainName) &&
				strings.Contains(message, alert.ExpirationDate.Format("2006-01-02")) &&
				strings.Contains(message, "Days Remaining:")
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.UInt8Range(1, 250),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: domain-expiration-monitor, Property 10: Alert deduplication
// For any domain and alert threshold combination, only one alert should be sent
// when that threshold is crossed, and subsequent evaluations should not generate duplicate alerts.
// Validates: Requirements 6.4, 6.5
func TestProperty_AlertDeduplication(t *testing.T) {
	dbPath := "test_alert_dedup.db"
	defer os.Remove(dbPath)

	db, err := repository.NewDB(dbPath, "sqlite3")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	alertRepo := repository.NewAlertRepository(db)
	configRepo := repository.NewConfigRepository(db)
	service := NewService(alertRepo, configRepo)

	properties := gopter.NewProperties(nil)

	testCounter := 0
	properties.Property("no duplicate alerts for same threshold", prop.ForAll(
		func(daysUntilExpiration uint8) bool {
			if daysUntilExpiration == 0 || daysUntilExpiration > 50 {
				return true
			}

			testCounter++
			d := &domain.Domain{
				ID:             time.Now().Format("20060102150405") + string(rune(testCounter)),
				Name:           "test" + string(rune(testCounter)) + ".com",
				ExpirationDate: time.Now().Add(time.Duration(daysUntilExpiration) * 24 * time.Hour),
				Nameservers:    domain.Strings{"ns1.example.com"},
				Registrant:     "Test",
				Registrar:      "Test",
				LastChecked:    time.Now(),
				NextCheck:      time.Now().Add(24 * time.Hour),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			// Set threshold higher than days until expiration so alert will trigger
			threshold := time.Duration(daysUntilExpiration+10) * 24 * time.Hour
			config, _ := configRepo.Get()
			config.SetAlertThresholds([]time.Duration{threshold})
			configRepo.Update(config)

			// Evaluate alerts twice
			service.EvaluateAlerts(d)
			service.EvaluateAlerts(d)

			// Should only have one alert (deduplication works)
			alerts, _ := alertRepo.GetByDomainID(d.ID)
			return len(alerts) == 1
		},
		gen.UInt8Range(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
