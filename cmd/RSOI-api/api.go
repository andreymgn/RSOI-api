package main

import (
	"log"

	api "github.com/andreymgn/RSOI-api/pkg/api"
	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"google.golang.org/grpc"
)

func runAPI(port int, postAddr, commentAddr, postStatsAddr, userAddr string) error {
	postConn, err := grpc.Dial(postAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer postConn.Close()
	pc := post.NewPostClient(postConn)

	commentConn, err := grpc.Dial(commentAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer commentConn.Close()
	cc := comment.NewCommentClient(commentConn)

	postStatsConn, err := grpc.Dial(postStatsAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer postStatsConn.Close()
	psc := poststats.NewPostStatsClient(postStatsConn)

	userConn, err := grpc.Dial(userAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer userConn.Close()
	uc := user.NewUserClient(userConn)

	server := api.NewServer(pc, cc, psc, uc)
	server.Start(port)

	return nil
}
