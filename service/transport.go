package service

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"bloat/model"

	"github.com/gorilla/mux"
)

const (
	HTML int = iota
	JSON
)

const (
	NOAUTH int = iota
	SESSION
	CSRF
)

func NewHandler(s *service, logger *log.Logger, staticfs fs.FS) http.Handler {
	r := mux.NewRouter()

	writeError := func(c *client, err error, t int, retry bool) {
		switch t {
		case HTML:
			c.w.WriteHeader(http.StatusInternalServerError)
			s.ErrorPage(c, err, retry)
		case JSON:
			c.w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(c.w).Encode(map[string]string{
				"error": err.Error(),
			})
		}
	}

	handle := func(f func(c *client) error, at int, rt int) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			var err error
			c := &client{
				ctx: req.Context(),
				w:   w,
				r:   req,
			}

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
			c.w.Header().Add("Content-Type", ct)

			err = c.authenticate(at)
			if err != nil {
				writeError(c, err, rt, req.Method == http.MethodGet)
				return
			}

			err = f(c)
			if err != nil {
				writeError(c, err, rt, req.Method == http.MethodGet)
				return
			}
		}
	}

	rootPage := handle(func(c *client) error {
		err := c.authenticate(SESSION)
		if err != nil {
			if err == errInvalidSession {
				c.redirect("/signin")
				return nil
			}
			return err
		}
		if !c.s.IsLoggedIn() {
			c.redirect("/signin")
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
		url, sess, err := s.NewSession(c, instance)
		if err != nil {
			return err
		}
		c.setSession(sess)
		c.redirect(url)
		return nil
	}, NOAUTH, HTML)

	timelinePage := handle(func(c *client) error {
		tType := mux.Vars(c.r)["type"]
		q := c.r.URL.Query()
		instance := q.Get("instance")
		list := q.Get("list")
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.TimelinePage(c, tType, instance, list, maxID, minID)
	}, SESSION, HTML)

	defaultTimelinePage := handle(func(c *client) error {
		c.redirect("/timeline/home")
		return nil
	}, SESSION, HTML)

	threadPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		reply := q.Get("reply")
		return s.ThreadPage(c, id, len(reply) > 1)
	}, SESSION, HTML)

	quickReplyPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		return s.QuickReplyPage(c, id)
	}, SESSION, HTML)

	likedByPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		return s.LikedByPage(c, id)
	}, SESSION, HTML)

	retweetedByPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		return s.RetweetedByPage(c, id)
	}, SESSION, HTML)

	reactionsPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		return s.ReactionsPage(c, id)
	}, SESSION, HTML)

	notificationsPage := handle(func(c *client) error {
		q := c.r.URL.Query()
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.NotificationPage(c, maxID, minID)
	}, SESSION, HTML)

	userPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		pageType := mux.Vars(c.r)["type"]
		q := c.r.URL.Query()
		maxID := q.Get("max_id")
		minID := q.Get("min_id")
		return s.UserPage(c, id, pageType, maxID, minID)
	}, SESSION, HTML)

	userSearchPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		sq := q.Get("q")
		offset, _ := strconv.Atoi(q.Get("offset"))
		return s.UserSearchPage(c, id, sq, offset)
	}, SESSION, HTML)

	mutePage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		return s.MutePage(c, id)
	}, SESSION, HTML)

	aboutPage := handle(func(c *client) error {
		return s.AboutPage(c)
	}, SESSION, HTML)

	emojisPage := handle(func(c *client) error {
		return s.EmojiPage(c)
	}, SESSION, HTML)

	searchPage := handle(func(c *client) error {
		q := c.r.URL.Query()
		sq := q.Get("q")
		qType := q.Get("type")
		offset, _ := strconv.Atoi(q.Get("offset"))
		return s.SearchPage(c, sq, qType, offset)
	}, SESSION, HTML)

	settingsPage := handle(func(c *client) error {
		return s.SettingsPage(c)
	}, SESSION, HTML)

	filtersPage := handle(func(c *client) error {
		return s.FiltersPage(c)
	}, SESSION, HTML)

	signin := handle(func(c *client) error {
		instance := c.r.FormValue("instance")
		url, sess, err := s.NewSession(c, instance)
		if err != nil {
			return err
		}
		c.setSession(sess)
		c.redirect(url)
		return nil
	}, NOAUTH, HTML)

	oauthCallback := handle(func(c *client) error {
		q := c.r.URL.Query()
		token := q.Get("code")
		err := s.Signin(c, token)
		if err != nil {
			return err
		}
		c.redirect("/")
		return nil
	}, SESSION, HTML)

	post := handle(func(c *client) error {
		content := c.r.FormValue("content")
		replyToID := c.r.FormValue("reply_to_id")
		format := c.r.FormValue("format")
		visibility := c.r.FormValue("visibility")
		subjectHeader := c.r.FormValue("subject")
		isNSFW := c.r.FormValue("is_nsfw") == "true"
		quickReply := c.r.FormValue("quickreply") == "true"
		files := c.r.MultipartForm.File["attachments"]

		id, err := s.Post(c, content, replyToID, format, visibility, subjectHeader, isNSFW, files)
		if err != nil {
			return err
		}

		var location string
		if len(replyToID) > 0 {
			if quickReply {
				location = "/quickreply/" + id + "#status-" + id
			} else {
				location = "/thread/" + replyToID + "#status-" + id
			}
		} else {
			location = c.r.FormValue("referrer")
		}
		c.redirect(location)
		return nil
	}, CSRF, HTML)

	like := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		_, err := s.Like(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	unlike := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		_, err := s.UnLike(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	retweet := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		_, err := s.Retweet(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	unretweet := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		_, err := s.UnRetweet(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	vote := handle(func(c *client) error {
		var err error
		id := mux.Vars(c.r)["id"]
		statusID := c.r.FormValue("status_id")
		choices := c.r.PostForm["choices"]
		convchoice := make([]int, len(choices))
		for i, v := range choices {
			convchoice[i], err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		}
		err = s.Vote(c, id, convchoice...)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + statusID)
		return nil
	}, CSRF, HTML)

	follow := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		var reblogs *bool
		if r, ok := q["reblogs"]; ok && len(r) > 0 {
			reblogs = new(bool)
			*reblogs = r[0] == "true"
		}
		err := s.Follow(c, id, reblogs)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unfollow := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnFollow(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	accept := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.Accept(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	reject := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.Reject(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	mute := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		notifications, _ := strconv.ParseBool(c.r.FormValue("notifications"))
		duration, err := strconv.ParseInt(c.r.FormValue("duration"), 10, 64)
		if err != nil {
			return err
		}
		err = s.Mute(c, id, notifications, duration)
		if err != nil {
			return err
		}
		c.redirect("/user/" + id)
		return nil
	}, CSRF, HTML)

	unMute := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnMute(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	block := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.Block(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unBlock := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnBlock(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	subscribe := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.Subscribe(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unSubscribe := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnSubscribe(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	settings := handle(func(c *client) error {
		visibility := c.r.FormValue("visibility")
		format := c.r.FormValue("format")
		copyScope := c.r.FormValue("copy_scope") == "true"
		threadInNewTab := c.r.FormValue("thread_in_new_tab") == "true"
		hideAttachments := c.r.FormValue("hide_attachments") == "true"
		maskNSFW := c.r.FormValue("mask_nsfw") == "true"
		ni, _ := strconv.Atoi(c.r.FormValue("notification_interval"))
		fluorideMode := c.r.FormValue("fluoride_mode") == "true"
		darkMode := c.r.FormValue("dark_mode") == "true"
		antiDopamineMode := c.r.FormValue("anti_dopamine_mode") == "true"
		hideUnsupportedNotifs := c.r.FormValue("hide_unsupported_notifs") == "true"
		css := c.r.FormValue("css")

		settings := &model.Settings{
			DefaultVisibility:     visibility,
			DefaultFormat:         format,
			CopyScope:             copyScope,
			ThreadInNewTab:        threadInNewTab,
			HideAttachments:       hideAttachments,
			MaskNSFW:              maskNSFW,
			NotificationInterval:  ni,
			FluorideMode:          fluorideMode,
			DarkMode:              darkMode,
			AntiDopamineMode:      antiDopamineMode,
			HideUnsupportedNotifs: hideUnsupportedNotifs,
			CSS:                   css,
		}

		err := s.SaveSettings(c, settings)
		if err != nil {
			return err
		}
		c.redirect("/")
		return nil
	}, CSRF, HTML)

	muteConversation := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.MuteConversation(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unMuteConversation := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnMuteConversation(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	delete := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.Delete(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	readNotifications := handle(func(c *client) error {
		q := c.r.URL.Query()
		maxID := q.Get("max_id")
		err := s.ReadNotifications(c, maxID)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	bookmark := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		err := s.Bookmark(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	unBookmark := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		rid := c.r.FormValue("retweeted_by_id")
		err := s.UnBookmark(c, id)
		if err != nil {
			return err
		}
		if len(rid) > 0 {
			id = rid
		}
		c.redirect(c.r.FormValue("referrer") + "#status-" + id)
		return nil
	}, CSRF, HTML)

	filter := handle(func(c *client) error {
		phrase := c.r.FormValue("phrase")
		wholeWord := c.r.FormValue("whole_word") == "true"
		err := s.Filter(c, phrase, wholeWord)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	unFilter := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.UnFilter(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	listsPage := handle(func(c *client) error {
		return s.ListsPage(c)
	}, SESSION, HTML)

	addList := handle(func(c *client) error {
		title := c.r.FormValue("title")
		err := s.AddList(c, title)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	removeList := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		err := s.RemoveList(c, id)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	renameList := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		title := c.r.FormValue("title")
		err := s.RenameList(c, id, title)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	listPage := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		sq := q.Get("q")
		return s.ListPage(c, id, sq)
	}, SESSION, HTML)

	listAddUser := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		uid := q.Get("uid")
		err := s.ListAddUser(c, id, uid)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	listRemoveUser := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		q := c.r.URL.Query()
		uid := q.Get("uid")
		err := s.ListRemoveUser(c, id, uid)
		if err != nil {
			return err
		}
		c.redirect(c.r.FormValue("referrer"))
		return nil
	}, CSRF, HTML)

	signout := handle(func(c *client) error {
		c.unsetSession()
		c.redirect("/")
		return nil
	}, CSRF, HTML)

	fLike := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		count, err := s.Like(c, id)
		if err != nil {
			return err
		}
		return c.writeJson(count)
	}, CSRF, JSON)

	fUnlike := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		count, err := s.UnLike(c, id)
		if err != nil {
			return err
		}
		return c.writeJson(count)
	}, CSRF, JSON)

	fRetweet := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		count, err := s.Retweet(c, id)
		if err != nil {
			return err
		}
		return c.writeJson(count)
	}, CSRF, JSON)

	fUnretweet := handle(func(c *client) error {
		id := mux.Vars(c.r)["id"]
		count, err := s.UnRetweet(c, id)
		if err != nil {
			return err
		}
		return c.writeJson(count)
	}, CSRF, JSON)

	r.HandleFunc("/", rootPage).Methods(http.MethodGet)
	r.HandleFunc("/nav", navPage).Methods(http.MethodGet)
	r.HandleFunc("/signin", signinPage).Methods(http.MethodGet)
	r.HandleFunc("/timeline/{type}", timelinePage).Methods(http.MethodGet)
	r.HandleFunc("/timeline", defaultTimelinePage).Methods(http.MethodGet)
	r.HandleFunc("/thread/{id}", threadPage).Methods(http.MethodGet)
	r.HandleFunc("/quickreply/{id}", quickReplyPage).Methods(http.MethodGet)
	r.HandleFunc("/likedby/{id}", likedByPage).Methods(http.MethodGet)
	r.HandleFunc("/retweetedby/{id}", retweetedByPage).Methods(http.MethodGet)
	r.HandleFunc("/reactions/{id}", reactionsPage).Methods(http.MethodGet)
	r.HandleFunc("/notifications", notificationsPage).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", userPage).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}/{type}", userPage).Methods(http.MethodGet)
	r.HandleFunc("/usersearch/{id}", userSearchPage).Methods(http.MethodGet)
	r.HandleFunc("/mute/{id}", mutePage).Methods(http.MethodGet)
	r.HandleFunc("/about", aboutPage).Methods(http.MethodGet)
	r.HandleFunc("/emojis", emojisPage).Methods(http.MethodGet)
	r.HandleFunc("/search", searchPage).Methods(http.MethodGet)
	r.HandleFunc("/settings", settingsPage).Methods(http.MethodGet)
	r.HandleFunc("/filters", filtersPage).Methods(http.MethodGet)
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
	r.HandleFunc("/filter", filter).Methods(http.MethodPost)
	r.HandleFunc("/unfilter/{id}", unFilter).Methods(http.MethodPost)
	r.HandleFunc("/lists", listsPage).Methods(http.MethodGet)
	r.HandleFunc("/list", addList).Methods(http.MethodPost)
	r.HandleFunc("/list/{id}", listPage).Methods(http.MethodGet)
	r.HandleFunc("/list/{id}/remove", removeList).Methods(http.MethodPost)
	r.HandleFunc("/list/{id}/rename", renameList).Methods(http.MethodPost)
	r.HandleFunc("/list/{id}/adduser", listAddUser).Methods(http.MethodPost)
	r.HandleFunc("/list/{id}/removeuser", listRemoveUser).Methods(http.MethodPost)
	r.HandleFunc("/signout", signout).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/like/{id}", fLike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unlike/{id}", fUnlike).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/retweet/{id}", fRetweet).Methods(http.MethodPost)
	r.HandleFunc("/fluoride/unretweet/{id}", fUnretweet).Methods(http.MethodPost)
	r.PathPrefix("/static").Handler(http.FileServer(http.FS(staticfs)))

	return r
}
