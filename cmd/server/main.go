package main

import (
	"database/sql"
	"docuflow/internal/db"
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

	// Routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("DocuFlow is running!"))
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
