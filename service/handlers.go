package service

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"spiderden.org/8b/conf"
	"spiderden.org/8b/render"
	"spiderden.org/masta"
)

const (
	noAuth = iota
	noCSRF
	noType
)

var router = httprouter.New()

type handle struct {
	meth   string
	am     authMode
	f      handler
	notype bool
}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := httprouter.ParamsFromContext(r.Context())

	t := &Transaction{
		Ctx:  r.Context(),
		W:    w,
		R:    r,
		Conf: conf.Get(),
		Vars: make(map[string]string, len(vars)),
		Qry:  make(map[string]string, len(r.URL.Query())),
	}

	err = t.authenticate(h.am)
	t.Rctx.W = w
	if err != nil {
		eerr := render.ErrorPage(t.Rctx, err, true)
		if eerr != nil {
			log.Println("error responding with error page:", err, eerr)
		}
	}

	for _, v := range vars {
		t.Vars[v.Key] = v.Value
	}

	for k, v := range r.URL.Query() {
		t.Qry[k] = v[0]
	}

	defer func(begin time.Time) {
		log.Printf("path=%s, err=%v, took=%v\n",
			r.URL.Path, err, time.Since(begin))
	}(time.Now())

	if !h.notype {
		w.Header().Set("Content-type", "text/html; charset=utf-8")
	}

	conf := conf.Get()

	t.W.Header().Set("Cache-Control", "private")
	t.W.Header().Set("Content-Security-Policy",
		"default-src "+conf.ClientWebsite+"/;"+
			"style-src "+conf.ClientWebsite+"/session/css "+conf.ClientWebsite+"/static/;"+
			"script-src "+conf.ClientWebsite+"/static/;"+
			"img-src *;"+
			"media-src *",
	)
	t.W.Header().Set("Referer-Policy", "same-origin")

	if t.Session != nil {
		t.Rctx.Conf = conf
		t.Rctx.UserID = t.Session.UserID
		t.Rctx.Settings = t.Session.Settings
		t.Rctx.CSRFToken = t.Session.CSRFToken
	}

	err = h.f(t)
	if err != nil {
		eerr := render.ErrorPage(t.Rctx, err, true)
		if eerr != nil {
			log.Println("error responding with error page:", err, eerr)
		}
	}
}

func reg(h handler, meth string, path string, opts ...int) {
	handle := handle{}
	handle.meth = meth

	noauth := false
	nocsrf := false

	for _, v := range opts {
		switch v {
		case noAuth:
			noauth = true
		case noCSRF:
			nocsrf = true
		case noType:
			handle.notype = true
		}
	}

	switch {
	case noauth && nocsrf || (meth == "GET" && noauth):
		handle.am = authAnon
	case nocsrf:
		handle.am = authSess
	case !noauth && !nocsrf:
		if meth == "POST" {
			handle.am = authSessCSRF
		} else {
			handle.am = authSess
		}
	default:
		panic("invalid registration of handler for " + path)
	}

	handle.f = h

	router.Handler(meth, path, handle)
}

type handler func(t *Transaction) error

func init() { reg(handleRoot, http.MethodGet, "/", noAuth, noCSRF) }
func handleRoot(t *Transaction) error {
	if !t.Session.IsLoggedIn() {
		t.redirect("/signin")
		return nil
	}

	return render.RootPage(t.Rctx)
}

func init() { reg(handleNav, http.MethodGet, "/nav") }
func handleNav(t *Transaction) error {
	user, err := t.GetAccountCurrentUser(t.Ctx)
	if err != nil {
		return err
	}

	return render.NavPage(t.Rctx, user)
}

func init() { reg(handleSigninGet, http.MethodGet, "/signin", noAuth) }
func handleSigninGet(t *Transaction) error {
	instance, single := t.Conf.SingleInstance()
	if !single {
		return render.SigninPage(t.Rctx)
	}

	url, sess, err := newSession(t, instance)
	if err != nil {
		return err
	}

	t.setSession(sess)
	t.redirect(url)
	return nil
}

