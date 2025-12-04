package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/domain-expiration-monitor/dem/internal/alert"
	"github.com/domain-expiration-monitor/dem/internal/repository"
	"github.com/domain-expiration-monitor/dem/internal/scheduler"
	"github.com/domain-expiration-monitor/dem/internal/web"
	"github.com/domain-expiration-monitor/dem/internal/whois"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()
	
	log.Println("Domain Expiration Monitor starting...")

	// Initialize database
	dbDriver := getEnv("DB_DRIVER", "sqlite3")
	var dbPath string
	
	if dbDriver == "mysql" {
		// Build MySQL connection string from environment variables
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "3306")
		dbName := getEnv("DB_NAME", "dem")
		dbUser := getEnv("DB_USER", "root")
		dbPassword := getEnv("DB_PASSWORD", "")
		
		dbPath = dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true&charset=utf8mb4"
		log.Printf("Connecting to MySQL database at %s:%s/%s...", dbHost, dbPort, dbName)
	} else {
		// SQLite
		dbPath = getEnv("DB_PATH", "dem.db")
		log.Printf("Connecting to SQLite database at %s...", dbPath)
	}
	
	db, err := repository.NewDB(dbPath, dbDriver)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database connected successfully")

	// Initialize repositories
	domainRepo := repository.NewDomainRepository(db)
	configRepo := repository.NewConfigRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	// Initialize services
	whoisSvc := whois.NewService()
	alertSvc := alert.NewService(alertRepo, configRepo)

	// Initialize scheduler
	sched := scheduler.NewScheduler(domainRepo, configRepo, whoisSvc, alertSvc)

	// Load all domains and start scheduler
	if err := sched.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Initialize web server
	server, err := web.NewServer(domainRepo, configRepo, alertRepo, whoisSvc, sched)
	if err != nil {
		log.Fatalf("Failed to initialize web server: %v", err)
	}

	// Start web server in goroutine
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	go func() {
		log.Printf("Starting web server on %s", httpAddr)
		if err := server.Start(httpAddr); err != nil {
			log.Fatalf("Web server error: %v", err)
		}
	}()

	log.Println("Domain Expiration Monitor initialized successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Stop scheduler
	if err := sched.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	log.Println("Shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
