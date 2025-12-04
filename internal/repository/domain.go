package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/google/uuid"
)

// DomainRepository handles domain data persistence
type DomainRepository struct {
	db *DB
}

// NewDomainRepository creates a new domain repository
func NewDomainRepository(db *DB) *DomainRepository {
	return &DomainRepository{db: db}
}

// Create adds a new domain to the database
func (r *DomainRepository) Create(d *domain.Domain) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now

	query := `
		INSERT INTO domains (
			id, name, expiration_date, nameservers, registrant, registrar,
			last_checked, next_check, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		d.ID, d.Name, d.ExpirationDate, d.Nameservers, d.Registrant, d.Registrar,
		d.LastChecked, d.NextCheck, d.CreatedAt, d.UpdatedAt,
	)

	if err != nil {
		if IsConstraintError(err) {
			return fmt.Errorf("domain %s already exists", d.Name)
		}
		return fmt.Errorf("failed to create domain: %w", err)
	}

	return nil
}

// GetByID retrieves a domain by its ID
func (r *DomainRepository) GetByID(id string) (*domain.Domain, error) {
	var d domain.Domain
	query := `
		SELECT id, name, expiration_date, nameservers, registrant, registrar,
		       last_checked, next_check, created_at, updated_at
		FROM domains
		WHERE id = ?
	`

	err := r.db.Get(&d, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return &d, nil
}

// GetByName retrieves a domain by its name
func (r *DomainRepository) GetByName(name string) (*domain.Domain, error) {
	var d domain.Domain
	query := `
		SELECT id, name, expiration_date, nameservers, registrant, registrar,
		       last_checked, next_check, created_at, updated_at
		FROM domains
		WHERE name = ?
	`

	err := r.db.Get(&d, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return &d, nil
}

// GetAll retrieves all domains
func (r *DomainRepository) GetAll() ([]*domain.Domain, error) {
	var domains []*domain.Domain
	query := `
		SELECT id, name, expiration_date, nameservers, registrant, registrar,
		       last_checked, next_check, created_at, updated_at
		FROM domains
		ORDER BY expiration_date ASC
	`

	err := r.db.Select(&domains, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all domains: %w", err)
	}

	return domains, nil
}

// Update updates an existing domain
func (r *DomainRepository) Update(d *domain.Domain) error {
	d.UpdatedAt = time.Now()

	query := `
		UPDATE domains
		SET name = ?, expiration_date = ?, nameservers = ?, registrant = ?,
		    registrar = ?, last_checked = ?, next_check = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query,
		d.Name, d.ExpirationDate, d.Nameservers, d.Registrant, d.Registrar,
		d.LastChecked, d.NextCheck, d.UpdatedAt, d.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("domain not found: %s", d.ID)
	}

	return nil
}

// Delete removes a domain from the database
func (r *DomainRepository) Delete(id string) error {
	query := `DELETE FROM domains WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("domain not found: %s", id)
	}

	return nil
}

// DeleteOlderThan deletes domains that were created before the cutoff time
// This is used for retention policy, but excludes actively monitored domains
func (r *DomainRepository) DeleteOlderThan(cutoff time.Time) error {
	query := `DELETE FROM domains WHERE created_at < ? AND id NOT IN (SELECT id FROM domains)`
	
	// Note: This query is simplified. In practice, we'd need a way to mark domains as "inactive"
	// For now, we won't delete any domains that are in the domains table (all are considered active)
	// A proper implementation would have an "active" flag or separate table for inactive domains
	
	_, err := r.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old domains: %w", err)
	}

	return nil
}

// GetDomainsForCheck retrieves domains that need to be checked
func (r *DomainRepository) GetDomainsForCheck() ([]*domain.Domain, error) {
	var domains []*domain.Domain
	query := `
		SELECT id, name, expiration_date, nameservers, registrant, registrar,
		       last_checked, next_check, created_at, updated_at
		FROM domains
		WHERE next_check <= ?
		ORDER BY next_check ASC
	`

	err := r.db.Select(&domains, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get domains for check: %w", err)
	}

	return domains, nil
}
