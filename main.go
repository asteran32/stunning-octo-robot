package main

import (
	"app/route"
	"log"
)

func main() {
	log.Print("Start server")
	route.RunAPI(":8080")
}
