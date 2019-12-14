package service

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

var (
	ctx       = context.Background()
	cookieAge = "31536000"
)

func getContextWithSession(ctx context.Context, req *http.Request) context.Context {
	sessionID, err := req.Cookie("session_id")
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, "session_id", sessionID.Value)
}

func NewHandler(s Service, staticDir string) http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/static").Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(path.Join(".", staticDir)))))

	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		location := "/signin"

		sessionID, _ := req.Cookie("session_id")
		if sessionID != nil && len(sessionID.Value) > 0 {
			location = "/timeline"
		}

		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		err := s.ServeSigninPage(ctx, w)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		instance := req.FormValue("instance")
		url, sessionId, err := s.GetAuthUrl(ctx, instance)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Set-Cookie", fmt.Sprintf("session_id=%s;max-age=%s", sessionId, cookieAge))
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodPost)

	r.HandleFunc("/oauth_callback", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		token := req.URL.Query().Get("code")
		_, err := s.GetUserToken(ctx, "", nil, token)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", "/timeline")
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/timeline", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		maxID := req.URL.Query().Get("max_id")
		sinceID := req.URL.Query().Get("since_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeTimelinePage(ctx, w, nil, maxID, sinceID, minID)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/thread/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		reply := req.URL.Query().Get("reply")
		err := s.ServeThreadPage(ctx, w, nil, id, len(reply) > 1)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/like/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.Like(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/unlike/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.UnLike(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/retweet/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.Retweet(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/unretweet/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.UnRetweet(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	r.HandleFunc("/post", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		content := req.FormValue("content")
		replyToID := req.FormValue("reply_to_id")
		id, err := s.PostTweet(ctx, w, nil, content, replyToID)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		location := "/timeline" + "#status-" + id
		if len(replyToID) > 0 {
			location = "/thread/" + replyToID + "#status-" + id
		}
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodPost)

	r.HandleFunc("/signout", func(w http.ResponseWriter, req *http.Request) {
		// TODO remove session from database
		w.Header().Add("Set-Cookie", fmt.Sprintf("session_id=;max-age=0"))
		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusSeeOther)
	}).Methods(http.MethodGet)

	return r
}
