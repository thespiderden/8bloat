package service

import (
	"context"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"time"

	"bloat/model"

	"github.com/gorilla/mux"
)

var (
	ctx       = context.Background()
	cookieAge = "31536000"
)

func NewHandler(s Service, staticDir string) http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/static").Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(path.Join(".", staticDir)))))

	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		location := "/signin"

		sessionID, _ := req.Cookie("session_id")
		if sessionID != nil && len(sessionID.Value) > 0 {
			location = "/timeline/home"
		}

		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusFound)
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
		url, sessionID, err := s.GetAuthUrl(ctx, instance)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   sessionID,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})

		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/oauth_callback", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		token := req.URL.Query().Get("code")
		_, err := s.GetUserToken(ctx, "", nil, token)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", "/timeline/home")
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodGet)

	r.HandleFunc("/timeline", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Location", "/timeline/home")
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodGet)

	r.HandleFunc("/timeline/{type}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		timelineType, _ := mux.Vars(req)["type"]
		maxID := req.URL.Query().Get("max_id")
		sinceID := req.URL.Query().Get("since_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeTimelinePage(ctx, w, nil, timelineType, maxID, sinceID, minID)
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

	r.HandleFunc("/likedby/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]

		err := s.ServeLikedByPage(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/retweetedby/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]

		err := s.ServeRetweetedByPage(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/following/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		id, _ := mux.Vars(req)["id"]
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeFollowingPage(ctx, w, nil, id, maxID, minID)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/followers/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		id, _ := mux.Vars(req)["id"]
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeFollowersPage(ctx, w, nil, id, maxID, minID)
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
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/unlike/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.UnLike(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/retweet/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.Retweet(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/unretweet/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)
		id, _ := mux.Vars(req)["id"]
		err := s.UnRetweet(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+id)
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/post", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		err := req.ParseMultipartForm(4 << 20)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		content := getMultipartFormValue(req.MultipartForm, "content")
		replyToID := getMultipartFormValue(req.MultipartForm, "reply_to_id")
		format := getMultipartFormValue(req.MultipartForm, "format")
		visibility := getMultipartFormValue(req.MultipartForm, "visibility")
		isNSFW := "on" == getMultipartFormValue(req.MultipartForm, "is_nsfw")

		files := req.MultipartForm.File["attachments"]

		id, err := s.PostTweet(ctx, w, nil, content, replyToID, format, visibility, isNSFW, files)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		location := "/timeline/home" + "#status-" + id
		if len(replyToID) > 0 {
			location = "/thread/" + replyToID + "#status-" + id
		}
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/notifications", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeNotificationPage(ctx, w, nil, maxID, minID)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/user/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		id, _ := mux.Vars(req)["id"]
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeUserPage(ctx, w, nil, id, maxID, minID)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/follow/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		id, _ := mux.Vars(req)["id"]

		err := s.Follow(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/unfollow/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		id, _ := mux.Vars(req)["id"]

		err := s.UnFollow(ctx, w, nil, id)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/about", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		err := s.ServeAboutPage(ctx, w, nil)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/emojis", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		err := s.ServeEmojiPage(ctx, w, nil)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/search", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		q := req.URL.Query().Get("q")
		qType := req.URL.Query().Get("type")
		offsetStr := req.URL.Query().Get("offset")

		var offset int
		var err error
		if len(offsetStr) > 1 {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil {
				s.ServeErrorPage(ctx, w, err)
				return
			}
		}

		err = s.ServeSearchPage(ctx, w, nil, q, qType, offset)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/settings", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		err := s.ServeSettingsPage(ctx, w, nil)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/settings", func(w http.ResponseWriter, req *http.Request) {
		ctx := getContextWithSession(context.Background(), req)

		visibility := req.FormValue("visibility")
		copyScope := req.FormValue("copy_scope") == "true"
		threadInNewTab := req.FormValue("thread_in_new_tab") == "true"
		maskNSFW := req.FormValue("mask_nsfw") == "true"
		settings := &model.Settings{
			DefaultVisibility: visibility,
			CopyScope:         copyScope,
			ThreadInNewTab:    threadInNewTab,
			MaskNSFW:          maskNSFW,
		}

		err := s.SaveSettings(ctx, w, nil, settings)
		if err != nil {
			s.ServeErrorPage(ctx, w, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodPost)

	r.HandleFunc("/signout", func(w http.ResponseWriter, req *http.Request) {
		// TODO remove session from database
		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   "",
			Expires: time.Now(),
		})
		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusFound)
	}).Methods(http.MethodGet)

	return r
}

func getContextWithSession(ctx context.Context, req *http.Request) context.Context {
	sessionID, err := req.Cookie("session_id")
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, "session_id", sessionID.Value)
}

func getMultipartFormValue(mf *multipart.Form, key string) (val string) {
	vals, ok := mf.Value[key]
	if !ok {
		return ""
	}
	if len(vals) < 1 {
		return ""
	}
	return vals[0]
}