func init() { reg(handleTimeline, http.MethodGet, "/timeline/:type") }
func handleTimeline(t *Transaction) error {
	tType := t.Vars["type"]
	instance := t.Qry["instance"]
	list := t.Qry["list"]
	maxID := t.Qry["max_id"]
	minID := t.Qry["min_id"]

	var nextLink, prevLink, title string
	var statuses []*masta.Status
	pg := masta.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: conf.MaxPagination,
	}

	var err error

	switch tType {
	default:
		return errInvalidArgument
	case "home":
		statuses, err = t.GetTimelineHome(t.Ctx, &pg)
		if err != nil {
			return err
		}
		title = "Timeline"
	case "direct":
		statuses, err = t.GetTimelineDirect(t.Ctx, &pg)
		if err != nil {
			return err
		}
		title = "Direct Timeline"
	case "local":
		statuses, err = t.GetTimelinePublic(t.Ctx, true, &pg)
		if err != nil {
			return err
		}
		title = "Local Timeline"
	case "remote":
		if len(instance) > 0 {
			statuses, err = t.PlGetTimelineRemote(t.Ctx, instance, &pg)
			if err != nil {
				return err
			}
		}
		title = "Remote Timeline"
	case "twkn":
		statuses, err = t.GetTimelinePublic(t.Ctx, false, &pg)
		if err != nil {
			return err
		}
		title = "The Whole Known Network"
	case "list":
		statuses, err = t.GetTimelineList(t.Ctx, list, &pg)
		if err != nil {
			return err
		}
		list, err := t.GetList(t.Ctx, list)
		if err != nil {
			return err
		}
		title = "List Timeline - " + list.Title
	}

	if (len(maxID) > 0 || len(minID) > 0) && len(statuses) > 0 {
		v := make(url.Values)
		v.Set("min_id", statuses[0].ID)
		if len(instance) > 0 {
			v.Set("instance", instance)
		}
		if len(list) > 0 {
			v.Set("list", list)
		}
		prevLink = "/timeline/" + tType + "?" + v.Encode()
	}

	if len(minID) > 0 || (len(pg.MaxID) > 0 && len(statuses) == 20) {
		v := make(url.Values)
		v.Set("max_id", pg.MaxID)
		if len(instance) > 0 {
			v.Set("instance", instance)
		}
		if len(list) > 0 {
			v.Set("list", list)
		}
		nextLink = "/timeline/" + tType + "?" + v.Encode()
	}

	data := &render.TimelineData{
		Title:    title,
		Type:     tType,
		Instance: instance,
		Statuses: statuses,
		NextLink: nextLink,
		PrevLink: prevLink,
	}

	return render.TimelinePage(t.Rctx, data)
}

func init() { reg(handleDefaultTimeline, http.MethodGet, "/timeline") }
func handleDefaultTimeline(t *Transaction) error {
	t.redirect("/timeline/home")
	return nil
}

