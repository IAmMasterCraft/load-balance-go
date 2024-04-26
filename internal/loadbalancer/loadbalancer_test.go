package loadbalancer

import (
	"github.com/jarcoal/httpmock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

// TestGetNextServer tests the getNextServer function
func TestGetNextServer(t *testing.T) {

	// Reset for testing
	tempCurrentServer := CurrentServer
	CurrentServer = 0
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: false},
		{URL: "http://localhost:8082", Alive: false},
		{URL: "http://localhost:8083", Alive: true},
	}

	// First call should skip the first two and return the third
	server := getNextServer()
	if server == nil || server.URL != "http://localhost:8083" {
		t.Errorf("Expected the third server, got %v", server)
	}

	// Test wrap-around and server selection correctness
	server = getNextServer()
	if server == nil || server.URL != "http://localhost:8083" {
		t.Errorf("Expected the third server again due to wrap-around, got %v", server)
	}

	// All servers down scenario
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: false},
		{URL: "http://localhost:8082", Alive: false},
		{URL: "http://localhost:8083", Alive: false},
	}
	server = getNextServer()
	if server != nil {
		t.Errorf("Expected no server available, but got %v", server)
	}

	// reset the server
	CurrentServer = tempCurrentServer
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: true},
		{URL: "http://localhost:8082", Alive: true},
		{URL: "http://localhost:8083", Alive: true},
		{URL: "http://localhost:8084", Alive: true},
		{URL: "http://localhost:8085", Alive: true},
	}
}

func TestGetNextServerWithEmptyList(t *testing.T) {
	BackendServers = []*Server{} // ensure the list is empty

	server := getNextServer()
	if server != nil {
		t.Errorf("Expected nil, got %v", server)
	}

	// reset the server
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: true},
		{URL: "http://localhost:8082", Alive: true},
		{URL: "http://localhost:8083", Alive: true},
		{URL: "http://localhost:8084", Alive: true},
		{URL: "http://localhost:8085", Alive: true},
	}
}

// TestHealthCheck tests the healthCheck function
func TestHealthCheck(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: true, mu: sync.Mutex{}},
		{URL: "http://localhost:8082", Alive: true, mu: sync.Mutex{}},
	}

	httpmock.RegisterResponder("GET", "http://localhost:8081", httpmock.NewStringResponder(500, "Internal Server Error"))
	httpmock.RegisterResponder("GET", "http://localhost:8082", httpmock.NewStringResponder(200, "OK"))

	healthCheck()

	if BackendServers[0].Alive {
		t.Errorf("Server at %s should be marked down but is still up", BackendServers[0].URL)
	}
	if !BackendServers[1].Alive {
		t.Errorf("Server at %s should be up but is marked down", BackendServers[1].URL)
	}

	// Testing with no response or error
	httpmock.Reset()
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
	healthCheck()

	if BackendServers[1].Alive {
		t.Errorf("Server at %s should be marked down due to no response, but is still up", BackendServers[1].URL)
	}

	// reset the server
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: true},
		{URL: "http://localhost:8082", Alive: true},
		{URL: "http://localhost:8083", Alive: true},
		{URL: "http://localhost:8084", Alive: true},
		{URL: "http://localhost:8085", Alive: true},
	}
}

// TestRequestHandler tests the RequestHandler function
func TestRequestHandler(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	url, _ := url.Parse(backend.URL)
	BackendServers = []*Server{{URL: url.String(), Alive: true}}

	server := httptest.NewServer(http.HandlerFunc(RequestHandler))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	BackendServers = []*Server{}
	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %s", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503, got %d", resp.StatusCode)
	}

	// reset the server
	BackendServers = []*Server{
		{URL: "http://localhost:8081", Alive: true},
		{URL: "http://localhost:8082", Alive: true},
		{URL: "http://localhost:8083", Alive: true},
		{URL: "http://localhost:8084", Alive: true},
		{URL: "http://localhost:8085", Alive: true},
	}
}
