package httpserver

import (
	"log"
	"net/http"
)

// StartServer starts the http server
func StartServer(port string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	log.Printf("Server started on port %s", port)
}
