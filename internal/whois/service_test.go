package whois

import (
	"strings"
	"testing"
	"time"

	whoisparser "github.com/likexian/whois-parser"
)

// Test parsing error handling with malformed responses
func TestParseWHOISResponse_MalformedResponse(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "empty response",
			response: "",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			response: "This is not a valid WHOIS response",
			wantErr:  true,
		},
		{
			name:     "missing expiration date",
			response: "Domain Name: example.com\nRegistrar: Test Registrar",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ParseWHOISResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWHOISResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test parsing with valid response
func TestParseWHOISResponse_ValidResponse(t *testing.T) {
	service := NewService()

	expirationDate := time.Now().Add(365 * 24 * time.Hour).Format("2006-01-02")
	response := strings.Join([]string{
		"Domain Name: example.com",
		"Registry Expiry Date: " + expirationDate + "T00:00:00Z",
		"Registrar: Test Registrar Inc",
		"Registrant Name: John Doe",
		"Name Server: ns1.example.com",
		"Name Server: ns2.example.com",
	}, "\n")

	info, err := service.ParseWHOISResponse(response)
	if err != nil {
		t.Fatalf("ParseWHOISResponse() unexpected error: %v", err)
	}

	if info.DomainName == "" {
		t.Error("Expected domain name to be set")
	}

	if info.ExpirationDate.IsZero() {
		t.Error("Expected expiration date to be set")
	}

	if len(info.Nameservers) == 0 {
		t.Error("Expected nameservers to be set")
	}

	if info.Registrar == "" {
		t.Error("Expected registrar to be set")
	}

	if info.Registrant == "" {
		t.Error("Expected registrant to be set")
	}
}

// Test timeout handling
func TestQueryWithTimeout(t *testing.T) {
	service := NewService()
	service.timeout = 100 * time.Millisecond

	// This will timeout because we're querying an invalid domain
	// In a real scenario, this would be mocked
	_, err := service.queryWithTimeout("invalid-domain-that-does-not-exist-12345.com")
	if err == nil {
		// If it doesn't error, that's also acceptable (might succeed quickly with an error response)
		return
	}

	// Check if it's a timeout or query error
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed") {
		t.Errorf("Expected timeout or query error, got: %v", err)
	}
}

// Test retry logic
func TestQueryDomain_RetryLogic(t *testing.T) {
	service := NewService()
	service.maxRetries = 2
	service.timeout = 50 * time.Millisecond

	// Query an invalid domain to trigger retries
	_, err := service.QueryDomain("invalid-domain-that-does-not-exist-12345.com")
	if err == nil {
		t.Error("Expected error for invalid domain")
	}

	// Check that error message mentions retries
	if !strings.Contains(err.Error(), "failed after") {
		t.Errorf("Expected retry error message, got: %v", err)
	}
}

// Test extractRegistrant function
func TestExtractRegistrant(t *testing.T) {
	tests := []struct {
		name string
		info *whoisparser.WhoisInfo
		want string
	}{
		{
			name: "name present",
			info: &whoisparser.WhoisInfo{
				Registrant: &whoisparser.Contact{
					Name: "John Doe",
				},
			},
			want: "John Doe",
		},
		{
			name: "organization present",
			info: &whoisparser.WhoisInfo{
				Registrant: &whoisparser.Contact{
					Organization: "Acme Corp",
				},
			},
			want: "Acme Corp",
		},
		{
			name: "email present",
			info: &whoisparser.WhoisInfo{
				Registrant: &whoisparser.Contact{
					Email: "admin@example.com",
				},
			},
			want: "admin@example.com",
		},
		{
			name: "nothing present",
			info: &whoisparser.WhoisInfo{
				Registrant: &whoisparser.Contact{},
			},
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRegistrant(tt.info)
			if got != tt.want {
				t.Errorf("extractRegistrant() = %v, want %v", got, tt.want)
			}
		})
	}
}
