package api

import (
	"context"
	"log"
	"time"

	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workerRequest struct {
	uid      string
	doneTime time.Time
}

func (s *Server) deletePostWorker() {
	ctx := context.Background()
	for {
		req := <-s.deletePostChannel

		if time.Now().Before(req.doneTime) {
			s.deletePostChannel <- req
			continue
		}

		_, err := s.postClient.client.DeletePost(ctx,
			&post.DeletePostRequest{Uid: req.uid},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unavailable {
					newReq := workerRequest{req.uid, time.Now().Add(time.Second * 5)}
					s.deletePostChannel <- newReq
					log.Printf("DeletePost rabotyaga: retrying %s", req.uid)
				}
			}
		}
	}
}

func (s *Server) deletePostStatsWorker() {
	ctx := context.Background()
	for {
		req := <-s.deletePostStatsChannel

		if time.Now().Before(req.doneTime) {
			s.deletePostStatsChannel <- req
			continue
		}

		_, err := s.postStatsClient.client.DeletePostStats(ctx,
			&poststats.DeletePostStatsRequest{PostUid: req.uid},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unavailable {
					newReq := workerRequest{req.uid, time.Now().Add(time.Second * 5)}
					s.deletePostStatsChannel <- newReq
					log.Printf("DeletePostStats rabotyaga: retrying %s", req.uid)
				}
			}
		}
	}
}

func (s *Server) deleteCommentWorker() {
	ctx := context.Background()
	for {
		req := <-s.deleteCommentChannel

		if time.Now().Before(req.doneTime) {
			s.deleteCommentChannel <- req
			continue
		}

		_, err := s.commentClient.client.DeleteComment(ctx,
			&comment.DeleteCommentRequest{Uid: req.uid},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unavailable {
					newReq := workerRequest{req.uid, time.Now().Add(time.Second * 5)}
					s.deleteCommentChannel <- newReq
					log.Printf("DeleteComment rabotyaga: retrying %s", req.uid)
				}
			}
		}
	}
}
