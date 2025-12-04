package whois

import (
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: domain-expiration-monitor, Property 1: Complete WHOIS data extraction
// For any valid domain name, performing a WHOIS query should return a DomainInfo structure
// containing all required fields: expiration date, nameservers, registrant information, and registrar information.
// Validates: Requirements 1.1, 1.2, 1.3, 1.4
func TestProperty_CompleteWHOISDataExtraction(t *testing.T) {
	service := NewService()
	properties := gopter.NewProperties(nil)

	properties.Property("parsed WHOIS response contains all required fields", prop.ForAll(
		func(seed uint32) bool {
			// Use seed to generate deterministic but varied test data
			domainName := "testdomain" + string(rune(seed%1000))
			registrar := "TestRegistrar" + string(rune(seed%100))
			registrant := "TestRegistrant" + string(rune(seed%100))
			
			// Create a mock WHOIS response
			expirationDate := time.Now().Add(365 * 24 * time.Hour).Format("2006-01-02")
			mockResponse := createMockWHOISResponse(domainName, expirationDate, registrar, registrant)

			// Parse the response
			info, err := service.ParseWHOISResponse(mockResponse)
			if err != nil {
				return false
			}

			// Verify all required fields are present
			return info.DomainName != "" &&
				!info.ExpirationDate.IsZero() &&
				len(info.Nameservers) > 0 &&
				info.Registrant != "" &&
				info.Registrar != ""
		},
		gen.UInt32(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// createMockWHOISResponse creates a mock WHOIS response for testing
func createMockWHOISResponse(domain, expirationDate, registrar, registrant string) string {
	return strings.Join([]string{
		"Domain Name: " + domain + ".com",
		"Registry Domain ID: 123456789_DOMAIN_COM-VRSN",
		"Registrar WHOIS Server: whois.example.com",
		"Registrar URL: http://www.example.com",
		"Updated Date: 2024-01-01T00:00:00Z",
		"Creation Date: 2020-01-01T00:00:00Z",
		"Registry Expiry Date: " + expirationDate + "T00:00:00Z",
		"Registrar: " + registrar,
		"Registrar IANA ID: 123",
		"Registrar Abuse Contact Email: abuse@example.com",
		"Registrar Abuse Contact Phone: +1.1234567890",
		"Domain Status: clientTransferProhibited",
		"Registry Registrant ID: REDACTED FOR PRIVACY",
		"Registrant Name: " + registrant,
		"Registrant Organization: " + registrant + " Inc",
		"Registrant Street: REDACTED FOR PRIVACY",
		"Registrant City: REDACTED FOR PRIVACY",
		"Registrant State/Province: CA",
		"Registrant Postal Code: REDACTED FOR PRIVACY",
		"Registrant Country: US",
		"Registrant Phone: REDACTED FOR PRIVACY",
		"Registrant Phone Ext: REDACTED FOR PRIVACY",
		"Registrant Fax: REDACTED FOR PRIVACY",
		"Registrant Fax Ext: REDACTED FOR PRIVACY",
		"Registrant Email: " + strings.ToLower(registrant) + "@example.com",
		"Name Server: ns1.example.com",
		"Name Server: ns2.example.com",
		"DNSSEC: unsigned",
	}, "\n")
}
