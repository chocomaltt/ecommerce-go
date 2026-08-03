package main

import (
	"log"
	"net/http"
	"os"
)

const serviceName = "payment-service"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	log.Printf("%s listening on :%s", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
