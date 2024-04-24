package main

import (
	"load-balancer/internal/loadbalancer"
	"load-balancer/pkg/httpserver"
)

func main() {
	println("Starting load balancer")
	httpserver.StartServer("9000", loadbalancer.RequestHandler)
}