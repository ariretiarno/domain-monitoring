package domain

import (
	"time"
)

// Alert represents a notification sent for a domain approaching expiration
type Alert struct {
	ID             string    `db:"id" json:"id"`
	DomainID       string    `db:"domain_id" json:"domain_id"`
	DomainName     string    `db:"domain_name" json:"domain_name"`
	Threshold      int64     `db:"threshold" json:"threshold"` // stored as nanoseconds
	ExpirationDate time.Time `db:"expiration_date" json:"expiration_date"`
	SentAt         time.Time `db:"sent_at" json:"sent_at"`
	Success        bool      `db:"success" json:"success"`
	ErrorMessage   string    `db:"error_message" json:"error_message"`
}

// GetThreshold returns the threshold as a time.Duration
func (a *Alert) GetThreshold() time.Duration {
	return time.Duration(a.Threshold)
}

// SetThreshold sets the threshold from a time.Duration
func (a *Alert) SetThreshold(d time.Duration) {
	a.Threshold = int64(d)
}

// DaysUntilExpiration calculates the number of days until expiration at the time the alert was sent
func (a *Alert) DaysUntilExpiration() int {
	duration := a.ExpirationDate.Sub(a.SentAt)
	return int(duration.Hours() / 24)
}
