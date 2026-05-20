// Package storage holds the SQLite-backed ledgers for the diagnostic and
// lab sides. They are intentionally separate databases so the lab side can
// stay clean of any tenant-derived data.
package storage

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Diagnostics holds opt-in tenant telemetry + public-complaint ingest.
type Diagnostics struct {
	db *sql.DB
	mu sync.Mutex
}

const diagSchema = `
CREATE TABLE IF NOT EXISTS telemetry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  tenant_hash TEXT,
  prompt_hash TEXT,
  model TEXT,
  prompt_tokens INTEGER,
  completion_tokens INTEGER,
  reasoning_tokens INTEGER,
  total_tokens INTEGER,
  retry_count INTEGER,
  task_class TEXT
);

CREATE TABLE IF NOT EXISTS complaints (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  source TEXT,           -- 'discord' | 'forum'
  thread_hash TEXT,
  redacted_excerpt TEXT,
  task_class TEXT
);

CREATE TABLE IF NOT EXISTS diagnostic_findings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  source_kind TEXT,      -- 'telemetry' | 'complaint'
  source_id INTEGER,
  task_class TEXT,
  waste_profile TEXT,
  severity REAL,
  evidence TEXT
);
`

// OpenDiagnostics opens the diagnostic DB.
func OpenDiagnostics(path string) (*Diagnostics, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(diagSchema); err != nil {
		return nil, err
	}
	return &Diagnostics{db: db}, nil
}

// Close releases the diagnostic db.
func (d *Diagnostics) Close() error { return d.db.Close() }

// AppendTelemetry persists one anonymised telemetry record.
func (d *Diagnostics) AppendTelemetry(
	ctx context.Context, tenantHash, promptHash, model, taskClass string,
	prompt, completion, reasoning, total, retry int,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.ExecContext(ctx, `INSERT INTO telemetry
		(ts, tenant_hash, prompt_hash, model,
		 prompt_tokens, completion_tokens, reasoning_tokens,
		 total_tokens, retry_count, task_class)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UnixMilli(), tenantHash, promptHash, model,
		prompt, completion, reasoning, total, retry, taskClass,
	)
	return err
}

// RecordAnalyserCall implements mimo.Ledger so the diagnostic side can
// account for its own MiMo classifier calls.
func (d *Diagnostics) RecordAnalyserCall(ctx context.Context, agent string, total int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, _ = d.db.ExecContext(ctx,
		`INSERT INTO diagnostic_findings (ts, source_kind, task_class, waste_profile, severity, evidence)
		 VALUES (?, 'analyser', ?, ?, 0, ?)`,
		time.Now().UnixMilli(), agent, "self", "analyser call recorded",
	)
}

// Lab holds synthetic-corpus runs and judge verdicts.
type Lab struct {
	db *sql.DB
	mu sync.Mutex
}

const labSchema = `
CREATE TABLE IF NOT EXISTS lab_probes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  task_id TEXT,
  task_class TEXT,
  domain TEXT,
  depth_pct INTEGER,
  output TEXT,
  total_tokens INTEGER,
  reasoning_tokens INTEGER
);

CREATE TABLE IF NOT EXISTS lab_findings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  task_id TEXT,
  task_class TEXT,
  min_viable_depth_pct INTEGER,
  quality_curve TEXT,
  failure_mode TEXT
);

CREATE TABLE IF NOT EXISTS analyser_calls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  agent TEXT,
  total_tokens INTEGER
);
`

// OpenLab opens the lab DB.
func OpenLab(path string) (*Lab, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(labSchema); err != nil {
		return nil, err
	}
	return &Lab{db: db}, nil
}

// Close releases the lab db.
func (l *Lab) Close() error { return l.db.Close() }

// AppendProbe persists one depth-graded probe.
func (l *Lab) AppendProbe(
	ctx context.Context,
	taskID, taskClass, domain string,
	depthPct, totalTokens, reasoningTokens int,
	output string,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.db.ExecContext(ctx, `INSERT INTO lab_probes
		(ts, task_id, task_class, domain, depth_pct, output, total_tokens, reasoning_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UnixMilli(), taskID, taskClass, domain,
		depthPct, output, totalTokens, reasoningTokens,
	)
	return err
}

// RecordAnalyserCall implements mimo.Ledger.
func (l *Lab) RecordAnalyserCall(ctx context.Context, agent string, total int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.db.ExecContext(ctx,
		`INSERT INTO analyser_calls (ts, agent, total_tokens) VALUES (?, ?, ?)`,
		time.Now().UnixMilli(), agent, total,
	)
}
