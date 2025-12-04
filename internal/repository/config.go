package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
)

// ConfigRepository handles configuration data persistence
type ConfigRepository struct {
	db *DB
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Get retrieves the application configuration
// If no configuration exists, it creates and returns default values
func (r *ConfigRepository) Get() (*domain.Config, error) {
	var config domain.Config
	query := `
		SELECT id, monitoring_interval, alert_thresholds, google_chat_webhook,
		       retention_period, updated_at
		FROM config
		WHERE id = 1
	`

	err := r.db.Get(&config, query)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default configuration
			defaultConfig := r.createDefaultConfig()
			if err := r.create(defaultConfig); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return &config, nil
}

// Update updates the application configuration
func (r *ConfigRepository) Update(config *domain.Config) error {
	config.ID = 1 // Ensure we're always updating the single config row
	config.UpdatedAt = time.Now()

	query := `
		UPDATE config
		SET monitoring_interval = ?, alert_thresholds = ?, google_chat_webhook = ?,
		    retention_period = ?, updated_at = ?
		WHERE id = 1
	`

	result, err := r.db.Exec(query,
		config.MonitoringInterval, config.AlertThresholds, config.GoogleChatWebhook,
		config.RetentionPeriod, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		// Config doesn't exist, create it
		return r.create(config)
	}

	return nil
}

// create inserts a new configuration record
func (r *ConfigRepository) create(config *domain.Config) error {
	config.ID = 1 // Ensure single config row
	config.UpdatedAt = time.Now()

	query := `
		INSERT INTO config (
			id, monitoring_interval, alert_thresholds, google_chat_webhook,
			retention_period, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		config.ID, config.MonitoringInterval, config.AlertThresholds,
		config.GoogleChatWebhook, config.RetentionPeriod, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	return nil
}

// createDefaultConfig creates a configuration with default values
func (r *ConfigRepository) createDefaultConfig() *domain.Config {
	config := &domain.Config{
		ID:                 1,
		GoogleChatWebhook:  "",
		UpdatedAt:          time.Now(),
	}

	// Set default monitoring interval: 1 day
	config.SetMonitoringInterval(24 * time.Hour)

	// Set default retention period: 90 days
	config.SetRetentionPeriod(90 * 24 * time.Hour)

	// Set default alert thresholds: 3 months, 2 months, 1 month, 1 week
	defaultThresholds := []time.Duration{
		90 * 24 * time.Hour,  // 3 months
		60 * 24 * time.Hour,  // 2 months
		30 * 24 * time.Hour,  // 1 month
		7 * 24 * time.Hour,   // 1 week
	}
	config.SetAlertThresholds(defaultThresholds)

	return config
}
