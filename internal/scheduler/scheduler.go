package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/alert"
	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/domain-expiration-monitor/dem/internal/repository"
	"github.com/domain-expiration-monitor/dem/internal/whois"
)

// Scheduler manages periodic WHOIS checks for domains
type Scheduler struct {
	domainRepo  *repository.DomainRepository
	configRepo  *repository.ConfigRepository
	whoisSvc    *whois.Service
	alertSvc    *alert.Service
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	workerPool  chan struct{}
	mu          sync.RWMutex
	scheduledDomains map[string]*time.Timer
}

// NewScheduler creates a new scheduler
func NewScheduler(
	domainRepo *repository.DomainRepository,
	configRepo *repository.ConfigRepository,
	whoisSvc *whois.Service,
	alertSvc *alert.Service,
) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		domainRepo:       domainRepo,
		configRepo:       configRepo,
		whoisSvc:         whoisSvc,
		alertSvc:         alertSvc,
		ctx:              ctx,
		cancel:           cancel,
		workerPool:       make(chan struct{}, 10), // 10 concurrent workers
		scheduledDomains: make(map[string]*time.Timer),
	}
}

// Start initializes and starts the scheduler
func (s *Scheduler) Start() error {
	// Load all domains
	domains, err := s.domainRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to load domains: %w", err)
	}

	// Schedule each domain
	for _, d := range domains {
		s.ScheduleDomain(d)
	}

	return nil
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() error {
	s.cancel()
	
	// Cancel all timers
	s.mu.Lock()
	for _, timer := range s.scheduledDomains {
		timer.Stop()
	}
	s.mu.Unlock()

	// Wait for all workers to finish (with timeout)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for workers to finish")
	}
}

// ScheduleDomain adds a domain to the monitoring schedule
func (s *Scheduler) ScheduleDomain(d *domain.Domain) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel existing timer if any
	if timer, exists := s.scheduledDomains[d.ID]; exists {
		timer.Stop()
	}

	// Calculate time until next check
	var delay time.Duration
	if time.Now().Before(d.NextCheck) {
		delay = time.Until(d.NextCheck)
	} else {
		delay = 0 // Check immediately
	}

	// Schedule the check
	timer := time.AfterFunc(delay, func() {
		s.checkDomain(d.ID)
	})

	s.scheduledDomains[d.ID] = timer
}

// UnscheduleDomain removes a domain from the monitoring schedule
func (s *Scheduler) UnscheduleDomain(domainID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timer, exists := s.scheduledDomains[domainID]; exists {
		timer.Stop()
		delete(s.scheduledDomains, domainID)
	}
}

// checkDomain performs a WHOIS check for a domain
func (s *Scheduler) checkDomain(domainID string) {
	// Acquire worker slot
	select {
	case s.workerPool <- struct{}{}:
		defer func() { <-s.workerPool }()
	case <-s.ctx.Done():
		return
	}

	s.wg.Add(1)
	defer s.wg.Done()

	// Get domain
	d, err := s.domainRepo.GetByID(domainID)
	if err != nil {
		// Domain might have been deleted
		return
	}

	// Get config for scheduling
	config, err := s.configRepo.Get()
	if err != nil {
		s.reschedule(d)
		return
	}

	// Perform WHOIS query
	info, err := s.whoisSvc.QueryDomain(d.Name)
	if err != nil {
		// WHOIS failed, but still evaluate alerts with existing data
		d.LastChecked = time.Now()
		d.NextCheck = time.Now().Add(config.GetMonitoringInterval())
		
		// Save updated check times
		if err := s.domainRepo.Update(d); err != nil {
			s.reschedule(d)
			return
		}
		
		// Evaluate alerts with existing expiration date
		if err := s.alertSvc.EvaluateAlerts(d); err != nil {
			// Log error but continue
		}
		
		s.reschedule(d)
		return
	}

	// Update domain with new WHOIS data
	d.ExpirationDate = info.ExpirationDate
	d.Nameservers = domain.Strings(info.Nameservers)
	d.Registrant = info.Registrant
	d.Registrar = info.Registrar
	d.LastChecked = time.Now()
	d.NextCheck = time.Now().Add(config.GetMonitoringInterval())

	// Save updated domain
	if err := s.domainRepo.Update(d); err != nil {
		s.reschedule(d)
		return
	}

	// Evaluate alerts
	if err := s.alertSvc.EvaluateAlerts(d); err != nil {
		// Log error but continue
	}

	// Reschedule next check
	s.reschedule(d)
}

// reschedule schedules the next check for a domain
func (s *Scheduler) reschedule(d *domain.Domain) {
	select {
	case <-s.ctx.Done():
		return
	default:
		s.ScheduleDomain(d)
	}
}
