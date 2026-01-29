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
	fmt.Fprintf(w, "Hello from %s! v0.1.0\n", name)
}

func main() {
	http.HandleFunc("/", handler)
	port := ":8080"
	http.ListenAndServe(port, nil)
}
