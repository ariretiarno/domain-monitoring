# Requirements Document

## Introduction

This document specifies the requirements for a Domain Expiration Monitor application built in Go. The system monitors domain expiration dates via WHOIS queries, displays domain information through a web UI, and sends configurable alerts to Google Chat when domains approach expiration. The application provides comprehensive configuration options for monitoring intervals, alert thresholds, and data retention.

## Glossary

- **Domain Expiration Monitor (DEM)**: The complete application system that monitors domain expiration dates
- **WHOIS Service**: The component responsible for querying WHOIS data for domains
- **Alert Service**: The component that sends notifications to Google Chat
- **Web UI**: The web-based user interface for configuration and monitoring
- **Monitoring Interval**: The time period between consecutive WHOIS checks for a domain
- **Alert Threshold**: A time period before expiration that triggers an alert (e.g., 3 months, 1 week)
- **Retention Period**: The duration for which domain monitoring history is stored
- **Domain Record**: A stored entry containing domain information and monitoring metadata
- **Google Chat Webhook**: An HTTPS endpoint for sending messages to Google Chat spaces

## Requirements

### Requirement 1

**User Story:** As a system administrator, I want to monitor domain expiration dates automatically, so that I can renew domains before they expire and avoid service disruptions.

#### Acceptance Criteria

1. WHEN the DEM performs a WHOIS query for a domain, THEN the DEM SHALL retrieve the domain expiration date
2. WHEN the DEM performs a WHOIS query for a domain, THEN the DEM SHALL extract nameserver information from the WHOIS response
3. WHEN the DEM performs a WHOIS query for a domain, THEN the DEM SHALL extract registrant information from the WHOIS response
4. WHEN the DEM performs a WHOIS query for a domain, THEN the DEM SHALL extract registrar information from the WHOIS response
5. WHEN a WHOIS query fails, THEN the DEM SHALL log the error and retry according to the configured monitoring interval

### Requirement 2

**User Story:** As a system administrator, I want to configure how often domains are checked, so that I can balance monitoring frequency with system resources and WHOIS rate limits.

#### Acceptance Criteria

1. THE DEM SHALL provide a default monitoring interval of one day for WHOIS queries
2. WHEN a user configures a monitoring interval via the Web UI, THEN the DEM SHALL apply the new interval to subsequent WHOIS queries
3. WHEN a monitoring interval is configured, THEN the DEM SHALL validate that the interval is at least one hour
4. WHEN the monitoring interval elapses for a domain, THEN the DEM SHALL execute a WHOIS query for that domain

### Requirement 3

**User Story:** As a system administrator, I want to view domain information through a web interface, so that I can easily monitor the status of all tracked domains without using command-line tools.

#### Acceptance Criteria

1. THE DEM SHALL provide a Web UI accessible via HTTP
2. WHEN a user accesses the Web UI, THEN the DEM SHALL display a list of all monitored domains
3. WHEN a user views a domain in the Web UI, THEN the DEM SHALL display the domain expiration date
4. WHEN a user views a domain in the Web UI, THEN the DEM SHALL display the nameserver information
5. WHEN a user views a domain in the Web UI, THEN the DEM SHALL display the registrant information
6. WHEN a user views a domain in the Web UI, THEN the DEM SHALL display the registrar information
7. WHEN a user views a domain in the Web UI, THEN the DEM SHALL display the time remaining until expiration

### Requirement 4

**User Story:** As a system administrator, I want to add and remove domains for monitoring through the web interface, so that I can manage my domain portfolio efficiently.

#### Acceptance Criteria

1. WHEN a user submits a new domain via the Web UI, THEN the DEM SHALL add the domain to the monitoring list
2. WHEN a user submits a new domain via the Web UI, THEN the DEM SHALL perform an immediate WHOIS query for validation
3. WHEN a user removes a domain via the Web UI, THEN the DEM SHALL stop monitoring that domain
4. WHEN a user removes a domain via the Web UI, THEN the DEM SHALL retain historical data according to the retention policy

### Requirement 5

**User Story:** As a system administrator, I want to receive alerts in Google Chat when domains approach expiration, so that my team is notified through our existing communication platform.

#### Acceptance Criteria

