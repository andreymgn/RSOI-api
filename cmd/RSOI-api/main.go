package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("PORT parse error")
		return
	}

	postServerAddr := os.Getenv("POST-ADDR")
	categoryServerAddr := os.Getenv("CATEGORY-ADDR")
	commentServerAddr := os.Getenv("COMMENT-ADDR")
	postStatsServerAddr := os.Getenv("POSTSTATS-ADDR")
	userServerAddr := os.Getenv("USER-ADDR")
	jaegerAddr := os.Getenv("JAEGER-ADDR")

	log.Printf("running API service on port %d\n", port)
	err = runAPI(port, postServerAddr, categoryServerAddr, commentServerAddr, postStatsServerAddr, userServerAddr, jaegerAddr)

	if err != nil {
		log.Printf("finished with error %v", err)
	}
}
