package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/domain-expiration-monitor/dem/internal/repository"
	"github.com/domain-expiration-monitor/dem/internal/scheduler"
	"github.com/domain-expiration-monitor/dem/internal/whois"
)

//go:embed templates/*
var templatesFS embed.FS

// Server represents the HTTP server
type Server struct {
	domainRepo  *repository.DomainRepository
	configRepo  *repository.ConfigRepository
	alertRepo   *repository.AlertRepository
	whoisSvc    *whois.Service
	scheduler   *scheduler.Scheduler
	templates   *template.Template
	mux         *http.ServeMux
}

// NewServer creates a new HTTP server
func NewServer(
	domainRepo *repository.DomainRepository,
	configRepo *repository.ConfigRepository,
	alertRepo *repository.AlertRepository,
	whoisSvc *whois.Service,
	sched *scheduler.Scheduler,
) (*Server, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
	
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		domainRepo: domainRepo,
		configRepo: configRepo,
		alertRepo:  alertRepo,
		whoisSvc:   whoisSvc,
		scheduler:  sched,
		templates:  tmpl,
		mux:        http.NewServeMux(),
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleDashboard)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/domains/", s.handleDomainDetail)
	s.mux.HandleFunc("/domains", s.handleDomains)
	s.mux.HandleFunc("/config", s.handleConfig)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Logging middleware
	log.Printf("%s %s", r.Method, r.URL.Path)
	s.mux.ServeHTTP(w, r)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, s)
}
