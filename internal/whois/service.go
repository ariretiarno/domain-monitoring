package whois

import (
	"fmt"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

// Service handles WHOIS queries and parsing
type Service struct {
	timeout time.Duration
	maxRetries int
}

// NewService creates a new WHOIS service
func NewService() *Service {
	return &Service{
		timeout:    30 * time.Second,
		maxRetries: 3,
	}
}

// QueryDomain performs a WHOIS lookup for a domain with retry logic
func (s *Service) QueryDomain(domainName string) (*domain.DomainInfo, error) {
	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt < s.maxRetries; attempt++ {
		info, err := s.queryWithTimeout(domainName)
		if err == nil {
			return info, nil
		}

		lastErr = err
		if attempt < s.maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", s.maxRetries, lastErr)
}

// queryWithTimeout performs a single WHOIS query with timeout
func (s *Service) queryWithTimeout(domainName string) (*domain.DomainInfo, error) {
	// Create a channel for the result
	type result struct {
		info *domain.DomainInfo
		err  error
	}
	resultChan := make(chan result, 1)

	go func() {
		info, err := s.query(domainName)
		resultChan <- result{info, err}
	}()

	select {
	case res := <-resultChan:
		return res.info, res.err
	case <-time.After(s.timeout):
		return nil, fmt.Errorf("WHOIS query timed out after %v", s.timeout)
	}
}

// query performs the actual WHOIS lookup and parsing
func (s *Service) query(domainName string) (*domain.DomainInfo, error) {
	// Perform WHOIS query
	rawResponse, err := whois.Whois(domainName)
	if err != nil {
		return nil, fmt.Errorf("WHOIS query failed: %w", err)
	}

	// Parse WHOIS response
	info, err := s.ParseWHOISResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("WHOIS parsing failed for %s: %w\nRaw response: %s", domainName, err, rawResponse)
	}

	return info, nil
}

// ParseWHOISResponse parses a raw WHOIS response into structured data
func (s *Service) ParseWHOISResponse(rawResponse string) (*domain.DomainInfo, error) {
	parsed, err := whoisparser.Parse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WHOIS response: %w", err)
	}

	// Check if Domain is nil
	if parsed.Domain == nil {
		return nil, fmt.Errorf("WHOIS response does not contain domain information")
	}
	
	// Extract expiration date - try multiple formats
	dateFormats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05 MST",
	}
	
	var expirationDate time.Time
	var parseErr error
	
	for _, format := range dateFormats {
		expirationDate, parseErr = time.Parse(format, parsed.Domain.ExpirationDate)
		if parseErr == nil {
			break
		}
	}
	
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse expiration date '%s': %w", parsed.Domain.ExpirationDate, parseErr)
	}

	// Extract created date
	var createdDate time.Time
	if parsed.Domain.CreatedDate != "" {
		for _, format := range dateFormats {
			createdDate, parseErr = time.Parse(format, parsed.Domain.CreatedDate)
			if parseErr == nil {
				break
			}
		}
	}

	// Extract updated date
	var updatedDate time.Time
	if parsed.Domain.UpdatedDate != "" {
		for _, format := range dateFormats {
			updatedDate, parseErr = time.Parse(format, parsed.Domain.UpdatedDate)
			if parseErr == nil {
				break
			}
		}
	}

	// Build domain info with nil checks
	domainName := ""
	nameservers := []string{}
	if parsed.Domain != nil {
		domainName = parsed.Domain.Domain
		nameservers = parsed.Domain.NameServers
	}
	
	registrarName := ""
	if parsed.Registrar != nil {
		registrarName = parsed.Registrar.Name
	}
	
	info := &domain.DomainInfo{
		DomainName:     domainName,
		ExpirationDate: expirationDate,
		Nameservers:    nameservers,
		Registrant:     extractRegistrant(&parsed),
		Registrar:      registrarName,
		CreatedDate:    createdDate,
		UpdatedDate:    updatedDate,
	}

	return info, nil
}

// extractRegistrant extracts registrant information from parsed WHOIS data
func extractRegistrant(parsed *whoisparser.WhoisInfo) string {
	if parsed.Registrant == nil {
		return "Unknown"
	}
	if parsed.Registrant.Name != "" {
		return parsed.Registrant.Name
	}
	if parsed.Registrant.Organization != "" {
		return parsed.Registrant.Organization
	}
	if parsed.Registrant.Email != "" {
		return parsed.Registrant.Email
	}
	return "Unknown"
}
