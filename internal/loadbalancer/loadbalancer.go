package loadbalancer

import (
	"net/http"
	"net/url"
	"log"
	"net/http/httputil"
	"sync"
	"time"
)

type Server struct {
	URL string
	Alive bool
	mu sync.Mutex
}

var backendServers = [] *Server{
	{
		URL: "http://localhost:8081",
		Alive: true,
	},
	{
		URL: "http://localhost:8082",
		Alive: true,
	},
	{
		URL: "http://localhost:8083",
		Alive: true,
	},
	{
		URL: "http://localhost:8084",
		Alive: true,
	},
	{
		URL: "http://localhost:8085",
		Alive: true,
	},
}

var currentServer int;

// use round robin to select the next server
func getNextServer() *Server {
	nextServer := backendServers[currentServer % len(backendServers)]
	if !nextServer.Alive {
		return getNextServer()
	}
	currentServer++
	return nextServer
}

// HealthCheck checks the health of the backend servers
func healthCheck() {
	for _, server := range backendServers {
		res, err := http.Get(server.URL)
		server.mu.Lock()
		if err != nil || res.StatusCode != 200 {
			server.Alive = false
			log.Printf("Server %s is down", server.URL)
		} else {
			server.Alive = true
		}
		server.mu.Unlock()
	}
}

// init starts the health check for the backend servers
func init() {
	go func() {
		for {
			healthCheck()
			time.Sleep(10 * time.Second)
		}
	}()
}

// RequestHandler to forward the request to the backend server
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	nextServer := getNextServer()
	if nextServer == nil {
		log.Printf("No server available")
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
    }
	remoteUrl, err := url.Parse(nextServer.URL)
	if err != nil {
		log.Printf("Error parsing backend URL: %s", err)
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
	}
	proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
	log.Printf("Forwarding request to %s", nextServer.URL)
	proxy.ServeHTTP(w, r)
}