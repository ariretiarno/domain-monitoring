package repository

// SQLite schema
const schema = `
CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    expiration_date DATETIME NOT NULL,
    nameservers TEXT NOT NULL,
    registrant TEXT NOT NULL,
    registrar TEXT NOT NULL,
    last_checked DATETIME NOT NULL,
    next_check DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);
CREATE INDEX IF NOT EXISTS idx_domains_expiration_date ON domains(expiration_date);
CREATE INDEX IF NOT EXISTS idx_domains_next_check ON domains(next_check);

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    monitoring_interval INTEGER NOT NULL,
    alert_thresholds TEXT NOT NULL,
    google_chat_webhook TEXT NOT NULL,
    retention_period INTEGER NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS alerts (
    id TEXT PRIMARY KEY,
    domain_id TEXT NOT NULL,
    domain_name TEXT NOT NULL,
    threshold INTEGER NOT NULL,
    expiration_date DATETIME NOT NULL,
    sent_at DATETIME NOT NULL,
    success INTEGER NOT NULL,
    error_message TEXT NOT NULL,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_domain_id ON alerts(domain_id);
CREATE INDEX IF NOT EXISTS idx_alerts_sent_at ON alerts(sent_at);
`


// MySQL schema
const schemaMySQL = `
CREATE TABLE IF NOT EXISTS domains (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    expiration_date DATETIME NOT NULL,
    nameservers JSON NOT NULL,
    registrant TEXT NOT NULL,
    registrar VARCHAR(255) NOT NULL,
    last_checked DATETIME NOT NULL,
    next_check DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_domains_name (name),
    INDEX idx_domains_expiration_date (expiration_date),
    INDEX idx_domains_next_check (next_check)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY,
    monitoring_interval BIGINT NOT NULL,
    alert_thresholds JSON NOT NULL,
    google_chat_webhook TEXT NOT NULL,
    retention_period BIGINT NOT NULL,
    updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    domain_id VARCHAR(255) NOT NULL,
    domain_name VARCHAR(255) NOT NULL,
    threshold BIGINT NOT NULL,
    expiration_date DATETIME NOT NULL,
    sent_at DATETIME NOT NULL,
    success TINYINT(1) NOT NULL,
    error_message TEXT NOT NULL,
    INDEX idx_alerts_domain_id (domain_id),
    INDEX idx_alerts_sent_at (sent_at),
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
