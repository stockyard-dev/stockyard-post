package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-post/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
	client *http.Client
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{
		db:     db,
		mux:    http.NewServeMux(),
		port:   port,
		limits: limits,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Forms CRUD
	s.mux.HandleFunc("POST /api/forms", s.handleCreateForm)
	s.mux.HandleFunc("GET /api/forms", s.handleListForms)
	s.mux.HandleFunc("GET /api/forms/{id}", s.handleGetForm)
	s.mux.HandleFunc("PUT /api/forms/{id}", s.handleUpdateForm)
	s.mux.HandleFunc("DELETE /api/forms/{id}", s.handleDeleteForm)

	// Submissions
	s.mux.HandleFunc("GET /api/forms/{id}/submissions", s.handleListSubmissions)
	s.mux.HandleFunc("GET /api/submissions/{id}", s.handleGetSubmission)
	s.mux.HandleFunc("DELETE /api/submissions/{id}", s.handleDeleteSubmission)

	// The form submission endpoint — HTML forms POST here
	s.mux.HandleFunc("POST /f/{id}", s.handleFormSubmit)

	// Export
	s.mux.HandleFunc("GET /api/forms/{id}/export", s.handleExport)

	// Status
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)

	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-post", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[post] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// --- Form submission (the hot path) ---

func (s *Server) handleFormSubmit(w http.ResponseWriter, r *http.Request) {
	formID := r.PathValue("id")

	form, err := s.db.GetForm(formID)
	if err != nil {
		http.Error(w, "Form not found", 404)
		return
	}
	if !form.Enabled {
		http.Error(w, "Form is disabled", 403)
		return
	}

	// CORS
	origin := r.Header.Get("Origin")
	if form.AllowedOrigins != "*" && origin != "" {
		allowed := false
		for _, o := range strings.Split(form.AllowedOrigins, ",") {
			if strings.TrimSpace(o) == origin {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "Origin not allowed", 403)
			return
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", form.AllowedOrigins)

	// Monthly limit
	if s.limits.MaxSubmissionsMonth > 0 {
		count, _ := s.db.MonthlySubmissionCount(formID)
		if LimitReached(s.limits.MaxSubmissionsMonth, count) {
			http.Error(w, "Monthly submission limit reached", 429)
			return
		}
	}

	// Parse form data
	contentType := r.Header.Get("Content-Type")
	data := make(map[string]string)

	if strings.HasPrefix(contentType, "application/json") {
		// JSON submission
		var jsonData map[string]any
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&jsonData); err == nil {
			for k, v := range jsonData {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
	} else {
		// URL-encoded form submission
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", 400)
			return
		}
		for k, v := range r.PostForm {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	// Honeypot check
	if form.HoneypotField != "" {
		if val, ok := data[form.HoneypotField]; ok && val != "" {
			// Bot detected — silently accept but don't store
			log.Printf("[post] honeypot triggered on form %s", formID)
			s.sendResponse(w, r, form, true)
			return
		}
		delete(data, form.HoneypotField) // Remove honeypot field from stored data
	}

	dataJSON, _ := json.Marshal(data)
	sourceIP := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		sourceIP = strings.Split(fwd, ",")[0]
	}

	sub, err := s.db.RecordSubmission(formID, string(dataJSON), sourceIP, r.UserAgent(), r.Referer())
	if err != nil {
		http.Error(w, "Failed to record submission", 500)
		return
	}

	log.Printf("[post] %s → form %s (%s) from %s", sub.ID, formID, form.Name, sourceIP)

	// Fire webhook asynchronously
	if form.WebhookURL != "" && s.limits.WebhookNotify {
		go s.fireWebhook(form, data, sub)
	}

	s.sendResponse(w, r, form, false)
}

func (s *Server) sendResponse(w http.ResponseWriter, r *http.Request, form *store.Form, isBot bool) {
	// Check if this is an AJAX request
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		writeJSON(w, 200, map[string]string{"status": "ok"})
		return
	}

	// Redirect
	if form.RedirectURL != "" {
		http.Redirect(w, r, form.RedirectURL, http.StatusSeeOther)
		return
	}

	// Default thank you page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(thankYouHTML))
}

func (s *Server) fireWebhook(form *store.Form, data map[string]string, sub *store.Submission) {
	payload := map[string]any{
		"form_id":       form.ID,
		"form_name":     form.Name,
		"submission_id": sub.ID,
		"data":          data,
		"source_ip":     sub.SourceIP,
		"submitted_at":  sub.CreatedAt,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", form.WebhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Post-Event", "submission")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[webhook] error sending to %s: %v", form.WebhookURL, err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// --- Form CRUD ---

func (s *Server) handleCreateForm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		RedirectURL string `json:"redirect_url"`
		EmailNotify string `json:"email_notify"`
		WebhookURL  string `json:"webhook_url"`
		Honeypot    string `json:"honeypot_field"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}

	if s.limits.MaxForms > 0 {
		forms, _ := s.db.ListForms()
		if LimitReached(s.limits.MaxForms, len(forms)) {
			writeJSON(w, 402, map[string]string{
				"error":   fmt.Sprintf("free tier limit: %d forms max — upgrade to Pro", s.limits.MaxForms),
				"upgrade": "https://stockyard.dev/post/",
			})
			return
		}
	}

	if req.WebhookURL != "" && !s.limits.WebhookNotify {
		writeJSON(w, 402, map[string]string{
			"error":   "webhook notifications require Pro — upgrade at https://stockyard.dev/post/",
			"upgrade": "https://stockyard.dev/post/",
		})
		return
	}

	form, err := s.db.CreateForm(req.Name, req.RedirectURL, req.EmailNotify, req.WebhookURL, req.Honeypot)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	submitURL := fmt.Sprintf("http://localhost:%d/f/%s", s.port, form.ID)
	writeJSON(w, 201, map[string]any{"form": form, "submit_url": submitURL})
}

func (s *Server) handleListForms(w http.ResponseWriter, r *http.Request) {
	forms, err := s.db.ListForms()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if forms == nil {
		forms = []store.Form{}
	}
	writeJSON(w, 200, map[string]any{"forms": forms, "count": len(forms)})
}

func (s *Server) handleGetForm(w http.ResponseWriter, r *http.Request) {
	form, err := s.db.GetForm(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "form not found"})
		return
	}
	submitURL := fmt.Sprintf("http://localhost:%d/f/%s", s.port, form.ID)
	writeJSON(w, 200, map[string]any{"form": form, "submit_url": submitURL})
}

func (s *Server) handleUpdateForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetForm(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "form not found"})
		return
	}
	var req struct {
		Name        *string `json:"name"`
		RedirectURL *string `json:"redirect_url"`
		EmailNotify *string `json:"email_notify"`
		WebhookURL  *string `json:"webhook_url"`
		Honeypot    *string `json:"honeypot_field"`
		Enabled     *bool   `json:"enabled"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	form, err := s.db.UpdateForm(id, req.Name, req.RedirectURL, req.EmailNotify, req.WebhookURL, req.Honeypot, req.Enabled)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"form": form})
}

