# Design Document

## Overview

The Domain Expiration Monitor (DEM) is a Go-based application that continuously monitors domain expiration dates through WHOIS queries and sends configurable alerts to Google Chat. The system consists of four main components: a WHOIS service for querying domain information, a scheduler for managing periodic checks, a web UI for configuration and monitoring, and an alert service for sending notifications.

The application uses a SQLite database for persistent storage, providing a lightweight solution that doesn't require external database infrastructure. The architecture follows clean architecture principles with clear separation between domain logic, data access, and external integrations.

## Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Web UI Layer                         │
│  (HTTP Server, Handlers, Templates, REST API)               │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                      Application Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Scheduler  │  │ WHOIS Service│  │Alert Service │     │
│  │              │  │              │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                      Domain Layer                            │
│  (Domain Models, Business Logic, Interfaces)                │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                   Data Access Layer                          │
│  (Repository Pattern, SQLite Database)                       │
└──────────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

1. **Monitoring Flow**: Scheduler → WHOIS Service → Repository → Alert Service → Google Chat
2. **Configuration Flow**: Web UI → Application Layer → Repository
3. **Query Flow**: Web UI → Repository → Web UI

## Components and Interfaces

### 1. WHOIS Service

**Responsibility**: Query WHOIS servers and parse domain information

**Key Operations**:
- `QueryDomain(domain string) (*DomainInfo, error)`: Performs WHOIS lookup
- `ParseWHOISResponse(response string) (*DomainInfo, error)`: Extracts structured data

**Implementation Notes**:
- Use the `likexian/whois` Go library for WHOIS queries
- Use the `likexian/whois-parser` library for parsing responses
- Implement retry logic with exponential backoff for failed queries
- Handle rate limiting by respecting WHOIS server delays
- Cache parsed results to avoid redundant queries

### 2. Scheduler

**Responsibility**: Manage periodic WHOIS checks for all monitored domains

**Key Operations**:
- `Start()`: Initialize and start the scheduler
- `Stop()`: Gracefully shutdown the scheduler
- `ScheduleDomain(domain *Domain)`: Add a domain to the monitoring schedule
- `UnscheduleDomain(domainID string)`: Remove a domain from monitoring

**Implementation Notes**:
- Use Go's `time.Ticker` for periodic execution
- Maintain a map of domain IDs to their next check time
- Run checks in separate goroutines to avoid blocking
- Use a worker pool pattern to limit concurrent WHOIS queries
- Persist schedule state to handle application restarts

### 3. Alert Service

**Responsibility**: Evaluate alert thresholds and send notifications to Google Chat

**Key Operations**:
- `EvaluateAlerts(domain *Domain) error`: Check if any thresholds are crossed
- `SendAlert(alert *Alert) error`: Send notification to Google Chat
- `FormatAlertMessage(alert *Alert) string`: Create human-readable message

**Implementation Notes**:
- Use Go's `net/http` package for webhook requests
- Implement retry logic with exponential backoff (max 3 retries)
- Track which alerts have been sent to prevent duplicates
- Format messages using Google Chat's card-based message format
- Include domain details, expiration date, and days remaining

### 4. Web UI

**Responsibility**: Provide user interface for configuration and monitoring

**Key Operations**:
- `GET /`: Dashboard showing all monitored domains
- `GET /domains/:id`: Detailed view of a specific domain
- `POST /domains`: Add a new domain to monitor
- `DELETE /domains/:id`: Remove a domain from monitoring
- `GET /config`: View current configuration
- `POST /config`: Update configuration settings

**Implementation Notes**:
- Use Go's `net/http` package with `html/template` for server-side rendering
- Alternatively, use `chi` or `gin` router for cleaner routing
- Implement RESTful API endpoints for AJAX interactions
- Use HTMX for dynamic UI updates without full page reloads
- Serve static assets (CSS, JS) from embedded filesystem using `embed` package
- Implement basic authentication for production deployments

### 5. Repository Layer

**Responsibility**: Persist and retrieve domain data, configuration, and alert history

**Key Interfaces**:

