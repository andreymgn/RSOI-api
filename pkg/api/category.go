package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	post "github.com/andreymgn/RSOI-post/pkg/post/proto"
)

func (s *Server) getCategories() http.HandlerFunc {
	type c struct {
		UID     string
		UserUID string
		Name    string
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
		categoryResponse, err := s.postClient.client.ListCategories(ctx,
			&post.ListCategoriesRequest{PageSize: sizeNum, PageNumber: pageNum},
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

func (s *Server) createCategory() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
	}

	type response struct {
		UID     string
		UserUID string
		Name    string
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
		c, err := s.postClient.client.CreateCategory(ctx,
			&post.CreateCategoryRequest{Name: req.Name, UserUid: userUID},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		response := response{c.Uid, c.UserUid, c.Name}
		json, err := json.Marshal(response)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}
