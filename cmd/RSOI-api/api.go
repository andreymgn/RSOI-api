package main

import (
	api "github.com/andreymgn/RSOI-api/pkg/api"
	category "github.com/andreymgn/RSOI-category/pkg/category/proto"
	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/andreymgn/RSOI/pkg/tracer"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func runAPI(port int, postAddr, categoryAddr, commentAddr, postStatsAddr, userAddr, jaegerAddr string) error {
	tracer, closer, err := tracer.NewTracer("api", jaegerAddr)
	if err != nil {
		return err
	}

	defer closer.Close()

	creds, err := credentials.NewClientTLSFromFile("/post-cert.pem", "")
	if err != nil {
		return err
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

	creds, err = credentials.NewClientTLSFromFile("/category-cert.pem", "")
	if err != nil {
		return err
	}

	categoryConn, err := grpc.Dial(categoryAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return err
	}

	defer categoryConn.Close()
	catc := category.NewCategoryClient(categoryConn)

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

	server := api.NewServer(pc, catc, cc, psc, uc, tracer)
	server.Start(port)

	return nil
}