1. WHEN a domain expiration date is within an alert threshold, THEN the Alert Service SHALL send a notification to the configured Google Chat webhook
2. WHEN sending an alert, THEN the Alert Service SHALL include the domain name in the message
3. WHEN sending an alert, THEN the Alert Service SHALL include the expiration date in the message
4. WHEN sending an alert, THEN the Alert Service SHALL include the time remaining until expiration in the message
5. WHEN a Google Chat webhook request fails, THEN the Alert Service SHALL log the error and retry up to three times

### Requirement 6

**User Story:** As a system administrator, I want to configure multiple alert thresholds for domain expiration, so that I receive timely notifications at different stages before expiration.

#### Acceptance Criteria

1. THE DEM SHALL provide default alert thresholds of three months, two months, one month, and one week before expiration
2. WHEN a user configures alert thresholds via the Web UI, THEN the DEM SHALL validate that each threshold is a positive duration
3. WHEN a user configures alert thresholds via the Web UI, THEN the DEM SHALL apply the new thresholds to all monitored domains
4. WHEN a domain crosses an alert threshold, THEN the DEM SHALL send exactly one alert for that threshold
5. WHEN a domain has already triggered an alert for a specific threshold, THEN the DEM SHALL NOT send duplicate alerts for the same threshold

### Requirement 7

**User Story:** As a system administrator, I want to configure the Google Chat webhook URL through the web interface, so that I can direct alerts to the appropriate chat space without modifying configuration files.

#### Acceptance Criteria

1. WHEN a user configures a Google Chat webhook URL via the Web UI, THEN the DEM SHALL validate that the URL uses HTTPS protocol
2. WHEN a user configures a Google Chat webhook URL via the Web UI, THEN the DEM SHALL store the webhook URL securely
3. WHEN a user configures a Google Chat webhook URL via the Web UI, THEN the DEM SHALL use the new webhook URL for subsequent alerts
4. WHEN no webhook URL is configured, THEN the DEM SHALL log alerts locally but not send external notifications

### Requirement 8

**User Story:** As a system administrator, I want to configure how long domain monitoring data is retained, so that I can manage storage requirements and comply with data retention policies.

#### Acceptance Criteria

1. WHEN a user configures a retention period via the Web UI, THEN the DEM SHALL validate that the retention period is at least one day
2. WHEN a user configures a retention period via the Web UI, THEN the DEM SHALL apply the retention policy to historical domain records
3. WHEN domain monitoring data exceeds the retention period, THEN the DEM SHALL delete the expired data
4. WHEN a domain is actively monitored, THEN the DEM SHALL retain the current domain information regardless of the retention period
5. THE DEM SHALL provide a default retention period of ninety days

### Requirement 9

**User Story:** As a system administrator, I want the application to handle errors gracefully, so that temporary failures do not cause the monitoring system to crash or lose data.

#### Acceptance Criteria

1. WHEN a WHOIS query times out, THEN the WHOIS Service SHALL log the timeout and schedule a retry
2. WHEN the DEM cannot parse WHOIS response data, THEN the WHOIS Service SHALL log the parsing error with the raw response
3. WHEN the database connection fails, THEN the DEM SHALL attempt to reconnect up to five times with exponential backoff
4. WHEN the Web UI encounters an error, THEN the DEM SHALL return an appropriate HTTP error status and error message
5. WHEN the Alert Service cannot reach the Google Chat webhook, THEN the Alert Service SHALL log the failure and continue monitoring other domains

### Requirement 10

**User Story:** As a system administrator, I want domain monitoring to run continuously in the background, so that I don't need to manually trigger checks and can rely on automated monitoring.

#### Acceptance Criteria

1. WHEN the DEM starts, THEN the DEM SHALL load all monitored domains from persistent storage
2. WHEN the DEM starts, THEN the DEM SHALL schedule WHOIS queries for all monitored domains according to their monitoring intervals
3. WHILE the DEM is running, THE DEM SHALL execute scheduled WHOIS queries at the configured intervals
4. WHILE the DEM is running, THE DEM SHALL evaluate alert thresholds after each WHOIS query
5. WHEN the DEM shuts down gracefully, THEN the DEM SHALL persist all domain data and configuration to storage
