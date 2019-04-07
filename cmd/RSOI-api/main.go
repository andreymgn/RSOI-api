package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("PORT parse error")
		return
	}

	postServerAddr := os.Getenv("POST-ADDR")
	categoryServerAddr := os.Getenv("CATEGORY-ADDR")
	commentServerAddr := os.Getenv("COMMENT-ADDR")
	postStatsServerAddr := os.Getenv("POSTSTATS-ADDR")
	userServerAddr := os.Getenv("USER-ADDR")
	jaegerAddr := os.Getenv("JAEGER-ADDR")

	fmt.Printf("running API service on port %d\n", port)
	err = runAPI(port, postServerAddr, categoryServerAddr, commentServerAddr, postStatsServerAddr, userServerAddr, jaegerAddr)

	if err != nil {
		fmt.Printf("finished with error %v", err)
	}
}
