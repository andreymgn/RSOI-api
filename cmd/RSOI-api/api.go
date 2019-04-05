package main

import (
	"log"

	api "github.com/andreymgn/RSOI-api/pkg/api"
	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/andreymgn/RSOI/pkg/tracer"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func runAPI(port int, postAddr, commentAddr, postStatsAddr, userAddr, jaegerAddr string) error {
	creds, err := credentials.NewClientTLSFromFile("/post-cert.pem", "")
	if err != nil {
		return err
	}

	tracer, closer, err := tracer.NewTracer("api", jaegerAddr)
	defer closer.Close()
	if err != nil {
		log.Fatal(err)
	}

	postConn, err := grpc.Dial(postAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return err
	}

	defer postConn.Close()
	pc := post.NewPostClient(postConn)

	creds, err = credentials.NewClientTLSFromFile("/comment-cert.pem", "")
	if err != nil {
		return err
	}

	commentConn, err := grpc.Dial(commentAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return err
	}

	defer commentConn.Close()
	cc := comment.NewCommentClient(commentConn)

	creds, err = credentials.NewClientTLSFromFile("/poststats-cert.pem", "")
	if err != nil {
		return err
	}

	postStatsConn, err := grpc.Dial(postStatsAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return err
	}

	defer postStatsConn.Close()
	psc := poststats.NewPostStatsClient(postStatsConn)

	creds, err = credentials.NewClientTLSFromFile("/user-cert.pem", "")
	if err != nil {
		return err
	}

	userConn, err := grpc.Dial(userAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return err
	}

	defer userConn.Close()
	uc := user.NewUserClient(userConn)

	server := api.NewServer(pc, cc, psc, uc, tracer)
	server.Start(port)

	return nil
}
