package main

import (
	"log"
	"net/http"
	"os"

	"arrow/internal/api"
	"arrow/internal/mongo"
	"arrow/internal/oref"
)

func main() {
	log.Println("Starting server...")

	mongo.Init()
	go oref.StartPolling()

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://127.0.0.1:8080"
	}

	corsMiddleware := api.CORS(corsOrigin)

	mux := http.NewServeMux()

	mux.Handle("/api/alerts",
		corsMiddleware(http.HandlerFunc(api.HandleAlerts)))

	mux.Handle("/api/history",
		corsMiddleware(http.HandlerFunc(api.HandleHistory)))

	fs := http.FileServer(http.Dir("ui/dist"))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, "ui/dist/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	}))

	addr := os.Getenv("ADDR_INFO")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	log.Printf("Serving at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}