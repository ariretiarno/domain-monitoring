# Changelog

## [Latest] - 2025-12-04

### Added - Docker Support & Improvements

- ✅ **Docker Deployment**: Complete Docker and Docker Compose setup
  - Multi-stage Dockerfile for optimized builds
  - Docker Compose with MySQL 8.0
  - Automatic database migrations on startup
  - Health checks for all services
  - Persistent volume for MySQL data
  - Isolated network for security
  - See `DOCKER_DEPLOYMENT.md` for full guide

- ✅ **Automatic Migrations**: Database migrations run automatically
  - Creates tables if they don't exist
  - Idempotent operations (safe to run multiple times)
  - Works with both SQLite and MySQL
  - No manual migration steps required
  - Verification script included (`verify-migration.sh`)

- ✅ **Environment Variable Support**: Enhanced configuration
  - MySQL connection from individual env vars (DB_HOST, DB_PORT, etc.)
  - Automatic connection string building
  - Docker-friendly configuration
  - Updated `.env.example` with all options

- ✅ **Helper Scripts**:
  - `docker-start.sh` - Easy Docker startup with validation
  - `verify-migration.sh` - Database migration verification
  - Updated `test_alert.sh` - Works immediately with test domains
  - `Makefile` - Convenient commands for all tasks

- ✅ **Comprehensive Documentation**:
  - `DOCKER_DEPLOYMENT.md` - Complete Docker guide
  - `DOCKER_QUICKREF.md` - Quick command reference
  - `DEPLOYMENT_SUMMARY.md` - Overview and checklist
  - `DOCKER_SETUP_COMPLETE.md` - Setup summary
  - Updated `README.md` with Docker instructions

### Fixed

- ✅ **Alert Evaluation**: Alerts now evaluated even when WHOIS fails
  - Important for test domains that don't exist
  - Uses existing expiration date from database
  - Ensures alerts work reliably
  - Fixes issue where test domains never triggered alerts

- ✅ **Test Alert Script**: Updated to trigger immediately
  - Sets `next_check` to past time
  - Works on app restart without waiting
  - Better testing experience

### Previous Updates - 2025-12-04

### Added
- ✅ **MySQL Support**: Full MySQL database support alongside SQLite
  - Automatic schema creation for both databases
  - Connection string configuration via environment variables
  - See `MYSQL_SETUP.md` for setup instructions

- ✅ **Environment Configuration**: `.env` file support
  - Database driver selection (sqlite3/mysql)
  - Database connection string configuration
  - HTTP server address configuration
  - Example file provided (`.env.example`)

- ✅ **Configurable Alert Thresholds**: Alert thresholds now configurable via UI
  - Edit thresholds in Configuration page
  - Comma-separated values (e.g., 90,60,30,7)
  - Validation for positive values
  - Immediate effect on all monitored domains

- ✅ **Multi-TLD Support**: Enhanced WHOIS parsing
  - Support for multiple date formats
  - Handles .ar, .com, .org, and other TLDs
  - Graceful handling of missing WHOIS fields
  - Nil pointer protection

- ✅ **Testing Tools**:
  - `test_alert.sh` - Script to test alert functionality
  - `TESTING.md` - Comprehensive testing guide
  - webhook.site integration instructions

### Fixed
- Fixed nil pointer dereference in WHOIS parsing
- Fixed date parsing for non-standard formats
- Fixed template rendering issues
- Added proper error handling for missing WHOIS data

### Documentation
- Added `MYSQL_SETUP.md` - MySQL setup guide
- Added `TESTING.md` - Testing guide
- Added `MYSQL_MIGRATION.md` - Migration guide
- Updated `README.md` with new features
- Added `.env.example` for configuration

## [Initial Release]

### Features
- Domain expiration monitoring via WHOIS
- Web UI for domain management
- Google Chat webhook integration
- SQLite database
- Configurable monitoring intervals
- Multiple alert thresholds
- Property-based testing
- Graceful shutdown
- Error handling and retry logic
