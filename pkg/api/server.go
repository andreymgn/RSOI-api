package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	category "github.com/andreymgn/RSOI-category/pkg/category/proto"
	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/andreymgn/RSOI/pkg/tracer"
	"github.com/rs/cors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MaxQueueLength = 100
)

type PostClient struct {
	client post.PostClient
}

type CategoryClient struct {
	client category.CategoryClient
}

type CommentClient struct {
	client comment.CommentClient
}

type PostStatsClient struct {
	client poststats.PostStatsClient
}

type UserClient struct {
	client user.UserClient
}

type Server struct {
	router                 *tracer.TracedRouter
	postClient             *PostClient
	categoryClient         *CategoryClient
	commentClient          *CommentClient
	postStatsClient        *PostStatsClient
	userClient             *UserClient
	deletePostChannel      chan workerRequest
	deletePostStatsChannel chan workerRequest
	deleteCommentChannel   chan workerRequest
}

// NewServer returns new instance of Server
func NewServer(pc post.PostClient, catc category.CategoryClient, cc comment.CommentClient, psc poststats.PostStatsClient, uc user.UserClient, tr opentracing.Tracer) *Server {
	return &Server{
		tracer.NewRouter(tr),
		&PostClient{pc},
		&CategoryClient{catc},
		&CommentClient{cc},
		&PostStatsClient{psc},
		&UserClient{uc},
		make(chan workerRequest, MaxQueueLength),
		make(chan workerRequest, MaxQueueLength),
		make(chan workerRequest, MaxQueueLength),
	}
}

func getAuthorizationToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	splitToken := strings.Split(auth, "Bearer ")
	if len(splitToken) == 2 {
		return splitToken[1]
	}

	return ""
}

func handleRPCError(w http.ResponseWriter, err error) {
	log.Println("API err:", err)
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
		return
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusUnprocessableEntity)
		return
	case codes.Unauthenticated:
		w.WriteHeader(http.StatusForbidden)
		return
	case codes.Unavailable:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Start starts HTTP server which can shut down gracefully
func (s *Server) Start(port int) {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Access-Control-Allow-Origin", "Authorization"},
		AllowCredentials: true,
	})
	s.router.Mux.Use(setContentType)
	s.routes()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      c.Handler(s.router),
	}

	go s.deletePostWorker()
	go s.deletePostStatsWorker()
	go s.deleteCommentWorker()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
