package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Domain represents a monitored domain with its WHOIS information
type Domain struct {
	ID             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	ExpirationDate time.Time `db:"expiration_date" json:"expiration_date"`
	Nameservers    Strings   `db:"nameservers" json:"nameservers"`
	Registrant     string    `db:"registrant" json:"registrant"`
	Registrar      string    `db:"registrar" json:"registrar"`
	LastChecked    time.Time `db:"last_checked" json:"last_checked"`
	NextCheck      time.Time `db:"next_check" json:"next_check"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// DaysUntilExpiration calculates the number of days until the domain expires
func (d *Domain) DaysUntilExpiration() int {
	duration := time.Until(d.ExpirationDate)
	return int(duration.Hours() / 24)
}

// IsExpired checks if the domain has already expired
func (d *Domain) IsExpired() bool {
	return time.Now().After(d.ExpirationDate)
}

// Strings is a custom type for storing string slices as JSON in the database
type Strings []string

// Value implements the driver.Valuer interface for database storage
func (s Strings) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *Strings) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		*s = []string{}
		return nil
	}
	
	return json.Unmarshal(bytes, s)
}