```go
type DomainRepository interface {
    Create(domain *Domain) error
    GetByID(id string) (*Domain, error)
    GetAll() ([]*Domain, error)
    Update(domain *Domain) error
    Delete(id string) error
    DeleteOlderThan(cutoff time.Time) error
}

type ConfigRepository interface {
    Get() (*Config, error)
    Update(config *Config) error
}

type AlertRepository interface {
    Create(alert *Alert) error
    GetByDomainID(domainID string) ([]*Alert, error)
    HasAlertBeenSent(domainID string, threshold time.Duration) (bool, error)
}
```

**Implementation Notes**:
- Use SQLite with `mattn/go-sqlite3` driver
- Use `jmoiron/sqlx` for easier query handling
- Implement connection pooling and prepared statements
- Create indexes on frequently queried columns (domain name, expiration date)
- Use transactions for operations that modify multiple tables

## Data Models

### Domain

```go
type Domain struct {
    ID              string    `db:"id" json:"id"`
    Name            string    `db:"name" json:"name"`
    ExpirationDate  time.Time `db:"expiration_date" json:"expiration_date"`
    Nameservers     []string  `db:"nameservers" json:"nameservers"` // JSON array in DB
    Registrant      string    `db:"registrant" json:"registrant"`
    Registrar       string    `db:"registrar" json:"registrar"`
    LastChecked     time.Time `db:"last_checked" json:"last_checked"`
    NextCheck       time.Time `db:"next_check" json:"next_check"`
    CreatedAt       time.Time `db:"created_at" json:"created_at"`
    UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
```

### Config

```go
type Config struct {
    ID                  int           `db:"id" json:"id"`
    MonitoringInterval  time.Duration `db:"monitoring_interval" json:"monitoring_interval"`
    AlertThresholds     []time.Duration `db:"alert_thresholds" json:"alert_thresholds"` // JSON array
    GoogleChatWebhook   string        `db:"google_chat_webhook" json:"google_chat_webhook"`
    RetentionPeriod     time.Duration `db:"retention_period" json:"retention_period"`
    UpdatedAt           time.Time     `db:"updated_at" json:"updated_at"`
}
```

### Alert

```go
type Alert struct {
    ID              string    `db:"id" json:"id"`
    DomainID        string    `db:"domain_id" json:"domain_id"`
    DomainName      string    `db:"domain_name" json:"domain_name"`
    Threshold       time.Duration `db:"threshold" json:"threshold"`
    ExpirationDate  time.Time `db:"expiration_date" json:"expiration_date"`
    SentAt          time.Time `db:"sent_at" json:"sent_at"`
    Success         bool      `db:"success" json:"success"`
    ErrorMessage    string    `db:"error_message" json:"error_message"`
}
```

### DomainInfo (WHOIS Response)

