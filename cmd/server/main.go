package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"final/internal/api"
)

func main() {
	db, err := sql.Open("sqlite3", "app.db?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schema, err := os.ReadFile("internal/schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	server := api.NewServer(db)
	server.Routes(mux)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
