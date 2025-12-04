package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domain-expiration-monitor/dem/internal/domain"
	"github.com/domain-expiration-monitor/dem/internal/repository"
)

// Service handles alert evaluation and sending
type Service struct {
	alertRepo  *repository.AlertRepository
	configRepo *repository.ConfigRepository
	httpClient *http.Client
}

// NewService creates a new alert service
func NewService(alertRepo *repository.AlertRepository, configRepo *repository.ConfigRepository) *Service {
	return &Service{
		alertRepo:  alertRepo,
		configRepo: configRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// EvaluateAlerts checks if any alert thresholds are crossed for a domain
func (s *Service) EvaluateAlerts(d *domain.Domain) error {
	config, err := s.configRepo.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	thresholds := config.GetAlertThresholds()
	timeUntilExpiration := time.Until(d.ExpirationDate)

	for _, threshold := range thresholds {
		// Check if we're within the threshold
		if timeUntilExpiration <= threshold && timeUntilExpiration > 0 {
			// Check if alert already sent
			alreadySent, err := s.alertRepo.HasAlertBeenSent(d.ID, threshold)
			if err != nil {
				return fmt.Errorf("failed to check if alert was sent: %w", err)
			}

			if !alreadySent {
				// Create and send alert
				alert := &domain.Alert{
					DomainID:       d.ID,
					DomainName:     d.Name,
					ExpirationDate: d.ExpirationDate,
					SentAt:         time.Now(),
				}
				alert.SetThreshold(threshold)

				if err := s.SendAlert(alert, config.GoogleChatWebhook); err != nil {
					alert.Success = false
					alert.ErrorMessage = err.Error()
				} else {
					alert.Success = true
				}

				// Save alert record
				if err := s.alertRepo.Create(alert); err != nil {
					return fmt.Errorf("failed to save alert: %w", err)
				}
			}
		}
	}

	return nil
}

// SendAlert sends an alert to Google Chat with retry logic
func (s *Service) SendAlert(alert *domain.Alert, webhookURL string) error {
	if webhookURL == "" {
		// No webhook configured, just log
		return fmt.Errorf("no webhook URL configured")
	}

	message := s.FormatAlertMessage(alert)

	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt < 3; attempt++ {
		err := s.sendToWebhook(webhookURL, message)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < 2 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// sendToWebhook sends a message to a Google Chat webhook
func (s *Service) sendToWebhook(webhookURL string, message string) error {
	payload := map[string]interface{}{
		"text": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := s.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// FormatAlertMessage creates a human-readable alert message
func (s *Service) FormatAlertMessage(alert *domain.Alert) string {
	daysRemaining := alert.DaysUntilExpiration()
	thresholdDays := int(alert.GetThreshold().Hours() / 24)

	return fmt.Sprintf(
		"🔔 Domain Expiration Alert\n\n"+
			"Domain: %s\n"+
			"Expiration Date: %s\n"+
			"Days Remaining: %d\n"+
			"Alert Threshold: %d days\n\n"+
			"Please renew this domain to avoid service disruption.",
		alert.DomainName,
		alert.ExpirationDate.Format("2006-01-02"),
		daysRemaining,
		thresholdDays,
	)
}