func init() { reg(handleThread, http.MethodGet, "/thread/:id") }
func handleThread(t *Transaction) error {
	reply := len(t.Qry["reply"]) > 0
	edit := len(t.Qry["edit"]) > 0

	status, err := t.Client.GetStatus(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	context, err := t.Client.GetStatusContext(t.Ctx, status.ID)
	if err != nil {
		return err
	}

	var src *masta.Source
	if edit {
		src, err = t.Client.GetStatusSource(t.Ctx, status.ID)
		if err != nil {
			return err
		}
	}

	return render.ThreadPage(t.Rctx, status, context, (edit || reply), src)
}

func init() { reg(handleQuickReply, http.MethodGet, "/quickreply/:id") }
func handleQuickReply(t *Transaction) error {
	status, err := t.GetStatus(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	var parent *masta.Status

	if status.InReplyToID != nil {
		parent, err = t.GetStatus(t.Ctx, *status.InReplyToID)
		if err != nil {
			return err
		}
	}

	return render.QuickReplyPage(t.Rctx, status, parent)
}

func init() { reg(handleLikedBy, http.MethodGet, "/likedby/:id") }
func handleLikedBy(t *Transaction) error {
	accts, err := t.GetFavouritedBy(t.Ctx, t.Vars["id"], nil)
	if err != nil {
		return err
	}

	return render.LikedByPage(t.Rctx, accts)
}

func init() { reg(handleRetweetedBy, http.MethodGet, "/retweetedby/:id") }
func handleRetweetedBy(t *Transaction) error {
	accts, err := t.GetRebloggedBy(t.Ctx, t.Vars["id"], nil)
	if err != nil {
		return err
	}

	return render.RetweetedByPage(t.Rctx, accts)
}

func init() { reg(handleReactions, http.MethodGet, "/reactions/:id") }
func handleReactions(t *Transaction) error {
	reactions, err := t.PlGetReactions(t.Ctx, t.Vars["id"], false)
	if err != nil {
		return err
	}

	return render.ReactionsPage(t.Rctx, reactions)
}

func init() { reg(handleEdits, http.MethodGet, "/status/:id/edits") }
func handleEdits(t *Transaction) error {
	edits, err := t.GetStatusHistory(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	current, err := t.GetStatus(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	return render.EditsPage(t.Rctx, edits, current)
}

func init() { reg(handleNotifications, http.MethodGet, "/notifications") }
func handleNotifications(t *Transaction) error {
	q := t.R.URL.Query()
	pg := masta.Pagination{
		MaxID: q.Get("max_id"),
		MinID: q.Get("min_id"),
		Limit: conf.MaxPagination,
	}

	var filter masta.NotificationFilter
	if t.Session.Settings.HideUnsupportedNotifs {
		// Explicitly include the supported types.
		// For now, only Pleroma supports this option, Mastadon
		// will simply ignore the unknown params.
		filter.Include = []string{"follow", "follow_request", "mention", "reblog", "favourite", "pleroma:emoji_reaction"}
	}

	if t.Session.Settings.AntiDopamineMode {
		filter.Exclude = []string{"follow", "favourite", "reblog"}
	}

	notifs, err := t.GetNotificationsOf(t.Ctx, filter, &pg)
	if err != nil {
		return err
	}

	return render.NotificationPage(t.Rctx, notifs)
}

func init() { reg(handleUser, http.MethodGet, "/user/:id") }
func init() { reg(handleUser, http.MethodGet, "/user/:id/:type") }
func handleUser(t *Transaction) error {
	var statuses []*masta.Status
	var users []*masta.Account

	id := t.Vars["id"]
	pageType := t.Vars["type"]

	maxID := t.Qry["max_id"]
	minID := t.Qry["min_id"]

	pg := masta.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: conf.MaxPagination,
	}

	var isAccounts bool

	acct, rel, err := t.GetAccountWithRelationship(t.Ctx, id)
	if err != nil {
		return err
	}

	rPageType := render.UserPageStatuses

	selected := true
	switch pageType {
	case "":
		statuses, err = t.GetAcctStatuses(t.Ctx, id, masta.AcctStatusOpts{
			Pagination: &pg,
		})
		if err != nil {
			return err
		}
	case "following":
		rPageType = render.UserPageFollowing
		users, err = t.GetAccountFollowing(t.Ctx, id, &pg)
		if err != nil {
			return err
		}
		isAccounts = true
	case "followers":
		rPageType = render.UserPageFollowers
		users, err = t.GetAccountFollowers(t.Ctx, id, &pg)
		if err != nil {
			return err
		}
		isAccounts = true
	case "pinned":
		rPageType = render.UserPagePinned
		statuses, err = t.GetAccountPinnedStatuses(t.Ctx, id)
		if err != nil {
			return err
		}
	case "media":
		rPageType = render.UserPageMedia
		statuses, err = t.GetAcctStatuses(t.Ctx, id, masta.AcctStatusOpts{
			OnlyMedia:  true,
			Pagination: &pg,
		})
		if err != nil {
			return err
		}
	default:
		selected = false
	}

	if !selected && t.Session.UserID != rel.ID {
		return errInvalidArgument
	} else if !selected {
		switch pageType {
		case "bookmarks":
			rPageType = render.UserPageBookmarks
			statuses, err = t.GetBookmarks(t.Ctx, &pg)
			if err != nil {
				return err
			}
		case "mutes":
			rPageType = render.UserPageMutes
			users, err = t.GetMutes(t.Ctx, &pg)
			if err != nil {
				return err
			}
			isAccounts = true

		case "blocks":
			rPageType = render.UserPageBlocks
			users, err = t.GetBlocks(t.Ctx, &pg)
			if err != nil {
				return err
			}
			isAccounts = true
		case "likes":
			rPageType = render.UserPageLikes
			statuses, err = t.GetFavourites(t.Ctx, &pg)
			if err != nil {
				return err
			}
		case "requests":
			rPageType = render.UserPageRequests
			users, err = t.GetFollowRequests(t.Ctx, &pg)
			if err != nil {
				return err
			}
			isAccounts = true
		default:
			return errInvalidArgument
		}
	}

	if rPageType != render.UserPagePinned {
		t.Rctx.Pagination = &pg
	}

	if isAccounts {
		return render.UserPage(t.Rctx, acct, rel, users, rPageType)
	}

	return render.UserPage(t.Rctx, acct, rel, statuses, rPageType)
}

func init() { reg(handleUserSearch, http.MethodGet, "/usersearch/:id") }
func handleUserSearch(t *Transaction) error {
	id := t.Vars["id"]
	q := t.R.URL.Query()
	sq := q.Get("q")
	offset, _ := strconv.Atoi(q.Get("offset"))

	user, err := t.GetAccount(t.Ctx, id)
	if err != nil {
		return err
	}

	var results *masta.Results
	if len(q) > 0 {
		results, err = t.DoSearch(t.Ctx, sq,
			masta.SearchOpts{
				Type:      "statuses",
				Resolve:   true,
				Offset:    offset,
				AccountID: id,
				Pagination: &masta.Pagination{
					Limit: conf.MaxPagination,
				},
			},
		)
		if err != nil {
			return err
		}
	} else {
		results = &masta.Results{}
	}

	return render.UserSearchPage(t.Rctx, offset, results, user, sq)
}

func init() { reg(handleAbout, http.MethodGet, "/about") }
func handleAbout(t *Transaction) error {
	return render.AboutPage(t.Rctx)
}

func init() { reg(handleEmojis, http.MethodGet, "/emojis") }
func handleEmojis(t *Transaction) error {
	emojis, err := t.GetInstanceEmojis(t.Ctx)
	if err != nil {
		return err
	}

	return render.EmojiPage(t.Rctx, emojis)
}

func init() { reg(handleSearch, http.MethodGet, "/search") }
func handleSearch(t *Transaction) error {
	q := t.R.URL.Query()
	sq := q.Get("q")
	qType := q.Get("type")
	offset, _ := strconv.Atoi(q.Get("offset"))

	var results *masta.Results
	if len(q) > 0 {
		var err error
		results, err = t.DoSearch(t.Ctx, sq,
			masta.SearchOpts{
				Type:      qType,
				Resolve:   true,
				Offset:    offset,
				Following: false,
				Pagination: &masta.Pagination{
					Limit: 20,
				},
			},
		)
		if err != nil {
			return err
		}
	} else {
		results = &masta.Results{}
	}

	return render.SearchPage(t.Rctx, results, sq, qType, offset)
}

func init() { reg(handleSettings, http.MethodGet, "/settings") }
func handleSettings(t *Transaction) error {
	return render.SettingsPage(t.Rctx)
}

func init() { reg(handleFilters, http.MethodGet, "/filters") }
func handleFilters(t *Transaction) error {
	filters, err := t.GetFilters(t.Ctx)
	if err != nil {
		return err
	}

	return render.FiltersPage(t.Rctx, filters)
}

func init() { reg(handleSigninPost, http.MethodPost, "/signin", noAuth, noCSRF) }
func handleSigninPost(t *Transaction) error {
	instance := t.R.FormValue("instance")

	url, sess, err := newSession(t, instance)
	if err != nil {
		return err
	}

	t.setSession(sess)
	t.redirect(url)

	return nil
}

func init() { reg(handleOAuthCallback, http.MethodGet, "/oauth_callback", noAuth, noCSRF) }
func handleOAuthCallback(t *Transaction) error {
	code := t.Qry["code"]

	if len(code) < 1 {
		return errInvalidArgument
	}

	t.Client = newMastaClient(&masta.Config{
		Server:       "https://" + t.Session.Instance,
		ClientID:     t.Session.ClientID,
		ClientSecret: t.Session.ClientSecret,
	})

	err := t.AuthenticateToken(t.Ctx, code, t.Conf.ClientWebsite+"/oauth_callback")
	if err != nil {
		return err
	}

	u, err := t.GetAccountCurrentUser(t.Ctx)
	if err != nil {
		return err
	}

	t.Session.UserID = u.ID
	t.Session.AccessToken = t.Client.Config.AccessToken

	err = t.setSession(t.Session)
	if err != nil {
		return err
	}

	t.redirect("/")
	return nil
}

func init() { reg(handlePost, http.MethodPost, "/post") }
func handlePost(t *Transaction) error {
	content := t.R.FormValue("content")
	replyToID := t.R.FormValue("reply_to_id")
	format := t.R.FormValue("format")
	visibility := t.R.FormValue("visibility")
	subjectHeader := t.R.FormValue("subject")
	isNSFW := t.R.FormValue("is_nsfw") == "true"
	quickReply := t.R.FormValue("quickreply") == "true"
	files := t.R.MultipartForm.File["attachments"]

	var mediaIDs []string
	for _, f := range files {
		var reader io.Reader
		reader, err := f.Open()
		if err != nil {
			return err
		}
		a, err := t.UploadMediaFromReader(t.Ctx, reader)
		if err != nil {
			return err
		}
		mediaIDs = append(mediaIDs, a.ID)
	}

	tweet := &masta.Toot{
		Status:      content,
		InReplyToID: replyToID,
		MediaIDs:    mediaIDs,
		ContentType: format,
		Visibility:  visibility,
		SpoilerText: subjectHeader,
		Sensitive:   isNSFW,
	}

	st, err := t.PostStatus(t.Ctx, tweet)
	if err != nil {
		return err
	}

	var location string
	if len(replyToID) > 0 {
		if quickReply {
			location = "/quickreply/" + st.ID + "#status-" + st.ID
		} else {
			location = "/thread/" + replyToID + "#status-" + st.ID
		}
	} else {
		location = t.R.FormValue("referrer")
	}
	t.redirect(location)
	return nil
}

func init() { reg(handleEdit, http.MethodPost, "/edit") }
func handleEdit(t *Transaction) error {
	originalID := t.R.FormValue("id")
	content := t.R.FormValue("content")
	replyToID := t.R.FormValue("reply_to_id")
	format := t.R.FormValue("format")
	visibility := t.R.FormValue("visibility")
	subjectHeader := t.R.FormValue("subject")
	isNSFW := t.R.FormValue("is_nsfw") == "true"
	files := t.R.MultipartForm.File["attachments"]
	alt := t.R.Form["alt_text"]
	mediaIDs := t.R.Form["media_ids"]

	var editedAttachments []masta.MediaAttribute
	if len(files) != 0 {
		mediaIDs = []string{}
		for _, f := range files {
			var reader io.Reader
			reader, err := f.Open()
			if err != nil {
				return err
			}
			a, err := t.UploadMediaFromReader(t.Ctx, reader)
			if err != nil {
				return err
			}
			mediaIDs = append(mediaIDs, a.ID)
		}
	} else if len(alt) <= len(mediaIDs) {
		for i, v := range alt {
			v := v
			editedAttachments = append(editedAttachments, masta.MediaAttribute{
				ID:          mediaIDs[i],
				Description: &v,
			})
		}
	}

	tweet := &masta.Toot{
		Status:              content,
		InReplyToID:         replyToID,
		MediaIDs:            mediaIDs,
		EditMediaAttributes: editedAttachments,
		ContentType:         format,
		Visibility:          visibility,
		SpoilerText:         subjectHeader,
		Sensitive:           isNSFW,
	}

	_, err := t.CompatUpdateStatus(t.Ctx, tweet, originalID)
	if err != nil {
		return err
	}

	t.redirect("/thread/" + originalID + "#post-" + originalID)
	return nil
}

func init() { reg(handleLike, http.MethodPost, "/like/:id") }
func handleLike(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Favourite(t.Ctx, id)
	if err != nil {
		return err
	}
	if len(rid) > 0 {
		id = rid
	}
	t.redirect(t.R.FormValue("referrer") + "#status-" + id)
	return nil
}

func init() { reg(handleUnlike, http.MethodPost, "/unlike/:id") }
func handleUnlike(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Unfavourite(t.Ctx, id)
	if err != nil {
		return err
	}
	if len(rid) > 0 {
		id = rid
	}
	t.redirect(t.R.FormValue("referrer") + "#status-" + id)
	return nil
}

func init() { reg(handleRetweet, http.MethodPost, "/retweet/:id") }
func handleRetweet(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Reblog(t.Ctx, id)
	if err != nil {
		return err
	}

	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)
	return nil
}

func init() { reg(handleUnretweet, http.MethodPost, "/unretweet/:id") }
func handleUnretweet(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Unreblog(t.Ctx, id)
	if err != nil {
		return err
	}

	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)

	return nil
}

func init() { reg(handleVote, http.MethodPost, "/vote/:id") }
func handleVote(t *Transaction) error {
	statusID := t.R.FormValue("status_id")
	choices := t.R.PostForm["choices"]

	var err error

	convchoice := make([]int, len(choices))
	for i, v := range choices {
		convchoice[i], err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	}
	_, err = t.PollVote(t.Ctx, t.Vars["id"], convchoice...)
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + statusID)
	return nil
}