func (s *Server) handleDeleteForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetForm(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "form not found"})
		return
	}
	s.db.DeleteForm(id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Submissions ---

func (s *Server) handleListSubmissions(w http.ResponseWriter, r *http.Request) {
	formID := r.PathValue("id")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	subs, err := s.db.ListSubmissions(formID, limit)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if subs == nil {
		subs = []store.Submission{}
	}
	writeJSON(w, 200, map[string]any{"submissions": subs, "count": len(subs)})
}

func (s *Server) handleGetSubmission(w http.ResponseWriter, r *http.Request) {
	sub, err := s.db.GetSubmission(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "submission not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"submission": sub})
}

func (s *Server) handleDeleteSubmission(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSubmission(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Export ---

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if !s.limits.ExportCSV {
		writeJSON(w, 402, map[string]string{
			"error":   "CSV export requires Pro — upgrade at https://stockyard.dev/post/",
			"upgrade": "https://stockyard.dev/post/",
		})
		return
	}
	formID := r.PathValue("id")
	subs, err := s.db.ListSubmissions(formID, 10000)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="submissions.json"`)
	writeJSON(w, 200, map[string]any{"submissions": subs, "count": len(subs)})
}

// --- Status ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func itoa(n int) string { return strconv.Itoa(n) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

const thankYouHTML = `<!DOCTYPE html><html><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Thank You</title>
<style>body{font-family:system-ui;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#1a1410;color:#f0e6d3}
.box{text-align:center;padding:3rem}.check{font-size:3rem;margin-bottom:1rem}h1{margin-bottom:.5rem}p{color:#bfb5a3}</style>
</head><body><div class="box"><div class="check">&#10003;</div><h1>Thank you!</h1><p>Your submission has been received.</p></div></body></html>`