```go
type DomainInfo struct {
    DomainName      string
    ExpirationDate  time.Time
    Nameservers     []string
    Registrant      string
    Registrar       string
    CreatedDate     time.Time
    UpdatedDate     time.Time
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Complete WHOIS data extraction
*For any* valid domain name, performing a WHOIS query should return a DomainInfo structure containing all required fields: expiration date, nameservers, registrant information, and registrar information.
**Validates: Requirements 1.1, 1.2, 1.3, 1.4**

### Property 2: Monitoring interval validation
*For any* monitoring interval configuration, the system should accept intervals of one hour or greater and reject intervals less than one hour.
**Validates: Requirements 2.3**

### Property 3: Configuration persistence
*For any* valid configuration change (monitoring interval, alert thresholds, webhook URL, retention period), saving the configuration and then retrieving it should return the same values.
**Validates: Requirements 2.2, 6.3, 7.3**

### Property 4: Domain display completeness
*For any* domain in the monitoring list, the rendered UI should contain all required information: domain name, expiration date, nameservers, registrant, registrar, and time remaining until expiration.
**Validates: Requirements 3.3, 3.4, 3.5, 3.6, 3.7**

### Property 5: Domain list completeness
*For any* set of domains in the database, the Web UI should display all of them in the domain list.
**Validates: Requirements 3.2**

### Property 6: Domain addition and retrieval
*For any* valid domain name submitted through the Web UI, adding it to the monitoring list should result in the domain being retrievable from the database with all its WHOIS information populated.
**Validates: Requirements 4.1, 4.2**

### Property 7: Domain removal stops monitoring
*For any* domain in the monitoring list, removing it should result in the domain no longer appearing in the active monitoring list and no further WHOIS queries being scheduled for it.
**Validates: Requirements 4.3**

### Property 8: Alert threshold triggering
*For any* domain with an expiration date and any configured alert threshold, when the time until expiration is less than or equal to the threshold, an alert should be generated.
**Validates: Requirements 5.1**

### Property 9: Alert message completeness
*For any* generated alert, the formatted message should contain the domain name, expiration date, and time remaining until expiration.
**Validates: Requirements 5.2, 5.3, 5.4**

### Property 10: Alert deduplication
*For any* domain and alert threshold combination, only one alert should be sent when that threshold is crossed, and subsequent evaluations should not generate duplicate alerts for the same threshold.
**Validates: Requirements 6.4, 6.5**

### Property 11: Alert threshold validation
*For any* alert threshold configuration, the system should accept positive durations and reject zero or negative durations.
**Validates: Requirements 6.2**

### Property 12: Webhook URL validation
*For any* webhook URL configuration, the system should accept URLs using the HTTPS protocol and reject URLs using HTTP or other protocols.
**Validates: Requirements 7.1**

### Property 13: Retention period validation
*For any* retention period configuration, the system should accept periods of one day or greater and reject periods less than one day.
**Validates: Requirements 8.1**

### Property 14: Retention policy application
*For any* configured retention period, historical domain records with timestamps older than the retention period should be deleted, while records within the retention period should be preserved.
**Validates: Requirements 8.2, 8.3**

### Property 15: Active domain preservation
*For any* domain currently in the active monitoring list, its database record should be preserved regardless of when it was created, even if it exceeds the retention period.
**Validates: Requirements 8.4**

### Property 16: HTTP error responses
*For any* error condition in the Web UI, the HTTP response should have a status code in the 4xx or 5xx range and include an error message in the response body.
**Validates: Requirements 9.4**

### Property 17: Startup domain loading
*For any* set of domains persisted in the database, starting the application should load all of them into the active monitoring list.
**Validates: Requirements 10.1**

### Property 18: Scheduler initialization
*For any* set of domains loaded on startup, each domain should have a scheduled WHOIS query based on its monitoring interval.
**Validates: Requirements 10.2**

### Property 19: Alert evaluation after query
*For any* WHOIS query that completes successfully, the alert service should evaluate all configured thresholds for that domain.
**Validates: Requirements 10.4**

### Property 20: Graceful shutdown persistence
*For any* system state (domains, configuration, alerts), performing a graceful shutdown and then restarting should restore the same state.
**Validates: Requirements 10.5**

## Error Handling

### WHOIS Query Errors

1. **Timeout Handling**: Implement configurable timeout (default 30 seconds) for WHOIS queries
2. **Retry Strategy**: Use exponential backoff with jitter for retries (1s, 2s, 4s, 8s)
3. **Rate Limiting**: Respect WHOIS server rate limits by adding delays between queries
4. **Parse Errors**: Log raw WHOIS response when parsing fails for debugging
5. **Network Errors**: Distinguish between temporary (retry) and permanent (alert admin) failures

### Database Errors

1. **Connection Failures**: Implement connection pool with automatic reconnection
2. **Transaction Rollback**: Wrap multi-step operations in transactions with proper rollback
3. **Constraint Violations**: Return user-friendly error messages for duplicate domains
4. **Migration Errors**: Validate schema version on startup and fail fast if incompatible

### Alert Service Errors

1. **Webhook Failures**: Retry up to 3 times with exponential backoff (1s, 2s, 4s)
2. **Invalid Webhook URL**: Validate URL format before attempting to send
3. **Timeout**: Set 10-second timeout for webhook HTTP requests
4. **Fallback**: Log alerts locally when webhook is unavailable

### Web UI Errors

1. **Invalid Input**: Return 400 Bad Request with validation error details
2. **Not Found**: Return 404 for non-existent domains
3. **Server Errors**: Return 500 with generic message, log detailed error
4. **Concurrent Modifications**: Use optimistic locking to detect conflicts

## Testing Strategy

### Unit Testing

The application will use Go's built-in `testing` package for unit tests. Unit tests will focus on:

- **WHOIS parsing logic**: Test parsing of various WHOIS response formats
- **Alert threshold evaluation**: Test edge cases around threshold boundaries
- **Configuration validation**: Test validation rules for all configuration fields
- **Date calculations**: Test time-until-expiration calculations with various dates
- **Message formatting**: Test alert message generation with different inputs

Unit tests should be co-located with source files using the `_test.go` suffix. Each package should have comprehensive unit test coverage for its core logic.

### Property-Based Testing

The application will use **`gopter`** (https://github.com/leanovate/gopter) for property-based testing. This library provides QuickCheck-style property testing for Go.

**Configuration**:
- Each property-based test MUST run a minimum of 100 iterations
- Use `gopter.NewProperties()` with `MinSuccessfulTests(100)`
- Configure appropriate generators for domain names, dates, and durations

**Test Organization**:
- Property-based tests should be in separate `_property_test.go` files
- Each test MUST include a comment tag in this exact format: `// Feature: domain-expiration-monitor, Property {number}: {property_text}`
- Each correctness property from the design document MUST be implemented by exactly ONE property-based test

