package main

import (
	"app/route"
	"log"
)

func main() {
	log.Print("Starting Go gin server Port :8080")
	// Run
	route.RunAPI(":8080")
}