func init() { reg(handleFollow, http.MethodPost, "/follow/:id") }
func handleFollow(t *Transaction) error {
	_, err := t.AccountFollow(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleUnfollow, http.MethodPost, "/unfollow/:id") }
func handleUnfollow(t *Transaction) error {
	_, err := t.AccountUnfollow(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleAccept, http.MethodPost, "/accept/:id") }
func handleAccept(t *Transaction) error {
	err := t.FollowRequestAuthorize(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleReject, http.MethodPost, "/reject/:id") }
func handleReject(t *Transaction) error {
	err := t.FollowRequestReject(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleMuteGet, http.MethodGet, "/mute/:id") }
func handleMuteGet(t *Transaction) error {
	acct, err := t.GetAccount(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	return render.MutePage(t.Rctx, acct)
}

func init() { reg(handleMutePost, http.MethodPost, "/mute/:id") }
func handleMutePost(t *Transaction) error {
	notifs, _ := strconv.ParseBool(t.R.FormValue("notifications"))

	duration, err := strconv.ParseInt(t.R.FormValue("duration"), 10, 64)
	if err != nil {
		return err
	}

	_, err = t.AccountMuteWith(t.Ctx, t.Vars["id"], masta.AccountMuteOpts{
		Duration:      duration,
		Notifications: notifs,
	},
	)
	if err != nil {
		return err
	}

	t.redirect("/user/" + t.Vars["id"])
	return nil
}

func init() { reg(handleUnmute, http.MethodPost, "/unmute/:id") }
func handleUnmute(t *Transaction) error {
	_, err := t.AccountUnmute(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleBlock, http.MethodPost, "/block/:id") }
func handleBlock(t *Transaction) error {
	_, err := t.AccountBlock(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleUnblock, http.MethodPost, "/unblock/:id") }
func handleUnblock(t *Transaction) error {
	_, err := t.AccountUnblock(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleSubscribe, http.MethodPost, "/subscribe/:id") }
func handleSubscribe(t *Transaction) error {
	_, err := t.PlAccountSubscribe(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleUnsubscribe, http.MethodPost, "/unsubscribe/:id") }
func handleUnsubscribe(t *Transaction) error {
	_, err := t.PlAccountUnsubscribe(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleRemoveFollower, http.MethodPost, "/removefollower/:id") }
func handleRemoveFollower(t *Transaction) error {
	_, err := t.AccountRemoveFollower(t.Ctx, t.Vars["id"])
	if err != nil {
		if aerr := err.(*masta.APIError); aerr != nil {
			// Pleroma might sometimes not let you remove a follower if it's mandated by the instance.
			if aerr.Code == 404 {
				t.redirect(t.R.FormValue("referrer"))
				return nil
			}
		}
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleSetSettings, http.MethodPost, "/settings") }
func handleSetSettings(t *Transaction) error {
	visibility := t.R.FormValue("visibility")
	format := t.R.FormValue("format")
	copyScope := t.R.FormValue("copy_scope") == "true"
	threadInNewTab := t.R.FormValue("thread_in_new_tab") == "true"
	hideAttachments := t.R.FormValue("hide_attachments") == "true"
	maskNSFW := t.R.FormValue("mask_nsfw") == "true"
	ni, _ := strconv.Atoi(t.R.FormValue("notification_interval"))
	fluorideMode := t.R.FormValue("fluoride_mode") == "true"
	darkMode := t.R.FormValue("dark_mode") == "true"
	antiDopamineMode := t.R.FormValue("anti_dopamine_mode") == "true"
	hideUnsupportedNotifs := t.R.FormValue("hide_unsupported_notifs") == "true"
	css := t.R.FormValue("css")

	settings := &render.Settings{
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
		Stamp:                 conf.ID(),
	}

	switch settings.NotificationInterval {
	case 0, 30, 60, 120, 300, 600:
	default:
		return errInvalidArgument
	}

	if len(settings.CSS) > 1<<20 {
		return errInvalidArgument
	}

	t.Session.Settings = *settings
	err := t.setSession(t.Session)
	if err != nil {
		return err
	}

	t.redirect("/")
	return nil
}

func init() { reg(handleMuteConversation, http.MethodPost, "/muteconv/:id") }
func handleMuteConversation(t *Transaction) error {
	_, err := t.MuteConversation(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleUnmuteConversation, http.MethodPost, "/unmuteconv/:id") }
func handleUnmuteConversation(t *Transaction) error {
	_, err := t.UnmuteConversation(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleDelete, http.MethodPost, "/delete/:id") }
func handleDelete(t *Transaction) error {
	err := t.DeleteStatus(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}
	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleReadNotifications, http.MethodPost, "/notifications/read") }
