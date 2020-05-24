package service

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"bloat/model"

	"github.com/gorilla/mux"
)

const (
	sessionExp = 365 * 24 * time.Hour
)

func newClient(w io.Writer, req *http.Request, csrfToken string) *model.Client {
	var sessionID string
	if req != nil {
		c, err := req.Cookie("session_id")
		if err == nil {
			sessionID = c.Value
		}
	}
	return &model.Client{
		Writer: w,
		Ctx: model.ClientCtx{
			SessionID: sessionID,
			CSRFToken: csrfToken,
		},
	}
}

func setSessionCookie(w http.ResponseWriter, sessionID string, exp time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: time.Now().Add(exp),
	})
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

func serveJson(w io.Writer, data interface{}) (err error) {
	var d = make(map[string]interface{})
	d["data"] = data
	return json.NewEncoder(w).Encode(d)
}

func serveJsonError(w http.ResponseWriter, err error) {
	var d = make(map[string]interface{})
	d["error"] = err.Error()
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(d)
	return
}

func NewHandler(s Service, staticDir string) http.Handler {
	r := mux.NewRouter()

	rootPage := func(w http.ResponseWriter, req *http.Request) {
		sessionID, _ := req.Cookie("session_id")
		if sessionID != nil && len(sessionID.Value) > 0 {
			c := newClient(w, req, "")
			err := s.ServeRootPage(c)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				s.ServeErrorPage(c, err)
				return
			}
		} else {
			w.Header().Add("Location", "/signin")
			w.WriteHeader(http.StatusFound)
		}
	}

	navPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		err := s.ServeNavPage(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	signinPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, nil, "")
		instance, ok := s.SingleInstance()
		if ok {
			url, sessionID, err := s.NewSession(instance)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				s.ServeErrorPage(c, err)
				return
			}

			setSessionCookie(w, sessionID, sessionExp)
			w.Header().Add("Location", url)
			w.WriteHeader(http.StatusFound)
		} else {
			err := s.ServeSigninPage(c)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				s.ServeErrorPage(c, err)
				return
			}
		}
	}

	timelinePage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		tType, _ := mux.Vars(req)["type"]
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeTimelinePage(c, tType, maxID, minID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	defaultTimelinePage := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Location", "/timeline/home")
		w.WriteHeader(http.StatusFound)
	}

	threadPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		id, _ := mux.Vars(req)["id"]
		reply := req.URL.Query().Get("reply")

		err := s.ServeThreadPage(c, id, len(reply) > 1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	likedByPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		id, _ := mux.Vars(req)["id"]

		err := s.ServeLikedByPage(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	retweetedByPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		id, _ := mux.Vars(req)["id"]

		err := s.ServeRetweetedByPage(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	notificationsPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeNotificationPage(c, maxID, minID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	userPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		id, _ := mux.Vars(req)["id"]
		pageType, _ := mux.Vars(req)["type"]
		maxID := req.URL.Query().Get("max_id")
		minID := req.URL.Query().Get("min_id")

		err := s.ServeUserPage(c, id, pageType, maxID, minID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	userSearchPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		id, _ := mux.Vars(req)["id"]
		q := req.URL.Query().Get("q")
		offsetStr := req.URL.Query().Get("offset")

		var offset int
		var err error
		if len(offsetStr) > 1 {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				s.ServeErrorPage(c, err)
				return
			}
		}

		err = s.ServeUserSearchPage(c, id, q, offset)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	aboutPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")

		err := s.ServeAboutPage(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	emojisPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")

		err := s.ServeEmojiPage(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	searchPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		q := req.URL.Query().Get("q")
		qType := req.URL.Query().Get("type")
		offsetStr := req.URL.Query().Get("offset")

		var offset int
		var err error
		if len(offsetStr) > 1 {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				s.ServeErrorPage(c, err)
				return
			}
		}

		err = s.ServeSearchPage(c, q, qType, offset)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	settingsPage := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")

		err := s.ServeSettingsPage(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}
	}

	signin := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, nil, "")
		instance := req.FormValue("instance")

		url, sessionID, err := s.NewSession(instance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		setSessionCookie(w, sessionID, sessionExp)
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}

	oauthCallback := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		token := req.URL.Query().Get("code")

		_, _, err := s.Signin(c, "", token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusFound)
	}

	post := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, "")
		err := req.ParseMultipartForm(4 << 20)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		c = newClient(w, req,
			getMultipartFormValue(req.MultipartForm, "csrf_token"))
		content := getMultipartFormValue(req.MultipartForm, "content")
		replyToID := getMultipartFormValue(req.MultipartForm, "reply_to_id")
		format := getMultipartFormValue(req.MultipartForm, "format")
		visibility := getMultipartFormValue(req.MultipartForm, "visibility")
		isNSFW := "on" == getMultipartFormValue(req.MultipartForm, "is_nsfw")
		files := req.MultipartForm.File["attachments"]

		id, err := s.Post(c, content, replyToID, format, visibility, isNSFW, files)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		location := req.Header.Get("Referer")
		if len(replyToID) > 0 {
			location = "/thread/" + replyToID + "#status-" + id
		}
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusFound)
	}

	like := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]
		retweetedByID := req.FormValue("retweeted_by_id")

		_, err := s.Like(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		rID := id
		if len(retweetedByID) > 0 {
			rID = retweetedByID
		}
		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+rID)
		w.WriteHeader(http.StatusFound)
	}

	unlike := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]
		retweetedByID := req.FormValue("retweeted_by_id")

		_, err := s.UnLike(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		rID := id
		if len(retweetedByID) > 0 {
			rID = retweetedByID
		}
		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+rID)
		w.WriteHeader(http.StatusFound)
	}

	retweet := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]
		retweetedByID := req.FormValue("retweeted_by_id")

		_, err := s.Retweet(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		rID := id
		if len(retweetedByID) > 0 {
			rID = retweetedByID
		}
		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+rID)
		w.WriteHeader(http.StatusFound)
	}

	unretweet := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]
		retweetedByID := req.FormValue("retweeted_by_id")

		_, err := s.UnRetweet(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		rID := id
		if len(retweetedByID) > 0 {
			rID = retweetedByID
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+rID)
		w.WriteHeader(http.StatusFound)
	}

	vote := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]
		statusID := req.FormValue("status_id")
		choices, _ := req.PostForm["choices"]

		err := s.Vote(c, id, choices)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer")+"#status-"+statusID)
		w.WriteHeader(http.StatusFound)
	}

	follow := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		var reblogs *bool
		r, ok := req.URL.Query()["reblogs"]
		if ok && len(r) > 0 {
			reblogs = new(bool)
			*reblogs = r[0] == "true"
		}

		err := s.Follow(c, id, reblogs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	unfollow := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.UnFollow(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	mute := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.Mute(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	unMute := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.UnMute(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	block := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.Block(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	unBlock := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.UnBlock(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	subscribe := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.Subscribe(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	unSubscribe := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.UnSubscribe(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	settings := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		visibility := req.FormValue("visibility")
		copyScope := req.FormValue("copy_scope") == "true"
		threadInNewTab := req.FormValue("thread_in_new_tab") == "true"
		hideAttachments := req.FormValue("hide_attachments") == "true"
		maskNSFW := req.FormValue("mask_nsfw") == "true"
		arn := req.FormValue("auto_refresh_notifications") == "true"
		fluorideMode := req.FormValue("fluoride_mode") == "true"
		darkMode := req.FormValue("dark_mode") == "true"

		settings := &model.Settings{
			DefaultVisibility:        visibility,
			CopyScope:                copyScope,
			ThreadInNewTab:           threadInNewTab,
			HideAttachments:          hideAttachments,
			MaskNSFW:                 maskNSFW,
			AutoRefreshNotifications: arn,
			FluorideMode:             fluorideMode,
			DarkMode:                 darkMode,
		}

		err := s.SaveSettings(c, settings)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusFound)
	}

	muteConversation := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.MuteConversation(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	unMuteConversation := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.UnMuteConversation(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	delete := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		err := s.Delete(c, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	readNotifications := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		maxID := req.URL.Query().Get("max_id")

		err := s.ReadNotifications(c, maxID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.ServeErrorPage(c, err)
			return
		}

		w.Header().Add("Location", req.Header.Get("Referer"))
		w.WriteHeader(http.StatusFound)
	}

	signout := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))

		s.Signout(c)

		setSessionCookie(w, "", 0)
		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusFound)
	}

	fLike := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		count, err := s.Like(c, id)
		if err != nil {
			serveJsonError(w, err)
			return
		}

		err = serveJson(w, count)
		if err != nil {
			serveJsonError(w, err)
			return
		}
	}

	fUnlike := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		count, err := s.UnLike(c, id)
		if err != nil {
			serveJsonError(w, err)
			return
		}

		err = serveJson(w, count)
		if err != nil {
			serveJsonError(w, err)
			return
		}
	}

	fRetweet := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		count, err := s.Retweet(c, id)
		if err != nil {
			serveJsonError(w, err)
			return
		}

		err = serveJson(w, count)
		if err != nil {
			serveJsonError(w, err)
			return
		}
	}

	fUnretweet := func(w http.ResponseWriter, req *http.Request) {
		c := newClient(w, req, req.FormValue("csrf_token"))
		id, _ := mux.Vars(req)["id"]

		count, err := s.UnRetweet(c, id)
		if err != nil {
			serveJsonError(w, err)
			return
		}

		err = serveJson(w, count)
		if err != nil {
			serveJsonError(w, err)
			return
		}
	}

	r.HandleFunc("/", rootPage).Methods(http.MethodGet)
	r.HandleFunc("/nav", navPage).Methods(http.MethodGet)
	r.HandleFunc("/signin", signinPage).Methods(http.MethodGet)
	r.HandleFunc("/timeline/{type}", timelinePage).Methods(http.MethodGet)
	r.HandleFunc("/timeline", defaultTimelinePage).Methods(http.MethodGet)
	r.HandleFunc("/thread/{id}", threadPage).Methods(http.MethodGet)
	r.HandleFunc("/likedby/{id}", likedByPage).Methods(http.MethodGet)
	r.HandleFunc("/retweetedby/{id}", retweetedByPage).Methods(http.MethodGet)
	r.HandleFunc("/notifications", notificationsPage).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", userPage).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}/{type}", userPage).Methods(http.MethodGet)
	r.HandleFunc("/usersearch/{id}", userSearchPage).Methods(http.MethodGet)
	r.HandleFunc("/about", aboutPage).Methods(http.MethodGet)
	r.HandleFunc("/emojis", emojisPage).Methods(http.MethodGet)
	r.HandleFunc("/search", searchPage).Methods(http.MethodGet)
	r.HandleFunc("/settings", settingsPage).Methods(http.MethodGet)
	r.HandleFunc("/signin", signin).Methods(http.MethodPost)
	r.HandleFunc("/oauth_callback", oauthCallback).Methods(http.MethodGet)
	r.HandleFunc("/post", post).Methods(http.MethodPost)
	r.HandleFunc("/like/{id}", like).Methods(http.MethodPost)
	r.HandleFunc("/unlike/{id}", unlike).Methods(http.MethodPost)
	r.HandleFunc("/retweet/{id}", retweet).Methods(http.MethodPost)
	r.HandleFunc("/unretweet/{id}", unretweet).Methods(http.MethodPost)
	r.HandleFunc("/vote/{id}", vote).Methods(http.MethodPost)
	r.HandleFunc("/follow/{id}", follow).Methods(http.MethodPost)
	r.HandleFunc("/unfollow/{id}", unfollow).Methods(http.MethodPost)
	r.HandleFunc("/mute/{id}", mute).Methods(http.MethodPost)
	r.HandleFunc("/unmute/{id}", unMute).Methods(http.MethodPost)
	r.HandleFunc("/block/{id}", block).Methods(http.MethodPost)
	r.HandleFunc("/unblock/{id}", unBlock).Methods(http.MethodPost)
	r.HandleFunc("/subscribe/{id}", subscribe).Methods(http.MethodPost)
	r.HandleFunc("/unsubscribe/{id}", unSubscribe).Methods(http.MethodPost)
	r.HandleFunc("/settings", settings).Methods(http.MethodPost)
	r.HandleFunc("/muteconv/{id}", muteConversation).Methods(http.MethodPost)
	r.HandleFunc("/unmuteconv/{id}", unMuteConversation).Methods(http.MethodPost)
	r.HandleFunc("/delete/{id}", delete).Methods(http.MethodPost)
	r.HandleFunc("/notifications/read", readNotifications).Methods(http.MethodPost)
	r.HandleFunc("/signout", signout).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/like/{id}", fLike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unlike/{id}", fUnlike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/retweet/{id}", fRetweet).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unretweet/{id}", fUnretweet).Methods(http.MethodPost)
	r.PathPrefix("/static").Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(staticDir))))

	return r
}
