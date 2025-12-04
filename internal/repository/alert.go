package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/google/uuid"
)

// AlertRepository handles alert data persistence
type AlertRepository struct {
	db *DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create adds a new alert to the database
func (r *AlertRepository) Create(alert *domain.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	query := `
		INSERT INTO alerts (
			id, domain_id, domain_name, threshold, expiration_date,
			sent_at, success, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		alert.ID, alert.DomainID, alert.DomainName, alert.Threshold,
		alert.ExpirationDate, alert.SentAt, alert.Success, alert.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetByDomainID retrieves all alerts for a specific domain
func (r *AlertRepository) GetByDomainID(domainID string) ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	query := `
		SELECT id, domain_id, domain_name, threshold, expiration_date,
		       sent_at, success, error_message
		FROM alerts
		WHERE domain_id = ?
		ORDER BY sent_at DESC
	`

	err := r.db.Select(&alerts, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts for domain: %w", err)
	}

	return alerts, nil
}

// HasAlertBeenSent checks if an alert has already been sent for a domain and threshold
// This checks for ANY alert attempt (successful or not) to prevent duplicate alerts
func (r *AlertRepository) HasAlertBeenSent(domainID string, threshold time.Duration) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM alerts
		WHERE domain_id = ? AND threshold = ?
	`

	err := r.db.Get(&count, query, domainID, int64(threshold))
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if alert was sent: %w", err)
	}

	return count > 0, nil
}

// GetRecentAlerts retrieves alerts sent within a specific time period
func (r *AlertRepository) GetRecentAlerts(since time.Time) ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	query := `
		SELECT id, domain_id, domain_name, threshold, expiration_date,
		       sent_at, success, error_message
		FROM alerts
		WHERE sent_at >= ?
		ORDER BY sent_at DESC
	`

	err := r.db.Select(&alerts, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}

	return alerts, nil
}

// DeleteOlderThan deletes alerts that were sent before the cutoff time
func (r *AlertRepository) DeleteOlderThan(cutoff time.Time) error {
	query := `DELETE FROM alerts WHERE sent_at < ?`

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old alerts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Log how many alerts were deleted (in production, use proper logging)
	_ = rows

	return nil
}

// GetFailedAlerts retrieves alerts that failed to send
func (r *AlertRepository) GetFailedAlerts() ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	query := `
		SELECT id, domain_id, domain_name, threshold, expiration_date,
		       sent_at, success, error_message
		FROM alerts
		WHERE success = 0
		ORDER BY sent_at DESC
	`

	err := r.db.Select(&alerts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed alerts: %w", err)
	}

	return alerts, nil
}
