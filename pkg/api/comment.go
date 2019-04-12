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
	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
)

func (s *Server) getPostComments() http.HandlerFunc {
	type c struct {
		UID        string
		UserUID    string
		PostUID    string
		Body       string
		ParentUID  string
		CreatedAt  time.Time
		ModifiedAt time.Time
	}

	type response struct {
		Comments   []c
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

		vars := mux.Vars(r)
		uid := vars["uid"]
		postUID := vars["postuid"]

		ctx := r.Context()
		checkExistsResponse, err := s.postClient.client.CheckPostExists(ctx,
			&post.CheckPostExistsRequest{Uid: postUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !checkExistsResponse.Exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		commentsResponse, err := s.commentClient.client.ListComments(ctx,
			&comment.ListCommentsRequest{PostUid: postUID, CommentUid: uid, PageSize: sizeNum, PageNumber: pageNum},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		comments := make([]c, len(commentsResponse.Comments))
		for i, singleComment := range commentsResponse.Comments {
			comments[i].UID = singleComment.Uid
			comments[i].PostUID = singleComment.PostUid
			comments[i].Body = "[deleted]"
			comments[i].ParentUID = singleComment.ParentUid
			comments[i].CreatedAt, err = ptypes.Timestamp(singleComment.CreatedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			comments[i].ModifiedAt, err = ptypes.Timestamp(singleComment.ModifiedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}

			if !singleComment.IsDeleted {
				comments[i].UserUID = singleComment.UserUid
				comments[i].Body = singleComment.Body

			}
		}

		resp := response{comments, sizeNum, pageNum}

		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getSingleComment() http.HandlerFunc {
	type response struct {
		UID        string
		UserUID    string
		PostUID    string
		Body       string
		ParentUID  string
		CreatedAt  time.Time
		ModifiedAt time.Time
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]
		postUID := vars["postuid"]

		ctx := r.Context()
		checkExistsResponse, err := s.postClient.client.CheckPostExists(ctx,
			&post.CheckPostExistsRequest{Uid: postUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !checkExistsResponse.Exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		singleComment, err := s.commentClient.client.GetComment(ctx,
			&comment.GetCommentRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		var res response
		res.UID = singleComment.Uid
		res.PostUID = singleComment.PostUid
		res.Body = "[deleted]"
		res.ParentUID = singleComment.ParentUid
		res.CreatedAt, err = ptypes.Timestamp(singleComment.CreatedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		res.ModifiedAt, err = ptypes.Timestamp(singleComment.ModifiedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !singleComment.IsDeleted {
			res.UserUID = singleComment.UserUid
			res.Body = singleComment.Body

		}

		json, err := json.Marshal(res)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) createComment() http.HandlerFunc {
	type request struct {
		Body      string `json:"body"`
		ParentUID string `json:"parent_uid"`
	}

	type response struct {
		UID        string
		UserUID    string
		PostUID    string
		Body       string
		ParentUID  string
		CreatedAt  time.Time
		ModifiedAt time.Time
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
		postUID := vars["postuid"]

		ctx := r.Context()
		checkExistsResponse, err := s.postClient.client.CheckPostExists(ctx,
			&post.CheckPostExistsRequest{Uid: postUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !checkExistsResponse.Exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		c, err := s.commentClient.client.CreateComment(ctx,
			&comment.CreateCommentRequest{PostUid: postUID, Body: req.Body, ParentUid: req.ParentUID, UserUid: userUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		createdAt, err := ptypes.Timestamp(c.CreatedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		modifiedAt, err := ptypes.Timestamp(c.ModifiedAt)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		response := response{c.Uid, c.UserUid, c.PostUid, c.Body, c.ParentUid, createdAt, modifiedAt}
		json, err := json.Marshal(response)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}

func (s *Server) updateComment() http.HandlerFunc {
	type request struct {
		Body string `json:"body"`
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
		postUID := vars["postuid"]

		ctx := r.Context()
		checkExistsResponse, err := s.postClient.client.CheckPostExists(ctx,
			&post.CheckPostExistsRequest{Uid: postUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !checkExistsResponse.Exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		owner, err := s.commentClient.client.GetOwner(ctx,
			&comment.GetOwnerRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if userUID != owner.OwnerUid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, err = s.commentClient.client.UpdateComment(ctx,
			&comment.UpdateCommentRequest{Uid: uid, Body: req.Body},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) deleteComment() http.HandlerFunc {
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
		postUID := vars["postuid"]
		categoryUID := vars["categoryuid"]

		ctx := r.Context()
		checkExistsResponse, err := s.postClient.client.CheckPostExists(ctx,
			&post.CheckPostExistsRequest{Uid: postUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		if !checkExistsResponse.Exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		owner, err := s.commentClient.client.GetOwner(ctx,
			&comment.GetOwnerRequest{Uid: uid},
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

		_, err = s.commentClient.client.RemoveContent(ctx,
			&comment.RemoveContentRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) reportComment() http.HandlerFunc {
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
		postUID := vars["postuid"]
		commentUID := vars["uid"]

		ctx := r.Context()
		report, err := s.categoryClient.client.CreateReport(ctx,
			&category.CreateReportRequest{CategoryUid: categoryUID, PostUid: postUID, CommentUid: commentUID, Reason: req.Reason},
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
