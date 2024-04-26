package loadbalancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL   string
	Alive bool
	mu    sync.Mutex
}

var (
	BackendServers = []*Server{
		{
			URL:   "http://localhost:8081",
			Alive: true,
		},
		{
			URL:   "http://localhost:8082",
			Alive: true,
		},
		{
			URL:   "http://localhost:8083",
			Alive: true,
		},
		{
			URL:   "http://localhost:8084",
			Alive: true,
		},
		{
			URL:   "http://localhost:8085",
			Alive: true,
		},
	}
	serversMutex sync.Mutex
)

var CurrentServer int

// select the next server
func getNextServer() *Server {
	serversMutex.Lock()
	defer serversMutex.Unlock()

	if len(BackendServers) == 0 {
		log.Println("No servers are available to handle the request.")
		return nil
	}

	start := CurrentServer
	for {
		nextServer := BackendServers[CurrentServer%len(BackendServers)]
		CurrentServer = (CurrentServer + 1) % len(BackendServers)

		if nextServer.Alive {
			return nextServer
		}

		if CurrentServer == start {
			log.Println("All servers are dead.")
			return nil
		}
	}
}

// HealthCheck checks the health of the backend servers
func healthCheck() {
	serversMutex.Lock()
	defer serversMutex.Unlock()
	for _, server := range BackendServers {
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

// starts the health check for the backend servers
func StartHealthChecks() {
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
