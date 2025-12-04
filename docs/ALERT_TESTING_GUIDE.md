# Alert Testing Guide

## How Alerts Work

Alerts are sent when:
1. A domain's expiration date is within an alert threshold (e.g., 30 days)
2. The scheduler checks the domain (runs periodically based on monitoring interval)
3. An alert for that threshold hasn't been sent yet
4. A webhook URL is configured

## Step-by-Step Testing

### Method 1: Quick Test with webhook.site (Recommended)

This method doesn't require Google Chat and shows you the exact webhook payload.

#### Step 1: Get a Test Webhook URL

1. Go to https://webhook.site
2. Copy your unique URL (e.g., `https://webhook.site/12345678-abcd-1234-5678-123456789abc`)

#### Step 2: Configure the Application

1. Start the application:
   ```bash
   ./bin/dem
   ```

2. Open http://localhost:8080/config

3. Paste the webhook.site URL into "Google Chat Webhook URL"

4. Configure alert thresholds (e.g., `90,60,30,7`)

5. Click "Save Configuration"

#### Step 3: Add a Test Domain

Run the test script to add a domain expiring in 25 days:

```bash
./test_alert.sh
```

This will:
- Add a test domain: `test-alert-domain.com`
- Set expiration to 25 days from now
- This will trigger the 30-day alert threshold

#### Step 4: Trigger the Alert

The scheduler checks domains based on the monitoring interval. To trigger immediately:

**Option A: Restart the application**
```bash
# Stop the app (Ctrl+C)
# Start again
./bin/dem
```

**Option B: Wait for the next scheduled check**
- Default monitoring interval is 24 hours
- The domain will be checked automatically

**Option C: Force immediate check via database**
```bash
# Set next_check to now
sqlite3 dem.db "UPDATE domains SET next_check = datetime('now') WHERE name = 'test-alert-domain.com'"

# Restart app to trigger check
```

#### Step 5: Verify Alert Sent

1. Go back to webhook.site
2. You should see a POST request with the alert message
3. The payload will look like:
   ```json
   {
     "text": "🔔 Domain Expiration Alert\n\nDomain: test-alert-domain.com\nExpiration Date: 2025-12-29\nDays Remaining: 25\nAlert Threshold: 30 days\n\nPlease renew this domain to avoid service disruption."
   }
   ```

4. Check the domain detail page: http://localhost:8080/domains/[domain-id]
5. Scroll to "Alert History" - you should see the sent alert

### Method 2: Test with Google Chat

#### Step 1: Create Google Chat Webhook

1. Open Google Chat (https://chat.google.com)
2. Go to a space or create a new one
3. Click the space name → "Manage webhooks"
4. Click "Add webhook"
5. Name it "Domain Monitor" and click "Save"
6. Copy the webhook URL

#### Step 2: Configure and Test

Follow the same steps as Method 1, but use your Google Chat webhook URL instead of webhook.site.

The alert will appear in your Google Chat space!

### Method 3: Test with Real Domain

1. Add a real domain that expires soon:
   ```bash
   # In the web UI, add a domain
   ```

2. Check the domain's expiration date on the dashboard

3. If it's within your alert thresholds, an alert will be sent on the next check

### Method 4: Manual Database Test

For advanced testing, manually insert a domain:

```bash
sqlite3 dem.db << EOF
INSERT INTO domains (
  id, name, expiration_date, nameservers, registrant, registrar,
  last_checked, next_check, created_at, updated_at
) VALUES (
  'test-123',
  'manual-test.com',
  datetime('now', '+20 days'),
  '["ns1.example.com"]',
  'Test User',
  'Test Registrar',
  datetime('now'),
  datetime('now'),
  datetime('now'),
  datetime('now')
);
EOF

# Restart app to trigger check
```

## Verifying Alerts Work

### Check 1: Alert History in UI

1. Go to dashboard: http://localhost:8080
2. Click on the test domain
3. Scroll to "Alert History" section
4. You should see:
   - Sent At: timestamp
   - Threshold: 30 days (or whatever threshold was crossed)
   - Status: ✓ Sent or ✗ Failed
   - Error: (empty if successful)

### Check 2: Database

```bash
# Check alerts table
sqlite3 dem.db "SELECT domain_name, threshold, sent_at, success FROM alerts ORDER BY sent_at DESC LIMIT 5;"
```

### Check 3: Application Logs

```bash
# Run with logging
./bin/dem 2>&1 | grep -i alert
```

## Testing Different Thresholds

### Test Multiple Thresholds

1. Configure thresholds: `90,60,30,14,7,1`

2. Add domains with different expiration dates:
   ```bash
   # 85 days - triggers 90-day alert
   # 55 days - triggers 60-day alert
   # 25 days - triggers 30-day alert
   # 10 days - triggers 14-day alert
   # 5 days - triggers 7-day alert
   # 12 hours - triggers 1-day alert
   ```

3. Each threshold will send exactly one alert

### Test Alert Deduplication

1. Add a domain expiring in 25 days
2. Wait for alert to be sent (30-day threshold)
3. Restart the app multiple times
4. Verify only ONE alert is sent (check Alert History)

## Troubleshooting

### No Alert Sent

**Check 1: Webhook URL configured?**
```bash
sqlite3 dem.db "SELECT google_chat_webhook FROM config;"
```

**Check 2: Domain within threshold?**
```bash
sqlite3 dem.db "SELECT name, expiration_date, julianday(expiration_date) - julianday('now') as days_remaining FROM domains;"
```

**Check 3: Alert already sent?**
```bash
sqlite3 dem.db "SELECT * FROM alerts WHERE domain_name = 'your-domain.com';"
```

**Check 4: Scheduler running?**
- Check application logs for "checking domain" messages
- Verify next_check time is in the past

### Alert Failed

Check the Alert History in the UI for error messages:
- "no webhook URL configured" - Set webhook in config
- "webhook returned status XXX" - Check webhook URL is valid
- "failed to send webhook" - Network issue or invalid URL

### Webhook Not Receiving

1. Test webhook manually:
   ```bash
   curl -X POST https://webhook.site/YOUR-ID \
     -H "Content-Type: application/json" \
     -d '{"text": "Test message"}'
   ```

2. Check webhook.site page is open and refreshed

3. For Google Chat, verify webhook URL is correct and space exists

## Expected Alert Format

```
🔔 Domain Expiration Alert

Domain: example.com
Expiration Date: 2025-12-29
Days Remaining: 25
Alert Threshold: 30 days

Please renew this domain to avoid service disruption.
```

## Advanced: Testing Alert Retry Logic

To test the retry logic (3 attempts with exponential backoff):

1. Configure an invalid webhook URL:
   ```
   https://invalid-webhook-that-does-not-exist.com/webhook
   ```

2. Add a test domain expiring soon

3. Check logs - you should see 3 retry attempts

4. Check Alert History - status will be "✗ Failed" with error message

## Monitoring Alerts in Production

### Check Alert Success Rate

```bash
sqlite3 dem.db "SELECT 
  COUNT(*) as total_alerts,
  SUM(success) as successful,
  COUNT(*) - SUM(success) as failed
FROM alerts;"
```

### View Recent Alerts

```bash
sqlite3 dem.db "SELECT 
  domain_name,
  datetime(sent_at) as sent,
  threshold/86400000000000 as threshold_days,
  CASE WHEN success = 1 THEN 'Success' ELSE 'Failed' END as status
FROM alerts 
ORDER BY sent_at DESC 
LIMIT 10;"
```

### Failed Alerts

```bash
sqlite3 dem.db "SELECT domain_name, error_message, datetime(sent_at) FROM alerts WHERE success = 0;"
```
