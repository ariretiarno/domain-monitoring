package domain

import (
	"time"
)

// DomainInfo represents the parsed WHOIS response for a domain
type DomainInfo struct {
	DomainName     string
	ExpirationDate time.Time
	Nameservers    []string
	Registrant     string
	Registrar      string
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

// IsValid checks if the DomainInfo contains all required fields
func (di *DomainInfo) IsValid() bool {
	return di.DomainName != "" &&
		!di.ExpirationDate.IsZero() &&
		len(di.Nameservers) > 0 &&
		di.Registrant != "" &&
		di.Registrar != ""
}

// DaysUntilExpiration calculates the number of days until the domain expires
func (di *DomainInfo) DaysUntilExpiration() int {
	duration := time.Until(di.ExpirationDate)
	return int(duration.Hours() / 24)
}
