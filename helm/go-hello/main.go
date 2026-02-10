package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "go-hello"
	}
	dbname := os.Getenv("DB_NAME")
	dbpass := os.Getenv("DB_PASS")
	fmt.Fprintf(w, "X Hello from %s! v0.1.2\n DB_NAME: %s, DB_PASS: %s", name, dbname, dbpass)
}

func main() {
	http.HandleFunc("/", handler)
	port := ":8080"
	http.ListenAndServe(port, nil)
}
