package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	category "github.com/andreymgn/RSOI-category/pkg/category/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
)

func (s *Server) getCategories() http.HandlerFunc {
	type c struct {
		UID         string
		UserUID     string
		Name        string
		Description string
	}

	type response struct {
		Categories []c
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
		categoryResponse, err := s.categoryClient.client.ListCategories(ctx,
			&category.ListCategoriesRequest{PageSize: sizeNum, PageNumber: pageNum},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		categories := make([]c, len(categoryResponse.Categories))
		for i, singleCategory := range categoryResponse.Categories {
			categories[i].UID = singleCategory.Uid
			categories[i].UserUID = singleCategory.UserUid
			categories[i].Name = singleCategory.Name
			categories[i].Description = singleCategory.Description
		}

		resp := response{categories, sizeNum, pageNum}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getCategoryInfo() http.HandlerFunc {
	type response struct {
		UID         string
		UserUID     string
		Name        string
		Description string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		c, err := s.categoryClient.client.GetCategoryInfo(ctx,
			&category.GetCategoryInfoRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		resp := response{uid, c.UserUid, c.Name, c.Description}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}
func (s *Server) createCategory() http.HandlerFunc {
	type request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type response struct {
		UID         string
		UserUID     string
		Name        string
		Description string
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

		ctx := r.Context()
		c, err := s.categoryClient.client.CreateCategory(ctx,
			&category.CreateCategoryRequest{Name: req.Name, Description: req.Description, UserUid: userUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		response := response{c.Uid, c.UserUid, c.Name, c.Description}
		json, err := json.Marshal(response)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}

func (s *Server) getReports() http.HandlerFunc {
	type report struct {
		UID         string
		CategoryUID string
		PostUID     string
		CommentUID  string
		Reason      string
		CreatedAt   time.Time
	}

	type response struct {
		Reports    []report
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
		categoryUID := vars["uid"]

		ctx := r.Context()
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

		reportsResponse, err := s.categoryClient.client.ListReports(ctx,
			&category.ListReportsRequest{CategoryUid: categoryUID, PageSize: sizeNum, PageNumber: pageNum},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		reports := make([]report, len(reportsResponse.Reports))
		for i, singleReport := range reportsResponse.Reports {
			reports[i].UID = singleReport.Uid
			reports[i].CategoryUID = singleReport.CategoryUid
			reports[i].PostUID = singleReport.PostUid
			reports[i].CommentUID = singleReport.CommentUid
			reports[i].Reason = singleReport.Reason
			reports[i].CreatedAt, err = ptypes.Timestamp(singleReport.CreatedAt)
			if err != nil {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{reports, sizeNum, pageNum}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) deleteReport() http.HandlerFunc {
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

		_, err = s.categoryClient.client.DeleteReport(ctx,
			&category.DeleteReportRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
