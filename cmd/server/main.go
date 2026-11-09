package main

import (
	"log"
	"net/http"
	"os"

	"final/internal/api"
	"final/internal/store"
)

func main() {
	schema, err := os.ReadFile("internal/schema.sql")
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.Open("app.db", string(schema))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	server := api.NewServer(db)
	server.Routes(mux)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