func handleReadNotifications(t *Transaction) error {
	err := t.PlReadNotificationsTo(t.Ctx, t.Qry["max_id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleBookmark, http.MethodPost, "/bookmark/:id") }
func handleBookmark(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Bookmark(t.Ctx, id)
	if err != nil {
		return err
	}

	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)
	return nil
}

func init() { reg(handleUnbookmark, http.MethodPost, "/unbookmark/:id") }
func handleUnbookmark(t *Transaction) error {
	id := t.Vars["id"]
	rid := t.R.FormValue("retweeted_by_id")

	_, err := t.Unbookmark(t.Ctx, id)
	if err != nil {
		return err
	}

	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)
	return nil
}

func init() { reg(handleProfilePage, http.MethodGet, "/profile") }
func init() { reg(handleProfilePage, http.MethodPost, "/profile") }
func handleProfilePage(t *Transaction) error {
	if t.R.Method == "GET" {
		acct, err := t.GetAccountCurrentUser(t.Ctx)
		if err != nil {
			return err
		}

		return render.ProfilePage(t.Rctx, acct)
	}

	name := t.R.FormValue("name")
	bio := t.R.FormValue("bio")
	bot := t.R.FormValue("bot") == "true"

	var fields []masta.Field
	for i := 0; i < 16; i++ {
		if t.R.FormValue(fmt.Sprintf("field-delete-%d", i)) == "true" {
			continue
		}
		k := t.R.FormValue(fmt.Sprintf("field-key-%d", i))
		if len(k) == 0 {
			continue
		}
		v := t.R.FormValue(fmt.Sprintf("field-value-%d", i))
		f := masta.Field{Name: k, Value: v}
		fields = append(fields, f)
	}

	newk := t.R.FormValue("field-new-key")
	if len(newk) != 0 {
		newv := t.R.FormValue("field-new-value")
		fields = append(fields, masta.Field{Name: newk, Value: newv})
	}

	locked := t.R.FormValue("locked") == "true"

	tertiary := func(key string) *bool {
		if val := t.R.FormValue(key); val == "ignore" {
			return nil
		} else {
			t := val == "true"
			return &t
		}
	}

	indexable := tertiary("noindex")
	if indexable != nil {
		*indexable = !*indexable
	}

	profile := &masta.Profile{
		DisplayName:          &name,
		Note:                 &bio,
		Fields:               &fields,
		Locked:               &locked,
		Bot:                  &bot,
		Indexable:            indexable,
		Discoverable:         tertiary("discoverable"),
		HideCollections:      tertiary("hide-collections"),
		PlHideFavorites:      tertiary("hide-favourites"),
		PlHideFollowers:      tertiary("hide-followers"),
		PlHideFollowersCount: tertiary("hide-followers-count"),
		PlHideFollows:        tertiary("hide-follows"),
		PlHideFollowsCount:   tertiary("hide-follows-count"),
	}

	if t.R.FormValue("profile-img-delete") == "true" {
		profile.Avatar = masta.EmptyFile()
	} else if f := t.R.MultipartForm.File["avatar"]; len(f) > 0 {
		profile.Avatar = &masta.File{}
		profile.Avatar.Name = f[0].Filename

		c, err := f[0].Open()
		if err != nil {
			return err
		}

		defer c.Close()

		profile.Avatar.Content = c
	}
	if t.R.FormValue("profile-banner-delete") == "true" {
		profile.Header = masta.EmptyFile()
	} else if f := t.R.MultipartForm.File["banner"]; len(f) > 0 {
		profile.Header = &masta.File{}
		profile.Header.Name = f[0].Filename

		c, err := f[0].Open()
		if err != nil {
			return err
		}

		defer c.Close()

		profile.Header.Content = c
	}

	_, err := t.AccountUpdate(t.Ctx, profile)
	if err != nil {
		return err
	}

	t.redirect("/profile")
	return nil
}

func init() { reg(handlePin, http.MethodPost, "/pin/:id") }
func handlePin(t *Transaction) error {
	id := t.Vars["id"]

	_, err := t.Pin(t.Ctx, id)
	if err != nil {
		return err
	}

	rid := t.R.FormValue("retweeted_by_id")
	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)

	return nil
}

