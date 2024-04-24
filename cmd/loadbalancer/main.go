package main

import (
	"load-balancer/internal/loadbalancer"
	"load-balancer/pkg/httpserver"
)

func main() {
	httpserver.StartServer("8080", loadbalancer.RequestHandler)
}