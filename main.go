package main

import (
	"database/sql"
	"docuflow/db"
	"docuflow/handlers"
	"log"
	"net/http"
)

type Config struct {
	Port string
	DB   *sql.DB
}

func main() {
	// Initialize Database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Router
	mux := http.NewServeMux()

	// Static Files
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	authHandler := &handlers.AuthHandler{DB: database}
	docHandler := &handlers.DocumentHandler{DB: database}
	revHandler := &handlers.RevisionHandler{DB: database}
	commentHandler := &handlers.CommentHandler{DB: database}
	searchHandler := &handlers.SearchHandler{DB: database}

	// Routes
	mux.HandleFunc("/", docHandler.ListDocuments)
	mux.HandleFunc("/documents/new", docHandler.NewDocument)
	mux.HandleFunc("/documents/view", docHandler.ViewDocument)
	mux.HandleFunc("/documents/edit", docHandler.EditDocument)
	mux.HandleFunc("/documents/autosave", docHandler.Autosave)

	mux.HandleFunc("/revisions", revHandler.ListRevisions)
	mux.HandleFunc("/revisions/view", revHandler.ViewRevision)
	mux.HandleFunc("/revisions/rollback", revHandler.Rollback)

	mux.HandleFunc("/comments", commentHandler.ListComments)
	mux.HandleFunc("/comments/add", commentHandler.AddComment)
	mux.HandleFunc("/comments/delete", commentHandler.DeleteComment)

	mux.HandleFunc("/search", searchHandler.Search)

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