func init() { reg(handleUnpin, http.MethodPost, "/unpin/:id") }
func handleUnpin(t *Transaction) error {
	id := t.Vars["id"]

	_, err := t.Unpin(t.Ctx, id)
	if err != nil {
		return err
	}

	rid := t.R.FormValue("retweeted_by_id")
	if len(rid) > 0 {
		id = rid
	}

	t.redirect(t.R.FormValue("referrer") + "#status-" + id)

	return nil
}

func init() { reg(handleFilter, http.MethodPost, "/filter") }
func handleFilter(t *Transaction) error {
	filter := &masta.Filter{
		Context:   []string{"home", "notifications", "public", "thread"},
		Phrase:    t.R.FormValue("phrase"),
		WholeWord: t.R.FormValue("whole_word") == "true",
	}

	_, err := t.CreateFilter(t.Ctx, filter)
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleUnfilter, http.MethodPost, "/unfilter/:id") }
func handleUnfilter(t *Transaction) error {
	err := t.DeleteFilter(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleLists, http.MethodGet, "/lists") }
func handleLists(t *Transaction) error {
	lists, err := t.GetLists(t.Ctx)
	if err != nil {
		return err
	}

	return render.ListsPage(t.Rctx, lists)
}

func init() { reg(handleAddList, http.MethodPost, "/list") }
func handleAddList(t *Transaction) error {
	title := t.R.FormValue("title")

	_, err := t.CreateList(t.Ctx, title)
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleRemoveList, http.MethodPost, "/list/:id/remove") }
func handleRemoveList(t *Transaction) error {
	err := t.DeleteList(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}
	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleRenameList, http.MethodPost, "/list/:id/rename") }