**Example Property Test Structure**:

```go
// Feature: domain-expiration-monitor, Property 1: Complete WHOIS data extraction
func TestProperty_CompleteWHOISDataExtraction(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("WHOIS query returns all required fields", 
        prop.ForAll(
            func(domain string) bool {
                info, err := whoisService.QueryDomain(domain)
                if err != nil {
                    return false
                }
                return info.ExpirationDate != nil &&
                       len(info.Nameservers) > 0 &&
                       info.Registrant != "" &&
                       info.Registrar != ""
            },
            genValidDomain(),
        ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

**Generators**:
- Domain names: Generate valid domain formats (alphanumeric + hyphens, valid TLDs)
- Dates: Generate dates within reasonable ranges (past to 10 years future)
- Durations: Generate positive durations from 1 hour to 1 year
- URLs: Generate valid HTTPS URLs for webhook testing

### Integration Testing

Integration tests will verify:
- Database operations with actual SQLite database
- HTTP handlers with test HTTP server
- End-to-end flows: add domain → query WHOIS → evaluate alerts → send notification

### Test Doubles

- **Mock WHOIS Service**: For testing without external WHOIS queries
- **Mock HTTP Client**: For testing webhook calls without external requests
- **In-Memory Database**: For fast unit tests that need persistence

## Deployment Considerations

### Configuration

- Use environment variables for sensitive data (webhook URLs)
- Provide configuration file (YAML/JSON) for non-sensitive settings
- Support configuration via command-line flags for container deployments

### Database

- SQLite database file should be stored in a persistent volume
- Implement automatic database migrations on startup
- Provide backup/restore utilities

### Monitoring

- Expose Prometheus metrics endpoint for monitoring
- Log structured JSON logs for easy parsing
- Include health check endpoint for container orchestration

### Security

- Implement rate limiting on API endpoints
- Add basic authentication for Web UI
- Validate and sanitize all user inputs
- Use HTTPS for production deployments
- Store webhook URLs encrypted at rest

## Performance Considerations

### Scalability

- Worker pool limits concurrent WHOIS queries (default: 10 workers)
- Database connection pool (default: 25 connections)
- Pagination for domain list in UI (50 domains per page)

### Optimization

- Cache WHOIS results to avoid redundant queries
- Use database indexes on frequently queried columns
- Batch database operations where possible
- Use prepared statements for repeated queries

### Resource Management

- Graceful shutdown with timeout (30 seconds)
- Context-based cancellation for long-running operations
- Proper cleanup of goroutines and connections
- Memory-efficient handling of large WHOIS responses
