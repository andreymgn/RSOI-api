package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	user "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) getUserInfo() http.HandlerFunc {
	type response struct {
		ID       string
		Username string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		getUserResponse, err := s.userClient.client.GetUserInfo(ctx,
			&user.GetUserInfoRequest{Uid: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		resp := response{getUserResponse.Uid, getUserResponse.Username}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) createUser() http.HandlerFunc {
	type request struct {
		Username string
		Password string
	}

	type response struct {
		ID       string
		Username string
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		createUserResponse, err := s.userClient.client.CreateUser(ctx,
			&user.CreateUserRequest{Token: s.userClient.token, Username: req.Username, Password: req.Password},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				createUserResponse, err = s.userClient.client.CreateUser(ctx,
					&user.CreateUserRequest{Token: s.userClient.token, Username: req.Username, Password: req.Password},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{createUserResponse.Uid, createUserResponse.Username}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}

func (s *Server) getToken() http.HandlerFunc {
	type request struct {
		Username string
		Password string
		Refresh  bool `json:",omitempty"`
	}

	type response struct {
		UID          string
		AccessToken  string
		RefreshToken string `json:",omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		accessTokenResponse, err := s.userClient.client.GetAccessToken(ctx,
			&user.GetTokenRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				accessTokenResponse, err = s.userClient.client.GetAccessToken(ctx,
					&user.GetTokenRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{}
		resp.AccessToken = accessTokenResponse.Token
		resp.UID = accessTokenResponse.Uid

		if req.Refresh {
			refreshTokenResponse, err := s.userClient.client.GetRefreshToken(ctx,
				&user.GetTokenRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password},
			)
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
					err := s.updateUserToken()
					if err != nil {
						handleRPCError(w, err)
						return
					}
					refreshTokenResponse, err = s.userClient.client.GetRefreshToken(ctx,
						&user.GetTokenRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password},
					)
					if err != nil {
						handleRPCError(w, err)
						return
					}
				} else {
					handleRPCError(w, err)
					return
				}
			}

			resp.RefreshToken = refreshTokenResponse.Token
		}

		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) refreshToken() http.HandlerFunc {
	type request struct {
		Token string
	}

	type response struct {
		AccessToken  string
		RefreshToken string
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		refreshTokenResponse, err := s.userClient.client.RefreshAccessToken(ctx,
			&user.RefreshAccessTokenRequest{ApiToken: s.userClient.token, RefreshToken: req.Token},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				refreshTokenResponse, err = s.userClient.client.RefreshAccessToken(ctx,
					&user.RefreshAccessTokenRequest{ApiToken: s.userClient.token, RefreshToken: req.Token},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{refreshTokenResponse.AccessToken, refreshTokenResponse.RefreshToken}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getUIDByToken(token string) (string, error) {
	ctx := context.Background()
	uid, err := s.userClient.client.GetUserByAccessToken(ctx,
		&user.GetUserByAccessTokenRequest{ApiToken: s.userClient.token, UserToken: token},
	)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			err := s.updateUserToken()
			if err != nil {
				return "", err
			}
			uid, err = s.userClient.client.GetUserByAccessToken(ctx,
				&user.GetUserByAccessTokenRequest{ApiToken: s.userClient.token, UserToken: token},
			)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return uid.Uid, nil
}

func (s *Server) createApp() http.HandlerFunc {
	type request struct {
		Name string
	}

	type response struct {
		ID     string
		Secret string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userToken := getAuthrizationToken(r)
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
		createAppResponse, err := s.userClient.client.CreateApp(ctx,
			&user.CreateAppRequest{ApiToken: s.userClient.token, Name: req.Name, Owner: userUID},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				createAppResponse, err = s.userClient.client.CreateApp(ctx,
					&user.CreateAppRequest{ApiToken: s.userClient.token, Name: req.Name, Owner: userUID},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{createAppResponse.Id, createAppResponse.Secret}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(json)
	}
}

func (s *Server) getAppInfo() http.HandlerFunc {
	type response struct {
		Owner string
		Name  string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]

		ctx := r.Context()
		getAppInfoResponse, err := s.userClient.client.GetAppInfo(ctx,
			&user.GetAppInfoRequest{Id: uid},
		)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		resp := response{getAppInfoResponse.Owner, getAppInfoResponse.Name}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getOAuthCode() http.HandlerFunc {
	type request struct {
		AppUID   string
		Username string
		Password string
	}

	type response struct {
		Code string
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		oauthCodeResponse, err := s.userClient.client.GetOAuthCode(ctx,
			&user.GetOAuthCodeRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password, AppUid: req.AppUID},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				oauthCodeResponse, err = s.userClient.client.GetOAuthCode(ctx,
					&user.GetOAuthCodeRequest{ApiToken: s.userClient.token, Username: req.Username, Password: req.Password, AppUid: req.AppUID},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{oauthCodeResponse.Code}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) getTokenFromOAuthCode() http.HandlerFunc {
	type response struct {
		AccessToken  string
		RefreshToken string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		if queryParams.Get("grant_type") != "authorization_code" {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		code := queryParams.Get("code")
		appID := queryParams.Get("client_id")
		appSecret := queryParams.Get("client_secret")
		ctx := r.Context()
		getTokenResponse, err := s.userClient.client.GetTokenFromCode(ctx,
			&user.GetTokenFromCodeRequest{ApiToken: s.userClient.token, Code: code, AppUid: appID, AppSecret: appSecret},
		)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
				err := s.updateUserToken()
				if err != nil {
					handleRPCError(w, err)
					return
				}
				getTokenResponse, err = s.userClient.client.GetTokenFromCode(ctx,
					&user.GetTokenFromCodeRequest{ApiToken: s.userClient.token, Code: code, AppUid: appID, AppSecret: appSecret},
				)
				if err != nil {
					handleRPCError(w, err)
					return
				}
			} else {
				handleRPCError(w, err)
				return
			}
		}

		resp := response{getTokenResponse.AccessToken, getTokenResponse.RefreshToken}
		json, err := json.Marshal(resp)
		if err != nil {
			handleRPCError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) updateUserToken() error {
	token, err := s.userClient.client.GetServiceToken(context.Background(),
		&user.GetServiceTokenRequest{AppId: s.userClient.appID, AppSecret: s.userClient.appSecret},
	)
	if err != nil {
		return err
	}

	s.userClient.token = token.Token
	return nil
}
