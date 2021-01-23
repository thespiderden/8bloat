package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"bloat/mastodon"
	"bloat/model"

	"github.com/gorilla/mux"
)

var (
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

const (
	sessionExp = 365 * 24 * time.Hour
)

type respType int

const (
	HTML respType = iota
	JSON
)

type authType int

const (
	NOAUTH authType = iota
	SESSION
	CSRF
)

type client struct {
	*mastodon.Client
	http.ResponseWriter
	Req       *http.Request
	CSRFToken string
	Session   model.Session
}

func (c *client) url() string {
	return c.Req.URL.RequestURI()
}

func setSessionCookie(w http.ResponseWriter, sid string, exp time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sid,
		Expires: time.Now().Add(exp),
	})
}

func writeJson(c *client, data interface{}) error {
	return json.NewEncoder(c).Encode(map[string]interface{}{
		"data": data,
	})
}

func redirect(c *client, url string) {
	c.Header().Add("Location", url)
	c.WriteHeader(http.StatusFound)
}

func NewHandler(s *service, logger *log.Logger, staticDir string) http.Handler {
	r := mux.NewRouter()

	writeError := func(c *client, err error, t respType) {
		switch t {
		case HTML:
			c.WriteHeader(http.StatusInternalServerError)
			s.ErrorPage(c, err)
		case JSON:
			c.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(c).Encode(map[string]string{
				"error": err.Error(),
			})
		}
	}

	authenticate := func(c *client, t authType) error {
		if t >= SESSION {
			cookie, err := c.Req.Cookie("session_id")
			if err != nil || len(cookie.Value) < 1 {
				return errInvalidSession
			}
			c.Session, err = s.sessionRepo.Get(cookie.Value)
			if err != nil {
				return errInvalidSession
			}
			app, err := s.appRepo.Get(c.Session.InstanceDomain)
			if err != nil {
				return err
			}
			c.Client = mastodon.NewClient(&mastodon.Config{
				Server:       app.InstanceURL,
				ClientID:     app.ClientID,
				ClientSecret: app.ClientSecret,
				AccessToken:  c.Session.AccessToken,
			})
		}
		if t >= CSRF {
			c.CSRFToken = c.Req.FormValue("csrf_token")
			if len(c.CSRFToken) < 1 || c.CSRFToken != c.Session.CSRFToken {
				return errInvalidCSRFToken
			}
		}
		return nil
	}

	handle := func(f func(c *client) error, at authType, rt respType) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			var err error
			c := &client{Req: req, ResponseWriter: w}

			defer func(begin time.Time) {
				logger.Printf("path=%s, err=%v, took=%v\n",
					req.URL.Path, err, time.Since(begin))
			}(time.Now())

			var ct string
			switch rt {
			case HTML:
				ct = "text/html; charset=utf-8"
			case JSON:
				ct = "application/json"
			}
			c.Header().Add("Content-Type", ct)

			err = authenticate(c, at)
			if err != nil {
				writeError(c, err, rt)
				return
			}

			err = f(c)
			if err != nil {
				writeError(c, err, rt)
				return
			}
		}
	}

	rootPage := handle(func(c *client) error {
		sid, _ := c.Req.Cookie("session_id")
		if sid == nil || len(sid.Value) < 0 {
			redirect(c, "/signin")
			return nil
		}
		session, err := s.sessionRepo.Get(sid.Value)
		if err != nil {
			if err == errInvalidSession {
				redirect(c, "/signin")
				return nil
			}
			return err
		}
		if len(session.AccessToken) < 1 {
			redirect(c, "/signin")
			return nil
		}
		return s.RootPage(c)
	}, NOAUTH, HTML)

	navPage := handle(func(c *client) error {
		return s.NavPage(c)
	}, SESSION, HTML)

	signinPage := handle(func(c *client) error {
		instance, ok := s.SingleInstance()
		if !ok {
			return s.SigninPage(c)
		}
		url, sid, err := s.NewSession(instance)
		if err != nil {
			return err
		}
		setSessionCookie(c, sid, sessionExp)
		redirect(c, url)
		return nil
	}, NOAUTH, HTML)

	timelinePage := handle(func(c *client) error {
		tType, _ := mux.Vars(c.Req)["type"]
		q := c.Req.URL.Query()
		instance := q.Get("instance")
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.TimelinePage(c, tType, instance, maxID, minID)
	}, SESSION, HTML)

	defaultTimelinePage := handle(func(c *client) error {
		redirect(c, "/timeline/home")
		return nil
	}, SESSION, HTML)

	threadPage := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		q := c.Req.URL.Query()
		reply := q.Get("reply")
		return s.ThreadPage(c, id, len(reply) > 1)
	}, SESSION, HTML)

	likedByPage := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		return s.LikedByPage(c, id)
	}, SESSION, HTML)

	retweetedByPage := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		return s.RetweetedByPage(c, id)
	}, SESSION, HTML)

	notificationsPage := handle(func(c *client) error {
		q := c.Req.URL.Query()
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.NotificationPage(c, maxID, minID)
	}, SESSION, HTML)

	userPage := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		pageType, _ := mux.Vars(c.Req)["type"]
		q := c.Req.URL.Query()
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.UserPage(c, id, pageType, maxID, minID)
	}, SESSION, HTML)

	userSearchPage := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		q := c.Req.URL.Query()
		sq := q.Get("q")
		offset, _ := strconv.Atoi(q.Get("offset"))
		return s.UserSearchPage(c, id, sq, offset)
	}, SESSION, HTML)

	aboutPage := handle(func(c *client) error {
		return s.AboutPage(c)
	}, SESSION, HTML)

	emojisPage := handle(func(c *client) error {
		return s.EmojiPage(c)
	}, SESSION, HTML)

	searchPage := handle(func(c *client) error {
		q := c.Req.URL.Query()
		sq := q.Get("q")
		qType := q.Get("type")
		offset, _ := strconv.Atoi(q.Get("offset"))
		return s.SearchPage(c, sq, qType, offset)
	}, SESSION, HTML)

	settingsPage := handle(func(c *client) error {
		return s.SettingsPage(c)
	}, SESSION, HTML)

	signin := handle(func(c *client) error {
		instance := c.Req.FormValue("instance")
		url, sid, err := s.NewSession(instance)
		if err != nil {
			return err
		}
		setSessionCookie(c, sid, sessionExp)
		redirect(c, url)
		return nil
	}, NOAUTH, HTML)

	oauthCallback := handle(func(c *client) error {
		q := c.Req.URL.Query()
		token := q.Get("code")
		token, userID, err := s.Signin(c, token)
		if err != nil {
			return err
		}

		c.Session.AccessToken = token
		c.Session.UserID = userID
		err = s.sessionRepo.Add(c.Session)
		if err != nil {
			return err
		}

		redirect(c, "/")
		return nil
	}, SESSION, HTML)

	post := handle(func(c *client) error {
		content := c.Req.FormValue("content")
		replyToID := c.Req.FormValue("reply_to_id")
		format := c.Req.FormValue("format")
		visibility := c.Req.FormValue("visibility")
		isNSFW := c.Req.FormValue("is_nsfw") == "on"
		files := c.Req.MultipartForm.File["attachments"]

		id, err := s.Post(c, content, replyToID, format, visibility, isNSFW, files)
		if err != nil {
			return err
		}

		location := c.Req.FormValue("referrer")
		if len(replyToID) > 0 {
			location = "/thread/" + replyToID + "#status-" + id
		}
		redirect(c, location)
		return nil
	}, CSRF, HTML)

	like := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		_, err := s.Like(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	unlike := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		_, err := s.UnLike(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	retweet := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		_, err := s.Retweet(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	unretweet := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		_, err := s.UnRetweet(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	vote := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		statusID := c.Req.FormValue("status_id")
		choices, _ := c.Req.PostForm["choices"]
		err := s.Vote(c, id, choices)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+statusID)
		return nil
	}, CSRF, HTML)

	follow := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		q := c.Req.URL.Query()
		var reblogs *bool
		if r, ok := q["reblogs"]; ok && len(r) > 0 {
			reblogs = new(bool)
			*reblogs = r[0] == "true"
		}
		err := s.Follow(c, id, reblogs)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unfollow := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.UnFollow(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	accept := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Accept(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	reject := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Reject(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	mute := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Mute(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unMute := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.UnMute(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	block := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Block(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unBlock := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.UnBlock(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	subscribe := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Subscribe(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unSubscribe := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.UnSubscribe(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	settings := handle(func(c *client) error {
		visibility := c.Req.FormValue("visibility")
		format := c.Req.FormValue("format")
		copyScope := c.Req.FormValue("copy_scope") == "true"
		threadInNewTab := c.Req.FormValue("thread_in_new_tab") == "true"
		hideAttachments := c.Req.FormValue("hide_attachments") == "true"
		maskNSFW := c.Req.FormValue("mask_nsfw") == "true"
		ni, _ := strconv.Atoi(c.Req.FormValue("notification_interval"))
		fluorideMode := c.Req.FormValue("fluoride_mode") == "true"
		darkMode := c.Req.FormValue("dark_mode") == "true"
		antiDopamineMode := c.Req.FormValue("anti_dopamine_mode") == "true"

		settings := &model.Settings{
			DefaultVisibility:    visibility,
			DefaultFormat:        format,
			CopyScope:            copyScope,
			ThreadInNewTab:       threadInNewTab,
			HideAttachments:      hideAttachments,
			MaskNSFW:             maskNSFW,
			NotificationInterval: ni,
			FluorideMode:         fluorideMode,
			DarkMode:             darkMode,
			AntiDopamineMode:     antiDopamineMode,
		}

		err := s.SaveSettings(c, settings)
		if err != nil {
			return err
		}
		redirect(c, "/")
		return nil
	}, CSRF, HTML)

	muteConversation := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.MuteConversation(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unMuteConversation := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.UnMuteConversation(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	delete := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		err := s.Delete(c, id)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	readNotifications := handle(func(c *client) error {
		q := c.Req.URL.Query()
		maxID := q.Get("max_id")
		err := s.ReadNotifications(c, maxID)
		if err != nil {
			return err
		}
		redirect(c, c.Req.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	bookmark := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		err := s.Bookmark(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	unBookmark := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		rid := c.Req.FormValue("retweeted_by_id")
		err := s.UnBookmark(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		redirect(c, c.Req.FormValue("referrer")+"#status-"+id)
		return nil
	}, CSRF, HTML)

	signout := handle(func(c *client) error {
		s.Signout(c)
		setSessionCookie(c, "", 0)
		redirect(c, "/")
		return nil
	}, CSRF, HTML)

	fLike := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		count, err := s.Like(c, id)
		if err != nil {
			return err
		}
		return writeJson(c, count)
	}, CSRF, JSON)

	fUnlike := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		count, err := s.UnLike(c, id)
		if err != nil {
			return err
		}
		return writeJson(c, count)
	}, CSRF, JSON)

	fRetweet := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		count, err := s.Retweet(c, id)
		if err != nil {
			return err
		}
		return writeJson(c, count)
	}, CSRF, JSON)

	fUnretweet := handle(func(c *client) error {
		id, _ := mux.Vars(c.Req)["id"]
		count, err := s.UnRetweet(c, id)
		if err != nil {
			return err
		}
		return writeJson(c, count)
	}, CSRF, JSON)

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
	r.HandleFunc("/accept/{id}", accept).Methods(http.MethodPost)
	r.HandleFunc("/reject/{id}", reject).Methods(http.MethodPost)
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
	r.HandleFunc("/bookmark/{id}", bookmark).Methods(http.MethodPost)
	r.HandleFunc("/unbookmark/{id}", unBookmark).Methods(http.MethodPost)
	r.HandleFunc("/signout", signout).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/like/{id}", fLike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unlike/{id}", fUnlike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/retweet/{id}", fRetweet).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unretweet/{id}", fUnretweet).Methods(http.MethodPost)
	r.PathPrefix("/static").Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(staticDir))))

	return r
}
