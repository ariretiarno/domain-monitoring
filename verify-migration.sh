#!/bin/bash
# Script to verify database migrations

set -e

echo "🔍 Verifying Database Migrations..."
echo ""

# Check if using Docker or local
if docker-compose ps mysql 2>/dev/null | grep -q "Up"; then
    echo "📦 Using Docker MySQL..."
    DB_CMD="docker-compose exec -T mysql mysql -u demuser -pdempassword dem"
else
    echo "💾 Using local database..."
    if [ -f "dem.db" ]; then
        DB_CMD="sqlite3 dem.db"
    else
        echo "❌ No database found. Please start the application first."
        exit 1
    fi
fi

echo ""
echo "Checking tables..."
echo ""

if docker-compose ps mysql 2>/dev/null | grep -q "Up"; then
    # MySQL
    echo "📋 Tables in database:"
    echo "SHOW TABLES;" | $DB_CMD
    
    echo ""
    echo "📊 Domains table structure:"
    echo "DESCRIBE domains;" | $DB_CMD
    
    echo ""
    echo "📊 Config table structure:"
    echo "DESCRIBE config;" | $DB_CMD
    
    echo ""
    echo "📊 Alerts table structure:"
    echo "DESCRIBE alerts;" | $DB_CMD
    
    echo ""
    echo "📈 Record counts:"
    echo "SELECT 'Domains' as table_name, COUNT(*) as count FROM domains
          UNION ALL
          SELECT 'Config', COUNT(*) FROM config
          UNION ALL
          SELECT 'Alerts', COUNT(*) FROM alerts;" | $DB_CMD
else
    # SQLite
    echo "📋 Tables in database:"
    echo ".tables" | $DB_CMD
    
    echo ""
    echo "📊 Domains table structure:"
    echo ".schema domains" | $DB_CMD
    
    echo ""
    echo "📊 Config table structure:"
    echo ".schema config" | $DB_CMD
    
    echo ""
    echo "📊 Alerts table structure:"
    echo ".schema alerts" | $DB_CMD
    
    echo ""
    echo "📈 Record counts:"
    echo "SELECT 'Domains' as table_name, COUNT(*) as count FROM domains
          UNION ALL
          SELECT 'Config', COUNT(*) FROM config
          UNION ALL
          SELECT 'Alerts', COUNT(*) FROM alerts;" | $DB_CMD
fi

echo ""
echo "✅ Migration verification complete!"
