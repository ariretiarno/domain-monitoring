#!/bin/bash
# Script to test alerts by setting a domain to expire soon

DB_PATH=${1:-dem.db}

echo "Testing alert functionality..."
echo "This will set a test domain to expire in 25 days to trigger alerts"

# Add a test domain with expiration in 25 days
EXPIRY_DATE=$(date -u -v+25d +"%Y-%m-%d %H:%M:%S" 2>/dev/null || date -u -d "+25 days" +"%Y-%m-%d %H:%M:%S")
DOMAIN_ID=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "test-$(date +%s)")

sqlite3 "$DB_PATH" <<EOF
INSERT INTO domains (id, name, expiration_date, nameservers, registrant, registrar, last_checked, next_check, created_at, updated_at)
VALUES (
  '$DOMAIN_ID',
  'test-alert-domain.com',
  '$EXPIRY_DATE',
  '["ns1.example.com","ns2.example.com"]',
  'Test User',
  'Test Registrar',
  datetime('now'),
  datetime('now', '-1 hour'),
  datetime('now'),
  datetime('now')
);
EOF

echo "✓ Test domain added with ID: $DOMAIN_ID"
echo "✓ Expiration date: $EXPIRY_DATE (25 days from now)"
echo "✓ Next check set to 1 hour ago (will trigger immediately on app restart)"
echo ""
echo "To test alerts:"
echo "1. Make sure webhook is configured at http://localhost:8080/config"
echo "2. Restart the application: ./bin/dem"
echo "3. The scheduler will immediately check the domain and send an alert (30-day threshold)"
echo "4. Check the domain detail page to see alert history"
echo "5. Or check webhook.site if you're using that for testing"
echo ""
echo "To remove test domain:"
echo "  sqlite3 $DB_PATH \"DELETE FROM domains WHERE id='$DOMAIN_ID'\""