func handleRenameList(t *Transaction) error {
	title := t.R.FormValue("title")

	_, err := t.RenameList(t.Ctx, t.Vars["id"], title)
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleList, http.MethodGet, "/list/:id") }
func handleList(t *Transaction) error {
	id := t.Vars["id"]
	q := t.Qry["q"]

	list, err := t.GetList(t.Ctx, id)
	if err != nil {
		return err
	}
	accounts, err := t.GetListAccounts(t.Ctx, id)
	if err != nil {
		return err
	}

	data := &render.ListData{
		List:     list,
		Accounts: accounts,
		Q:        q,
	}

	if len(q) > 0 {
		data.SearchAccounts = []*masta.Account{}

		// Do it ourselves, since Mastodon doesn't support filtering searches down
		// to followers.
		following, err := t.GetAccountFollowing(t.Ctx, t.Session.UserID, nil)
		if err != nil {
			return err
		}

		lowq := strings.ToLower(q)
		for _, v := range following {
			skip := false

			for _, j := range accounts {
				if v.ID == j.ID {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			if v.Acct == q || (q[0] == '@' && strings.HasPrefix(v.Acct, q[1:])) {
				data.SearchAccounts = append([]*masta.Account{v}, data.Accounts...)
				continue
			}

			lowacct := strings.ToLower(v.Acct)
			if strings.Contains(lowacct, lowq) || strings.Contains(v.DisplayName, lowq) {
				data.SearchAccounts = append(data.SearchAccounts, v)
				continue
			}
		}
	}

	return render.ListPage(t.Rctx, data)
}

func init() { reg(handleListAddUser, http.MethodPost, "/list/:id/adduser") }
func handleListAddUser(t *Transaction) error {
	err := t.AddToList(t.Ctx, t.Vars["id"], t.Qry["uid"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))
	return nil
}

func init() { reg(handleListRemoveUser, http.MethodPost, "/list/:id/removeuser") }
func handleListRemoveUser(t *Transaction) error {
	err := t.RemoveFromList(t.Ctx, t.Vars["id"], t.Qry["uid"])
	if err != nil {
		return err
	}

	t.redirect(t.R.FormValue("referrer"))

	return nil
}

func init() { reg(handleSignout, http.MethodPost, "/signout", noType) }
func handleSignout(t *Transaction) error {
	t.unsetSession()
	t.redirect("/")
	return nil
}

func init() { reg(handleFluorideLike, http.MethodPost, "/fluoride/like/:id", noType) }
func handleFluorideLike(t *Transaction) error {
	t.W.Header().Set("Content-Type", "application/json")

	st, err := t.Favourite(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	return t.writeJson(st.FavouritesCount)
}

func init() { reg(handleFluorideRetweet, http.MethodPost, "/fluoride/retweet/:id", noType) }
func handleFluorideRetweet(t *Transaction) error {
	t.W.Header().Set("Content-Type", "application/json")
	st, err := t.Reblog(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	return t.writeJson(st.ReblogsCount)
}

func init() { reg(handleFluorideUnretweet, http.MethodPost, "/fluoride/unretweet/:id", noType) }
func handleFluorideUnretweet(t *Transaction) error {
	t.W.Header().Set("Content-Type", "application/json")
	count, err := t.Unreblog(t.Ctx, t.Vars["id"])
	if err != nil {
		return err
	}

	return t.writeJson(count)
}

func init() { reg(handleUserCSS, http.MethodGet, "/session/css", noType) }
func handleUserCSS(t *Transaction) error {
	stamp := t.Qry["stamp"]
	if stamp != "" && stamp != t.Session.Settings.Stamp {
		t.W.WriteHeader(http.StatusNotFound)
		return nil
	}

	if stamp != "" {
		t.W.Header().Set("Cache-Control", "public, immutable, max-age=31556952, stale-while-revalidate=31556952")
	}

	t.W.Header().Set("Content-Type", "text/css")
	t.W.Write([]byte(t.Session.Settings.CSS))
	return nil
}

//go:embed static/*
var embedfs embed.FS

var assetfs = &staticfs{
	underlying: embedfs,
}

var fserve = http.FileServer(http.FS(assetfs))

func init() { reg(handleStatic, http.MethodGet, "/static/:asset", noType, noAuth) }
func handleStatic(t *Transaction) error {
	stamp := t.Qry["stamp"]
	if stamp != "" && stamp == t.Conf.AssetStamp {
		t.W.Header().Set("Cache-Control", "public, immutable, max-age=31556952, stale-while-revalidate=31556952")
	}

	fserve.ServeHTTP(t.W, t.R)
	return nil
}
