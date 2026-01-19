package handlers

import (
	"database/sql"
	"docuflow/models"
	"html/template"
	"net/http"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

type DocumentHandler struct {
	DB *sql.DB
}

func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// Mock user ID from session (real app would parse cookie/context)
	// For now, assume a dummy user ID "1" if logged in.
	// We need Middleware for auth, but strictly following immediate request...

	rows, err := h.DB.Query("SELECT id, title, updated_at FROM documents ORDER BY updated_at DESC")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.Title, &d.UpdatedAt); err != nil {
			continue
		}
		docs = append(docs, d)
	}

	tmpl := template.Must(template.ParseFiles("web/templates/base.html", "web/templates/document_list.html"))
	tmpl.Execute(w, struct {
		User      string
		Documents []models.Document
	}{
		User:      GetBaseData(r).User,
		Documents: docs,
	})
}

func (h *DocumentHandler) NewDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("web/templates/base.html", "web/templates/document_edit.html"))
		tmpl.Execute(w, GetBaseData(r))
		return
	}

	// Create
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Hardcoded owner for now
	ownerID := 1

	res, err := h.DB.Exec("INSERT INTO documents (title, content, owner_id) VALUES (?, ?, ?)", title, content, ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	http.Redirect(w, r, "/documents/view?id="+string(rune(id)), http.StatusSeeOther) // Simplistic redirect
}

func (h *DocumentHandler) ViewDocument(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var doc models.Document
	err := h.DB.QueryRow("SELECT id, title, content FROM documents WHERE id = ?", id).Scan(&doc.ID, &doc.Title, &doc.Content)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Render Markdown
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	docContent := []byte(doc.Content)
	htmlBytes := markdown.ToHTML(docContent, p, nil)

	tmpl := template.Must(template.ParseFiles("web/templates/base.html", "web/templates/document_view.html"))
	tmpl.Execute(w, struct {
		User     string
		Document models.Document
		Content  template.HTML
	}{
		User:     GetBaseData(r).User,
		Document: doc,
		Content:  template.HTML(htmlBytes),
	})
}

func (h *DocumentHandler) EditDocument(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if r.Method == "GET" {
		var doc models.Document
		err := h.DB.QueryRow("SELECT id, title, content FROM documents WHERE id = ?", id).Scan(&doc.ID, &doc.Title, &doc.Content)
		if err != nil {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		tmpl := template.Must(template.ParseFiles("web/templates/base.html", "web/templates/document_edit.html"))
		tmpl.Execute(w, struct {
			User string
			models.Document
		}{
			User:     GetBaseData(r).User,
			Document: doc,
		})
		return
	}

	// POST: Update document
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Get old content for revision
	var oldContent string
	h.DB.QueryRow("SELECT content FROM documents WHERE id = ?", id).Scan(&oldContent)

	// Save revision before updating
	h.DB.Exec("INSERT INTO revisions (document_id, content, editor_id, change_summary) VALUES (?, ?, ?, ?)",
		id, oldContent, 1, "Manual save")

	// Update document
	_, err := h.DB.Exec("UPDATE documents SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", title, content, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/documents/view?id="+id, http.StatusSeeOther)
}

// Autosave endpoint for HTMX
func (h *DocumentHandler) Autosave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	content := r.FormValue("content")

	_, err := h.DB.Exec("UPDATE documents SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", content, id)
	if err != nil {
		http.Error(w, "Save failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<span style="color: #22c55e;">Saved</span>`))
}
