package main

import (
	"app/redis"
	"app/route"
	"log"
)

func main() {
	log.Print("Start server")
	// DB init
	// Redis init
	redis.Init()
	// Run
	route.RunAPI(":8080")
}
