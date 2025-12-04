# Testing Guide

## Testing Alerts

### Method 1: Using webhook.site (No Google Chat needed)

1. Go to https://webhook.site and copy your unique URL
2. In the app, go to Configuration and paste the webhook URL (change `https://` to `https://` if needed)
3. Add a test domain or run the test script:
   ```bash
   ./test_alert.sh
   ```
4. Wait for the scheduler to run (or restart the app to trigger immediate check)
5. Check webhook.site to see the alert payload

### Method 2: Using Google Chat

1. Create a Google Chat webhook:
   - Open Google Chat
   - Go to a space
   - Click space name → Manage webhooks
   - Create a webhook and copy the URL

2. Configure in the app:
   - Go to http://localhost:8080/config
   - Paste the webhook URL
   - Save configuration

3. Trigger an alert:
   ```bash
   ./test_alert.sh
   ```

### Method 3: Manual Database Update

```bash
# Set a domain to expire in 25 days (triggers 30-day alert)
sqlite3 dem.db "UPDATE domains SET expiration_date = datetime('now', '+25 days') WHERE name = 'your-domain.com'"

# Restart the app to trigger immediate check
```

### Method 4: Run Alert Tests

```bash
# Run the property-based tests for alerts
go test ./internal/alert -v -run TestProperty

# This will test:
# - Alert threshold triggering
# - Alert message completeness
# - Alert deduplication
```

## Viewing Alert History

1. Go to the dashboard: http://localhost:8080
2. Click on any domain
3. Scroll to "Alert History" section
4. You'll see all sent alerts with status and timestamps

## Alert Thresholds

Default thresholds (configurable in database):
- 90 days (3 months)
- 60 days (2 months)
- 30 days (1 month)
- 7 days (1 week)

Each threshold triggers only once per domain.
