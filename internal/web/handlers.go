package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
)

// handleHealth returns the health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDashboard displays the main dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	domains, err := s.domainRepo.GetAll()
	if err != nil {
		s.renderError(w, "Failed to load domains", err, http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Domains": domains,
		"Now":     time.Now(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
		s.renderError(w, "Failed to render template", err, http.StatusInternalServerError)
	}
}

// handleDomainDetail displays details for a specific domain
func (s *Server) handleDomainDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/domains/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	d, err := s.domainRepo.GetByID(id)
	if err != nil {
		s.renderError(w, "Domain not found", err, http.StatusNotFound)
		return
	}

	alerts, err := s.alertRepo.GetByDomainID(id)
	if err != nil {
		alerts = []*domain.Alert{}
	}

	data := map[string]interface{}{
		"Domain": d,
		"Alerts": alerts,
		"Now":    time.Now(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "domain-detail", data); err != nil {
		s.renderError(w, "Failed to render template", err, http.StatusInternalServerError)
	}
}

// handleDomains handles domain management (add/delete)
func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleAddDomain(w, r)
	case http.MethodDelete:
		s.handleDeleteDomain(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAddDomain adds a new domain
func (s *Server) handleAddDomain(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", err, http.StatusBadRequest)
		return
	}

	domainName := r.FormValue("domain")
	if domainName == "" {
		s.renderError(w, "Domain name is required", nil, http.StatusBadRequest)
		return
	}

	// Perform immediate WHOIS query
	info, err := s.whoisSvc.QueryDomain(domainName)
	if err != nil {
		s.renderError(w, "Failed to query domain", err, http.StatusBadRequest)
		return
	}

	// Create domain
	d := &domain.Domain{
		Name:           domainName,
		ExpirationDate: info.ExpirationDate,
		Nameservers:    domain.Strings(info.Nameservers),
		Registrant:     info.Registrant,
		Registrar:      info.Registrar,
		LastChecked:    time.Now(),
		NextCheck:      time.Now().Add(24 * time.Hour),
	}

	if err := s.domainRepo.Create(d); err != nil {
		s.renderError(w, "Failed to add domain", err, http.StatusInternalServerError)
		return
	}

	// Schedule monitoring
	s.scheduler.ScheduleDomain(d)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleDeleteDomain removes a domain
func (s *Server) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		s.renderError(w, "Domain ID is required", nil, http.StatusBadRequest)
		return
	}

	// Unschedule monitoring
	s.scheduler.UnscheduleDomain(id)

	// Delete domain
	if err := s.domainRepo.Delete(id); err != nil {
		s.renderError(w, "Failed to delete domain", err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleConfig handles configuration management
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPost:
		s.handleUpdateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetConfig displays the configuration page
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.configRepo.Get()
	if err != nil {
		s.renderError(w, "Failed to load configuration", err, http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Config": config,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "config-page", data); err != nil {
		s.renderError(w, "Failed to render template", err, http.StatusInternalServerError)
	}
}

// handleUpdateConfig updates the configuration
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", err, http.StatusBadRequest)
		return
	}

	config, err := s.configRepo.Get()
	if err != nil {
		s.renderError(w, "Failed to load configuration", err, http.StatusInternalServerError)
		return
	}

	// Parse and validate monitoring interval
	intervalHours := r.FormValue("monitoring_interval")
	if intervalHours != "" {
		var hours int
		fmt.Sscanf(intervalHours, "%d", &hours)
		if hours < 1 {
			s.renderError(w, "Monitoring interval must be at least 1 hour", nil, http.StatusBadRequest)
			return
		}
		config.SetMonitoringInterval(time.Duration(hours) * time.Hour)
	}

	// Parse webhook URL
	webhook := r.FormValue("webhook_url")
	if webhook != "" && !strings.HasPrefix(webhook, "https://") {
		s.renderError(w, "Webhook URL must use HTTPS", nil, http.StatusBadRequest)
		return
	}
	config.GoogleChatWebhook = webhook

	// Parse retention period
	retentionDays := r.FormValue("retention_period")
	if retentionDays != "" {
		var days int
		fmt.Sscanf(retentionDays, "%d", &days)
		if days < 1 {
			s.renderError(w, "Retention period must be at least 1 day", nil, http.StatusBadRequest)
			return
		}
		config.SetRetentionPeriod(time.Duration(days) * 24 * time.Hour)
	}

	// Parse alert thresholds
	thresholdsStr := r.FormValue("alert_thresholds")
	if thresholdsStr != "" {
		thresholdDays := strings.Split(thresholdsStr, ",")
		thresholds := make([]time.Duration, 0, len(thresholdDays))
		
		for _, dayStr := range thresholdDays {
			dayStr = strings.TrimSpace(dayStr)
			if dayStr == "" {
				continue
			}
			
			var days int
			_, err := fmt.Sscanf(dayStr, "%d", &days)
			if err != nil || days <= 0 {
				s.renderError(w, fmt.Sprintf("Invalid threshold value: %s", dayStr), nil, http.StatusBadRequest)
				return
			}
			
			thresholds = append(thresholds, time.Duration(days)*24*time.Hour)
		}
		
		if len(thresholds) == 0 {
			s.renderError(w, "At least one alert threshold is required", nil, http.StatusBadRequest)
			return
		}
		
		config.SetAlertThresholds(thresholds)
	}

	// Update configuration
	if err := s.configRepo.Update(config); err != nil {
		s.renderError(w, "Failed to update configuration", err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/config", http.StatusSeeOther)
}

// renderError renders an error page
func (s *Server) renderError(w http.ResponseWriter, message string, err error, statusCode int) {
	w.WriteHeader(statusCode)
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}
	fmt.Fprintf(w, "<html><body><h1>Error</h1><p>%s</p></body></html>", errorMsg)
}
