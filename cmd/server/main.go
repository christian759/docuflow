package main

import (
	"database/sql"
	"docuflow/internal/db"
	"docuflow/internal/handlers"
	"html/template"
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

	// Routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Basic session check for template data
		cookie, err := r.Cookie("session_token")
		var username string
		if err == nil {
			username = cookie.Value
		}

		tmpl := template.Must(template.ParseFiles("web/templates/index.html", "web/templates/base.html"))
		data := map[string]interface{}{
			"User": username,
		}
		tmpl.Execute(w, data)
	})

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
