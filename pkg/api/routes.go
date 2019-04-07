package api

import "net/http"

func (s *Server) routes() {
	categoryRouter := s.router.Mux.PathPrefix("/api/categories").Subrouter()
	categoryRouter.HandleFunc("/", s.getCategories()).Methods("GET")
	categoryRouter.HandleFunc("/", s.createCategory()).Methods("POST")
	categoryRouter.HandleFunc("/{uid}/posts", s.getPosts()).Methods("GET")
	categoryRouter.HandleFunc("/{uid}/posts", s.createPost()).Methods("POST")

	categoryRouter.HandleFunc("/{categoryuid}/posts/{uid}", s.getPost()).Methods("GET")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{uid}", s.updatePost()).Methods("PATCH")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{uid}", s.deletePost()).Methods("DELETE")

	categoryRouter.HandleFunc("/{categoryuid}/posts/{uid}/like", s.likePost()).Methods("PATCH")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{uid}/dislike", s.dislikePost()).Methods("PATCH")

	categoryRouter.HandleFunc("/{categoryuid}/posts/{postuid}/comments/", s.getPostComments()).Methods("GET")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{postuid}/comments/", s.createComment()).Methods("POST")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{postuid}/comments/{uid}", s.getPostComments()).Methods("GET")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{postuid}/comments/{uid}", s.updateComment()).Methods("PATCH")
	categoryRouter.HandleFunc("/{categoryuid}/posts/{postuid}/comments/{uid}", s.deleteComment()).Methods("DELETE")

	s.router.Mux.HandleFunc("/api/user", s.createUser()).Methods("POST")
	s.router.Mux.HandleFunc("/api/user/{uid}", s.getUserInfo()).Methods("GET")
	s.router.Mux.HandleFunc("/api/auth/token", s.getToken()).Methods("POST")
	s.router.Mux.HandleFunc("/api/auth/refresh", s.refreshToken()).Methods("POST")

	s.router.Mux.HandleFunc("/api/oauth/app", s.createApp()).Methods("POST")
	s.router.Mux.HandleFunc("/api/oauth/app/{uid}", s.getAppInfo()).Methods("GET")
	s.router.Mux.HandleFunc("/api/oauth/authorize", s.getOAuthCode()).Methods("POST")
	s.router.Mux.HandleFunc("/api/oauth/token", s.getTokenFromOAuthCode()).Methods("GET")

	s.router.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hello, world!")) })
}
