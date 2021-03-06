package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	category "github.com/andreymgn/RSOI-category/pkg/category/proto"
	comment "github.com/andreymgn/RSOI-comment/pkg/comment/proto"
	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
	poststats "github.com/andreymgn/RSOI-poststats/pkg/poststats/proto"
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) getPosts() http.HandlerFunc {
	type p struct {
		UID         string
		UserUID     string
		CategoryUID string
		Title       string
		URL         string
		CreatedAt   time.Time
		ModifiedAt  time.Time
		NumLikes    int32
		NumDislikes int32
		NumViews    int32
	}

	type response struct {
		Posts      []p
		PageSize   int32
		PageNumber int32
	}

	return func(w http.ResponseWriter, r *http.Request) {
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		var pageNum, sizeNum int32 = 0, 10
		if page != "" {
			n, err := strconv.Atoi(page)
			if err != nil {
				http.Error(w, "can't parse query parameter `page`", http.StatusBadRequest)
				return
			}
			pageNum = int32(n)
		}

		if size != "" {
			n, err := strconv.Atoi(size)
			if err != nil {
				http.Error(w, "can't parse query parameter `size`", http.StatusBadRequest)
				return
			}
			sizeNum = int32(n)
		}

		ctx := r.Context()
		postResponse, err := s.postClient.client.ListPosts(ctx,
			&post.ListPostsRequest{PageSize: sizeNum, PageNumber: pageNum},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		posts := make([]p, len(postResponse.Posts))
		for i, singlePostResponse := range postResponse.Posts {
			posts[i].UID = singlePostResponse.Uid
			posts[i].UserUID = singlePostResponse.UserUid
			posts[i].CategoryUID = singlePostResponse.CategoryUid
			posts[i].Title = singlePostResponse.Title
			posts[i].URL = singlePostResponse.Url
			posts[i].CreatedAt, err = ptypes.Timestamp(singlePostResponse.CreatedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			posts[i].ModifiedAt, err = ptypes.Timestamp(singlePostResponse.ModifiedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			postStats, err := s.postStatsClient.client.GetPostStats(ctx,
				&poststats.GetPostStatsRequest{PostUid: posts[i].UID},
			)
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
				posts[i].NumLikes = -1
				posts[i].NumDislikes = -1
				posts[i].NumViews = -1

			} else if err != nil {
				handleRPCError(w, err)
				return
			} else {
				posts[i].NumLikes = postStats.NumLikes
				posts[i].NumDislikes = postStats.NumDislikes
				posts[i].NumViews = postStats.NumViews
			}
		}

		resp := response{posts, sizeNum, pageNum}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getPostsByCategory() http.HandlerFunc {
	type p struct {
		UID         string
		UserUID     string
		CategoryUID string
		Title       string
		URL         string
		CreatedAt   time.Time
		ModifiedAt  time.Time
		NumLikes    int32
		NumDislikes int32
		NumViews    int32
	}

	type response struct {
		Posts      []p
		PageSize   int32
		PageNumber int32
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		var pageNum, sizeNum int32 = 0, 10
		if page != "" {
			n, err := strconv.Atoi(page)
			if err != nil {
				http.Error(w, "can't parse query parameter `page`", http.StatusBadRequest)
				return
			}
			pageNum = int32(n)
		}

		if size != "" {
			n, err := strconv.Atoi(size)
			if err != nil {
				http.Error(w, "can't parse query parameter `size`", http.StatusBadRequest)
				return
			}
			sizeNum = int32(n)
		}

		ctx := r.Context()
		postResponse, err := s.postClient.client.ListPostsByCategory(ctx,
			&post.ListPostsByCategoryRequest{PageSize: sizeNum, PageNumber: pageNum, CategoryUid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		posts := make([]p, len(postResponse.Posts))
		for i, singlePostResponse := range postResponse.Posts {
			posts[i].UID = singlePostResponse.Uid
			posts[i].UserUID = singlePostResponse.UserUid
			posts[i].CategoryUID = singlePostResponse.CategoryUid
			posts[i].Title = singlePostResponse.Title
			posts[i].URL = singlePostResponse.Url
			posts[i].CreatedAt, err = ptypes.Timestamp(singlePostResponse.CreatedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			posts[i].ModifiedAt, err = ptypes.Timestamp(singlePostResponse.ModifiedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			postStats, err := s.postStatsClient.client.GetPostStats(ctx,
				&poststats.GetPostStatsRequest{PostUid: posts[i].UID},
			)
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
				posts[i].NumLikes = -1
				posts[i].NumDislikes = -1
				posts[i].NumViews = -1

			} else if err != nil {
				handleRPCError(w, err)
				return
			} else {
				posts[i].NumLikes = postStats.NumLikes
				posts[i].NumDislikes = postStats.NumDislikes
				posts[i].NumViews = postStats.NumViews
			}
		}

		resp := response{posts, sizeNum, pageNum}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) createPost() http.HandlerFunc {
	type request struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	type response struct {
		UID         string
		UserUID     string
		Title       string
		URL         string
		CreatedAt   time.Time
		ModifiedAt  time.Time
		NumLikes    int32
		NumDislikes int32
		NumViews    int32
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var req request
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		ctx := r.Context()
		p, err := s.postClient.client.CreatePost(ctx,
			&post.CreatePostRequest{Title: req.Title, Url: req.URL, UserUid: userUID, CategoryUid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		_, err = s.postStatsClient.client.CreatePostStats(ctx,
			&poststats.CreatePostStatsRequest{PostUid: p.Uid},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unavailable {
					_, err := s.postClient.client.DeletePost(ctx,
						&post.DeletePostRequest{Uid: p.Uid},
					)
					if err != nil {
						handleRPCError(w, err)
						return
					}
					w.WriteHeader(http.StatusServiceUnavailable)
				} else {
					handleRPCError(w, err)
					return
				}

			} else {
				handleRPCError(w, err)
				return
			}
		}

		createdAt, err := ptypes.Timestamp(p.CreatedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		modifiedAt, err := ptypes.Timestamp(p.ModifiedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		response := response{p.Uid, p.UserUid, p.Title, p.Url, createdAt, modifiedAt, 0, 0, 0}
		json, err := json.Marshal(response)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}

func (s *Server) getPost() http.HandlerFunc {
	type response struct {
		UID         string
		UserUID     string
		CategoryUID string
		Title       string
		URL         string
		CreatedAt   time.Time
		ModifiedAt  time.Time
		NumLikes    int32
		NumDislikes int32
		NumViews    int32
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		postResponse, err := s.postClient.client.GetPost(ctx,
			&post.GetPostRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		var res response
		res.UID = postResponse.Uid
		res.UserUID = postResponse.UserUid
		res.CategoryUID = postResponse.CategoryUid
		res.Title = postResponse.Title
		res.URL = postResponse.Url
		res.CreatedAt, err = ptypes.Timestamp(postResponse.CreatedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		res.ModifiedAt, err = ptypes.Timestamp(postResponse.ModifiedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		postStats, err := s.postStatsClient.client.GetPostStats(ctx,
			&poststats.GetPostStatsRequest{PostUid: res.UID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		res.NumLikes = postStats.NumLikes
		res.NumDislikes = postStats.NumDislikes
		res.NumViews = postStats.NumViews

		json, err := json.Marshal(res)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		_, err = s.postStatsClient.client.IncreaseViews(ctx,
			&poststats.IncreaseViewsRequest{PostUid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) updatePost() http.HandlerFunc {
	type request struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var req request
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		owner, err := s.postClient.client.GetPostOwner(ctx,
			&post.GetPostOwnerRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID != owner.OwnerUid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, err = s.postClient.client.UpdatePost(ctx,
			&post.UpdatePostRequest{Uid: uid, Title: req.Title, Url: req.URL},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) deletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		uid := vars["uid"]
		categoryUID := vars["categoryuid"]

		ctx := r.Context()
		owner, err := s.postClient.client.GetPostOwner(ctx,
			&post.GetPostOwnerRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID != owner.OwnerUid {
			// Check if current user is global admin
			userInfo, err := s.userClient.client.GetUserInfo(ctx,
				&user.GetUserInfoRequest{Uid: userUID},
			)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			if !userInfo.IsAdmin {
				// Check if current user is category admin
				categoryInfo, err := s.categoryClient.client.GetCategoryInfo(ctx,
					&category.GetCategoryInfoRequest{Uid: categoryUID},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}

				if userUID != categoryInfo.UserUid {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}

		s.deletePostChannel <- workerRequest{uid, time.Now()}

		s.deletePostStatsChannel <- workerRequest{uid, time.Now()}

		comments, err := s.commentClient.client.ListComments(ctx,
			&comment.ListCommentsRequest{PostUid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		for _, c := range comments.Comments {
			s.deleteCommentChannel <- workerRequest{c.Uid, time.Now()}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) likePost() http.HandlerFunc {
	type response struct {
		Success   bool
		FirstTime bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		changed, err := s.postStatsClient.client.LikePost(ctx,
			&poststats.LikePostRequest{PostUid: uid, UserUid: userUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		res := response{changed.Success, changed.FirstTime}

		json, err := json.Marshal(res)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) dislikePost() http.HandlerFunc {
	type response struct {
		Success   bool
		FirstTime bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		changed, err := s.postStatsClient.client.DislikePost(ctx,
			&poststats.DislikePostRequest{PostUid: uid, UserUid: userUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		res := response{changed.Success, changed.FirstTime}

		json, err := json.Marshal(res)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) reportPost() http.HandlerFunc {
	type request struct {
		Reason string
	}

	type response struct {
		UID         string
		CategoryUID string
		PostUID     string
		CommentUID  string
		Reason      string
		CreatedAt   time.Time
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthorizationToken(r)
		if userToken == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUID, err := s.getUIDByToken(userToken)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var req request
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		vars := mux.Vars(r)
		categoryUID := vars["categoryuid"]
		postUID := vars["uid"]

		ctx := r.Context()
		report, err := s.categoryClient.client.CreateReport(ctx,
			&category.CreateReportRequest{CategoryUid: categoryUID, PostUid: postUID, Reason: req.Reason, CommentUid: uuid.Nil.String()},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		createdAt, err := ptypes.Timestamp(report.CreatedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		response := response{report.Uid, report.CategoryUid, report.PostUid, report.CommentUid, report.Reason, createdAt}
		json, err := json.Marshal(response)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}
