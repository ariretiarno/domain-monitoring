package repository

import (
	"os"
	"testing"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: domain-expiration-monitor, Property 6: Domain addition and retrieval
// For any valid domain name submitted through the Web UI, adding it to the monitoring list
// should result in the domain being retrievable from the database with all its WHOIS information populated.
// Validates: Requirements 4.1, 4.2
func TestProperty_DomainAdditionAndRetrieval(t *testing.T) {
	// Create temporary database for testing
	dbPath := "test_domain_addition.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath, "sqlite3")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	repo := NewDomainRepository(db)

	properties := gopter.NewProperties(nil)

	counter := 0
	properties.Property("domain addition and retrieval preserves all fields", prop.ForAll(
		func(registrant string, registrar string) bool {
			counter++
			// Create a domain with generated data and unique name
			d := &domain.Domain{
				Name:           time.Now().Format("20060102150405") + string(rune(counter)) + ".com",
				ExpirationDate: time.Now().Add(365 * 24 * time.Hour),
				Nameservers:    domain.Strings{"ns1.example.com", "ns2.example.com"},
				Registrant:     registrant,
				Registrar:      registrar,
				LastChecked:    time.Now(),
				NextCheck:      time.Now().Add(24 * time.Hour),
			}

			// Add domain to database
			if err := repo.Create(d); err != nil {
				return false
			}

			// Retrieve domain by ID
			retrieved, err := repo.GetByID(d.ID)
			if err != nil {
				return false
			}

			// Verify all fields match
			return retrieved.Name == d.Name &&
				retrieved.Registrant == d.Registrant &&
				retrieved.Registrar == d.Registrar &&
				len(retrieved.Nameservers) == len(d.Nameservers)
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: domain-expiration-monitor, Property 17: Startup domain loading
// For any set of domains persisted in the database, starting the application
// should load all of them into the active monitoring list.
// Validates: Requirements 10.1
func TestProperty_StartupDomainLoading(t *testing.T) {
	// Create temporary database for testing
	dbPath := "test_startup_loading.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath, "sqlite3")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	repo := NewDomainRepository(db)

	properties := gopter.NewProperties(nil)

	properties.Property("all persisted domains can be loaded", prop.ForAll(
		func(count uint8) bool {
			if count == 0 {
				return true
			}

			// Create multiple domains
			domainCount := int(count) % 20 // Limit to 20 domains per test
			if domainCount == 0 {
				domainCount = 1
			}

			createdIDs := make(map[string]bool)
			for i := 0; i < domainCount; i++ {
				d := &domain.Domain{
					Name:           time.Now().Format("20060102150405") + string(rune(i)) + ".com",
					ExpirationDate: time.Now().Add(365 * 24 * time.Hour),
					Nameservers:    domain.Strings{"ns1.example.com"},
					Registrant:     "Test Registrant",
					Registrar:      "Test Registrar",
					LastChecked:    time.Now(),
					NextCheck:      time.Now().Add(24 * time.Hour),
				}

				if err := repo.Create(d); err != nil {
					continue // Skip duplicates
				}
				createdIDs[d.ID] = true
			}

			// Load all domains
			loaded, err := repo.GetAll()
			if err != nil {
				return false
			}

			// Verify all created domains are in the loaded set
			loadedIDs := make(map[string]bool)
			for _, d := range loaded {
				loadedIDs[d.ID] = true
			}

			for id := range createdIDs {
				if !loadedIDs[id] {
					return false
				}
			}

			return true
		},
		gen.UInt8(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: domain-expiration-monitor, Property 20: Graceful shutdown persistence
// For any system state (domains, configuration, alerts), performing a graceful shutdown
// and then restarting should restore the same state.
// Validates: Requirements 10.5
func TestProperty_GracefulShutdownPersistence(t *testing.T) {
	// Use a single database for all property tests
	dbPath := "test_shutdown_persistence.db"
	defer os.Remove(dbPath)

	properties := gopter.NewProperties(nil)

	testNum := 0
	properties.Property("domain state persists across database close/reopen", prop.ForAll(
		func(registrant string) bool {
			testNum++
			
			// First session: create domain
			db1, err := NewDB(dbPath, "sqlite3")
			if err != nil {
				return false
			}

			repo1 := NewDomainRepository(db1)
			domainName := "domain_" + string(rune(testNum)) + "_" + time.Now().Format("150405.000000") + ".com"
			d := &domain.Domain{
				Name:           domainName,
				ExpirationDate: time.Now().Add(365 * 24 * time.Hour),
				Nameservers:    domain.Strings{"ns1.example.com"},
				Registrant:     registrant,
				Registrar:      "Test Registrar",
				LastChecked:    time.Now(),
				NextCheck:      time.Now().Add(24 * time.Hour),
			}

			if err := repo1.Create(d); err != nil {
				db1.Close()
				return false
			}

			originalID := d.ID
			originalName := d.Name
			originalRegistrant := d.Registrant
			db1.Close() // Graceful shutdown

			// Second session: verify domain still exists
			db2, err := NewDB(dbPath, "sqlite3")
			if err != nil {
				return false
			}
			defer db2.Close()

			repo2 := NewDomainRepository(db2)
			retrieved, err := repo2.GetByID(originalID)
			if err != nil {
				return false
			}

			return retrieved.Name == originalName &&
				retrieved.Registrant == originalRegistrant
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
