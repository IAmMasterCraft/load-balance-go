package httpserver

import (
	"net/http"
	"log"
)

// StartServer starts the http server
func StartServer(port string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Failed to start server: %s", err)
    }
}