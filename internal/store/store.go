package store

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type DB struct{ db *sql.DB }
type Article struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Author      string `json:"author"`
	Slug        string `json:"slug"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at"`
	CreatedAt   string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "post.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS articles(id TEXT PRIMARY KEY,title TEXT NOT NULL,body TEXT DEFAULT '',author TEXT DEFAULT '',slug TEXT DEFAULT '',category TEXT DEFAULT '',status TEXT DEFAULT 'draft',published_at TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
	resource TEXT NOT NULL,
	record_id TEXT NOT NULL,
	data TEXT NOT NULL DEFAULT '{}',
	PRIMARY KEY(resource, record_id)
)`)
	return &DB{db: db}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }
func (d *DB) Create(e *Article) error {
	e.ID = genID()
	e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO articles(id,title,body,author,slug,category,status,published_at,created_at)VALUES(?,?,?,?,?,?,?,?,?)`, e.ID, e.Title, e.Body, e.Author, e.Slug, e.Category, e.Status, e.PublishedAt, e.CreatedAt)
	return err
}
func (d *DB) Get(id string) *Article {
	var e Article
	if d.db.QueryRow(`SELECT id,title,body,author,slug,category,status,published_at,created_at FROM articles WHERE id=?`, id).Scan(&e.ID, &e.Title, &e.Body, &e.Author, &e.Slug, &e.Category, &e.Status, &e.PublishedAt, &e.CreatedAt) != nil {
		return nil
	}
	return &e
}
func (d *DB) List() []Article {
	rows, _ := d.db.Query(`SELECT id,title,body,author,slug,category,status,published_at,created_at FROM articles ORDER BY created_at DESC`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Article
	for rows.Next() {
		var e Article
		rows.Scan(&e.ID, &e.Title, &e.Body, &e.Author, &e.Slug, &e.Category, &e.Status, &e.PublishedAt, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}
func (d *DB) Update(e *Article) error {
	_, err := d.db.Exec(`UPDATE articles SET title=?,body=?,author=?,slug=?,category=?,status=?,published_at=? WHERE id=?`, e.Title, e.Body, e.Author, e.Slug, e.Category, e.Status, e.PublishedAt, e.ID)
	return err
}
func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM articles WHERE id=?`, id)
	return err
}
func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&n)
	return n
}

func (d *DB) Search(q string, filters map[string]string) []Article {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (title LIKE ? OR body LIKE ? OR slug LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["category"]; ok && v != "" {
		where += " AND category=?"
		args = append(args, v)
	}
	if v, ok := filters["status"]; ok && v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	rows, _ := d.db.Query(`SELECT id,title,body,author,slug,category,status,published_at,created_at FROM articles WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Article
	for rows.Next() {
		var e Article
		rows.Scan(&e.ID, &e.Title, &e.Body, &e.Author, &e.Slug, &e.Category, &e.Status, &e.PublishedAt, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

func (d *DB) Stats() map[string]any {
	m := map[string]any{"total": d.Count()}
	rows, _ := d.db.Query(`SELECT status,COUNT(*) FROM articles GROUP BY status`)
	if rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_status"] = by
	}
	return m
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
