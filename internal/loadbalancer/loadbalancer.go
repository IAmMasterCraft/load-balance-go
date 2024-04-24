package loadbalancer

import (
	"net/http"
	"net/url"
	"log"
	"net/http/httputil"
)

var backendServers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
	"http://localhost:8084",
}

var currentServer int;

// use round robin to select the next server
func getNextServer() string {
	nextServer := backendServers[currentServer % len(backendServers)]
	currentServer++
	return nextServer
}

// RequestHandler to forward the request to the backend server
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	nextServer := getNextServer()
	remoteUrl, err := url.Parse(nextServer)
	if err != nil {
		log.Printf("Error parsing backend URL: %s", err)
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
	}
	proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
	proxy.ServeHTTP(w, r)
}