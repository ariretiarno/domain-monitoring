package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Config represents the application configuration
type Config struct {
	ID                 int       `db:"id" json:"id"`
	MonitoringInterval int64     `db:"monitoring_interval" json:"monitoring_interval"` // stored as nanoseconds
	AlertThresholds    Durations `db:"alert_thresholds" json:"alert_thresholds"`
	GoogleChatWebhook  string    `db:"google_chat_webhook" json:"google_chat_webhook"`
	RetentionPeriod    int64     `db:"retention_period" json:"retention_period"` // stored as nanoseconds
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// GetMonitoringInterval returns the monitoring interval as a time.Duration
func (c *Config) GetMonitoringInterval() time.Duration {
	return time.Duration(c.MonitoringInterval)
}

// SetMonitoringInterval sets the monitoring interval from a time.Duration
func (c *Config) SetMonitoringInterval(d time.Duration) {
	c.MonitoringInterval = int64(d)
}

// GetRetentionPeriod returns the retention period as a time.Duration
func (c *Config) GetRetentionPeriod() time.Duration {
	return time.Duration(c.RetentionPeriod)
}

// SetRetentionPeriod sets the retention period from a time.Duration
func (c *Config) SetRetentionPeriod(d time.Duration) {
	c.RetentionPeriod = int64(d)
}

// GetAlertThresholds returns the alert thresholds as []time.Duration
func (c *Config) GetAlertThresholds() []time.Duration {
	return []time.Duration(c.AlertThresholds)
}

// SetAlertThresholds sets the alert thresholds from []time.Duration
func (c *Config) SetAlertThresholds(thresholds []time.Duration) {
	c.AlertThresholds = Durations(thresholds)
}

// Durations is a custom type for storing duration slices as JSON in the database
type Durations []time.Duration

// Value implements the driver.Valuer interface for database storage
func (d Durations) Value() (driver.Value, error) {
	if d == nil {
		return json.Marshal([]int64{})
	}
	
	// Convert durations to nanoseconds for storage
	nanos := make([]int64, len(d))
	for i, dur := range d {
		nanos[i] = int64(dur)
	}
	return json.Marshal(nanos)
}

// Scan implements the sql.Scanner interface for database retrieval
func (d *Durations) Scan(value interface{}) error {
	if value == nil {
		*d = []time.Duration{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		*d = []time.Duration{}
		return nil
	}
	
	var nanos []int64
	if err := json.Unmarshal(bytes, &nanos); err != nil {
		return err
	}
	
	// Convert nanoseconds back to durations
	durations := make([]time.Duration, len(nanos))
	for i, nano := range nanos {
		durations[i] = time.Duration(nano)
	}
	*d = durations
	return nil
}
