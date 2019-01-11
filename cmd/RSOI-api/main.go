package main

import (
	"flag"
	"fmt"
)

func main() {
	portNum := flag.Int("port", -1, "Port on which service well listen")
	postServerAddr := flag.String("post-server", "", "Address of post server")
	commentServerAddr := flag.String("comment-server", "", "Address of comment server")
	postStatsServerAddr := flag.String("post-stats-server", "", "Address of post stats server")
	userServerAddr := flag.String("user-server", "", "Address of user server")
	jaegerAddr := flag.String("jaeger-addr", "", "Jaeger address")

	flag.Parse()

	port := *portNum
	ps := *postServerAddr
	cs := *commentServerAddr
	pss := *postStatsServerAddr
	us := *userServerAddr
	ja := *jaegerAddr

	fmt.Printf("running API service on port %d\n", port)
	err := runAPI(port, ps, cs, pss, us, ja)

	if err != nil {
		fmt.Printf("finished with error %v", err)
	}
}
