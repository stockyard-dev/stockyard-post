package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "post.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Conn() *sql.DB { return db.conn }
func (db *DB) Close() error  { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS forms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    redirect_url TEXT DEFAULT '',
    email_notify TEXT DEFAULT '',
    webhook_url TEXT DEFAULT '',
    honeypot_field TEXT DEFAULT '',
    allowed_origins TEXT DEFAULT '*',
    enabled INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS submissions (
    id TEXT PRIMARY KEY,
    form_id TEXT NOT NULL,
    data_json TEXT DEFAULT '{}',
    source_ip TEXT DEFAULT '',
    user_agent TEXT DEFAULT '',
    referer TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_submissions_form ON submissions(form_id);
CREATE INDEX IF NOT EXISTS idx_submissions_time ON submissions(created_at);
`)
	return err
}

// --- Form types ---

type Form struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	RedirectURL    string `json:"redirect_url"`
	EmailNotify    string `json:"email_notify"`
	WebhookURL     string `json:"webhook_url"`
	HoneypotField  string `json:"honeypot_field"`
	AllowedOrigins string `json:"allowed_origins"`
	Enabled        bool   `json:"enabled"`
	CreatedAt      string `json:"created_at"`
	Submissions    int    `json:"submission_count"`
}

func (db *DB) CreateForm(name, redirectURL, emailNotify, webhookURL, honeypot string) (*Form, error) {
	id := genID(8) // short ID for use in HTML forms
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec(`INSERT INTO forms (id,name,redirect_url,email_notify,webhook_url,honeypot_field,created_at)
		VALUES (?,?,?,?,?,?,?)`, id, name, redirectURL, emailNotify, webhookURL, honeypot, now)
	if err != nil {
		return nil, err
	}
	return &Form{ID: id, Name: name, RedirectURL: redirectURL, EmailNotify: emailNotify,
		WebhookURL: webhookURL, HoneypotField: honeypot, AllowedOrigins: "*", Enabled: true, CreatedAt: now}, nil
}

func (db *DB) ListForms() ([]Form, error) {
	rows, err := db.conn.Query(`SELECT f.id,f.name,f.redirect_url,f.email_notify,f.webhook_url,f.honeypot_field,
		f.allowed_origins,f.enabled,f.created_at,
		(SELECT COUNT(*) FROM submissions WHERE form_id=f.id)
		FROM forms f ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Form
	for rows.Next() {
		var f Form
		var en int
		rows.Scan(&f.ID, &f.Name, &f.RedirectURL, &f.EmailNotify, &f.WebhookURL, &f.HoneypotField,
			&f.AllowedOrigins, &en, &f.CreatedAt, &f.Submissions)
		f.Enabled = en == 1
		out = append(out, f)
	}
	return out, rows.Err()
}

func (db *DB) GetForm(id string) (*Form, error) {
	var f Form
	var en int
	err := db.conn.QueryRow(`SELECT f.id,f.name,f.redirect_url,f.email_notify,f.webhook_url,f.honeypot_field,
		f.allowed_origins,f.enabled,f.created_at,
		(SELECT COUNT(*) FROM submissions WHERE form_id=f.id)
		FROM forms f WHERE f.id=?`, id).
		Scan(&f.ID, &f.Name, &f.RedirectURL, &f.EmailNotify, &f.WebhookURL, &f.HoneypotField,
			&f.AllowedOrigins, &en, &f.CreatedAt, &f.Submissions)
	if err != nil {
		return nil, err
	}
	f.Enabled = en == 1
	return &f, nil
}

func (db *DB) UpdateForm(id string, name, redirectURL, emailNotify, webhookURL, honeypot *string, enabled *bool) (*Form, error) {
	if name != nil {
		db.conn.Exec("UPDATE forms SET name=? WHERE id=?", *name, id)
	}
	if redirectURL != nil {
		db.conn.Exec("UPDATE forms SET redirect_url=? WHERE id=?", *redirectURL, id)
	}
	if emailNotify != nil {
		db.conn.Exec("UPDATE forms SET email_notify=? WHERE id=?", *emailNotify, id)
	}
	if webhookURL != nil {
		db.conn.Exec("UPDATE forms SET webhook_url=? WHERE id=?", *webhookURL, id)
	}
	if honeypot != nil {
		db.conn.Exec("UPDATE forms SET honeypot_field=? WHERE id=?", *honeypot, id)
	}
	if enabled != nil {
		en := 0
		if *enabled {
			en = 1
		}
		db.conn.Exec("UPDATE forms SET enabled=? WHERE id=?", en, id)
	}
	return db.GetForm(id)
}

func (db *DB) DeleteForm(id string) error {
	db.conn.Exec("DELETE FROM submissions WHERE form_id=?", id)
	_, err := db.conn.Exec("DELETE FROM forms WHERE id=?", id)
	return err
}

// --- Submissions ---

type Submission struct {
	ID        string `json:"id"`
	FormID    string `json:"form_id"`
	Data      string `json:"data"`
	SourceIP  string `json:"source_ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) RecordSubmission(formID, dataJSON, sourceIP, userAgent, referer string) (*Submission, error) {
	id := "sub_" + genID(10)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec(`INSERT INTO submissions (id,form_id,data_json,source_ip,user_agent,referer,created_at)
		VALUES (?,?,?,?,?,?,?)`, id, formID, dataJSON, sourceIP, userAgent, referer, now)
	if err != nil {
		return nil, err
	}
	return &Submission{ID: id, FormID: formID, Data: dataJSON, SourceIP: sourceIP,
		UserAgent: userAgent, Referer: referer, CreatedAt: now}, nil
}

func (db *DB) ListSubmissions(formID string, limit int) ([]Submission, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := db.conn.Query(`SELECT id,form_id,data_json,source_ip,user_agent,referer,created_at
		FROM submissions WHERE form_id=? ORDER BY created_at DESC LIMIT ?`, formID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Submission
	for rows.Next() {
		var s Submission
		rows.Scan(&s.ID, &s.FormID, &s.Data, &s.SourceIP, &s.UserAgent, &s.Referer, &s.CreatedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) GetSubmission(id string) (*Submission, error) {
	var s Submission
	err := db.conn.QueryRow(`SELECT id,form_id,data_json,source_ip,user_agent,referer,created_at
		FROM submissions WHERE id=?`, id).
		Scan(&s.ID, &s.FormID, &s.Data, &s.SourceIP, &s.UserAgent, &s.Referer, &s.CreatedAt)
	return &s, err
}

func (db *DB) DeleteSubmission(id string) error {
	_, err := db.conn.Exec("DELETE FROM submissions WHERE id=?", id)
	return err
}

func (db *DB) MonthlySubmissionCount(formID string) (int, error) {
	cutoff := time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM submissions WHERE form_id=? AND created_at>=?", formID, cutoff).Scan(&count)
	return count, err
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var forms, submissions int
	db.conn.QueryRow("SELECT COUNT(*) FROM forms").Scan(&forms)
	db.conn.QueryRow("SELECT COUNT(*) FROM submissions").Scan(&submissions)
	return map[string]any{"forms": forms, "submissions": submissions}
}

func (db *DB) Cleanup(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec("DELETE FROM submissions WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
