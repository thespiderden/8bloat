package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"bloat/model"
	"bloat/renderer"
	"bloat/util"

	"spiderden.org/masta"

	ua "github.com/mileusna/useragent"
)

var (
	errInvalidArgument  = errors.New("invalid argument")
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

type service struct {
	cname       string
	cscope      string
	cwebsite    string
	instance    string
	postFormats []model.PostFormat
	renderer    renderer.Renderer
}

func NewService(cname string, cscope string, cwebsite string,
	instance string, postFormats []model.PostFormat,
	renderer renderer.Renderer,
) *service {
	return &service{
		cname:       cname,
		cscope:      cscope,
		cwebsite:    cwebsite,
		instance:    instance,
		postFormats: postFormats,
		renderer:    renderer,
	}
}

func (s *service) cdata(c *client, title string, count int, rinterval int,
	target string,
) (data *renderer.CommonData) {
	data = &renderer.CommonData{
		Title:           title + " - " + s.cname,
		Count:           count,
		RefreshInterval: rinterval,
		Target:          target,
	}
	if c != nil && c.s.IsLoggedIn() {
		data.CSRFToken = c.s.CSRFToken
	}
	return
}

func (s *service) ErrorPage(c *client, err error, retry bool) error {
	var errStr string
	var sessionErr bool
	if err != nil {
		errStr = err.Error()
		if me, ok := err.(*masta.APIError); ok {
			switch me.Code {
			case http.StatusForbidden, http.StatusUnauthorized:
				sessionErr = true
			}
		}
	}
	cdata := s.cdata(nil, "error", 0, 0, "")
	data := &renderer.ErrorData{
		CommonData: cdata,
		Err:        errStr,
		Retry:      retry,
		SessionErr: sessionErr,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.ErrorPage, data)
}

func (s *service) SigninPage(c *client) (err error) {
	cdata := s.cdata(nil, "signin", 0, 0, "")
	data := &renderer.SigninData{
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.SigninPage, data)
}

func (s *service) RootPage(c *client) (err error) {
	data := &renderer.RootData{
		Title: s.cname,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.RootPage, data)
}

func (s *service) NavPage(c *client) (err error) {
	u, err := c.GetAccountCurrentUser(c.ctx)
	if err != nil {
		return
	}
	pctx := model.PostContext{
		DefaultVisibility: c.s.Settings.DefaultVisibility,
		DefaultFormat:     c.s.Settings.DefaultFormat,
		Formats:           s.postFormats,
		UserAgent:         ua.Parse(c.r.UserAgent()),
	}
	cdata := s.cdata(c, "nav", 0, 0, "main")
	data := &renderer.NavData{
		User:        u,
		CommonData:  cdata,
		PostContext: pctx,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.NavPage, data)
}

func (s *service) TimelinePage(c *client, tType, instance, listId, maxID,
	minID string,
) (err error) {
	var nextLink, prevLink, title string
	var statuses []*masta.Status
	pg := masta.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	switch tType {
	default:
		return errInvalidArgument
	case "home":
		statuses, err = c.GetTimelineHome(c.ctx, &pg)
		if err != nil {
			return err
		}
		title = "Timeline"
	case "direct":
		statuses, err = c.GetTimelineDirect(c.ctx, &pg)
		if err != nil {
			return err
		}
		title = "Direct Timeline"
	case "local":
		statuses, err = c.GetTimelinePublic(c.ctx, true, &pg)
		if err != nil {
			return err
		}
		title = "Local Timeline"
	case "remote":
		if len(instance) > 0 {
			statuses, err = c.PlGetTimelineRemote(c.ctx, instance, &pg)
			if err != nil {
				return err
			}
		}
		title = "Remote Timeline"
	case "twkn":
		statuses, err = c.GetTimelinePublic(c.ctx, false, &pg)
		if err != nil {
			return err
		}
		title = "The Whole Known Network"
	case "list":
		statuses, err = c.GetTimelineList(c.ctx, listId, &pg)
		if err != nil {
			return err
		}
		list, err := c.GetList(c.ctx, listId)
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
		if len(listId) > 0 {
			v.Set("list", listId)
		}
		prevLink = "/timeline/" + tType + "?" + v.Encode()
	}

	if len(minID) > 0 || (len(pg.MaxID) > 0 && len(statuses) == 20) {
		v := make(url.Values)
		v.Set("max_id", pg.MaxID)
		if len(instance) > 0 {
			v.Set("instance", instance)
		}
		if len(listId) > 0 {
			v.Set("list", listId)
		}
		nextLink = "/timeline/" + tType + "?" + v.Encode()
	}

	cdata := s.cdata(c, tType+" timeline ", 0, 0, "")
	data := &renderer.TimelineData{
		Title:      title,
		Type:       tType,
		Instance:   instance,
		Statuses:   statuses,
		NextLink:   nextLink,
		PrevLink:   prevLink,
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.TimelinePage, data)
}

func (s *service) ListsPage(c *client) (err error) {
	lists, err := c.GetLists(c.ctx)
	if err != nil {
		return
	}

	cdata := s.cdata(c, "Lists", 0, 0, "")
	data := renderer.ListsData{
		Lists:      lists,
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.ListsPage, data)
}

func (s *service) AddList(c *client, title string) (err error) {
	_, err = c.CreateList(c.ctx, title)
	return err
}

func (s *service) RemoveList(c *client, id string) (err error) {
	return c.DeleteList(c.ctx, id)
}

func (s *service) RenameList(c *client, id, title string) (err error) {
	_, err = c.RenameList(c.ctx, id, title)
	return err
}

func (s *service) ListPage(c *client, id string, q string) (err error) {
	list, err := c.GetList(c.ctx, id)
	if err != nil {
		return
	}
	accounts, err := c.GetListAccounts(c.ctx, id)
	if err != nil {
		return
	}
	var searchAccounts []*masta.Account
	if len(q) > 0 {
		result, err := c.DoSearch(c.ctx, q, masta.SearchOpts{
			Type:      "accounts",
			Resolve:   true,
			AccountID: id,
			Following: true,
			Pagination: &masta.Pagination{
				Limit: 20,
			},
		})
		if err != nil {
			return err
		}
		searchAccounts = result.Accounts
	}
	cdata := s.cdata(c, "List "+list.Title, 0, 0, "")
	data := renderer.ListData{
		List:           list,
		Accounts:       accounts,
		Q:              q,
		SearchAccounts: searchAccounts,
		CommonData:     cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.ListPage, data)
}

func (s *service) ListAddUser(c *client, id string, uid string) (err error) {
	return c.AddToList(c.ctx, id, uid)
}

func (s *service) ListRemoveUser(c *client, id string, uid string) (err error) {
	return c.RemoveFromList(c.ctx, id, uid)
}

func (s *service) ThreadPage(c *client, id string, reply bool) (err error) {
	var pctx model.PostContext

	status, err := c.GetStatus(c.ctx, id)
	if err != nil {
		return
	}

	if reply {
		var content string
		var visibility string
		if c.s.UserID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for i := range status.Mentions {
			if status.Mentions[i].ID != c.s.UserID &&
				status.Mentions[i].ID != status.Account.ID {
				content += "@" + status.Mentions[i].Acct + " "
			}
		}

		isDirect := status.Visibility == "direct"
		if isDirect || c.s.Settings.CopyScope {
			visibility = status.Visibility
		} else {
			visibility = c.s.Settings.DefaultVisibility
		}

		pctx = model.PostContext{
			DefaultVisibility: visibility,
			DefaultFormat:     c.s.Settings.DefaultFormat,
			Formats:           s.postFormats,
			ReplyContext: &model.ReplyContext{
				InReplyToID:        id,
				InReplyToName:      status.Account.Acct,
				ReplyContent:       content,
				ReplySubjectHeader: status.SpoilerText,
				ForceVisibility:    isDirect,
			},
			UserAgent: ua.Parse(c.r.UserAgent()),
		}
	}

	context, err := c.GetStatusContext(c.ctx, id)
	if err != nil {
		return
	}

	statuses := append(append(context.Ancestors, status), context.Descendants...)
	replymap := make(map[masta.ID][]renderer.ThreadReplyData)
	nomap := make(map[masta.ID]int)

	statusdata := make([]*renderer.StatusData, len(statuses))

	for i, status := range statuses {
		no := i + 1
		nomap[status.ID] = no

		data := renderer.StatusData{
			No:          &no,
			Status:      status,
			Replies:     []renderer.ThreadReplyData{},
			ShowReplies: true,
		}

		statusdata[i] = &data

		if replyee := status.InReplyToID; replyee != nil {
			replyee := *replyee
			replydata := renderer.ThreadReplyData{
				No: no,
				ID: status.ID,
			}

			_, ok := replymap[replyee]
			if !ok {
				replymap[replyee] = []renderer.ThreadReplyData{replydata}
			} else {
				replyarr := replymap[replyee]
				replymap[replyee] = append(replyarr, replydata)
			}
		}

	}

	for i, v := range statusdata {
		v.Replies = replymap[v.Status.ID]

		if id := statuses[i].InReplyToID; id != nil {
			no := nomap[*id]
			v.InReplyToNo = &no
		}
	}

	cdata := s.cdata(c, "post by "+status.Account.DisplayName, 0, 0, "")
	data := &renderer.ThreadData{
		Statuses:    statusdata,
		PostContext: pctx,
		CommonData:  cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.ThreadPage, data)
}

func (s *service) QuickReplyPage(c *client, id string) (err error) {
	status, err := c.GetStatus(c.ctx, id)
	if err != nil {
		return
	}

	var ancestor *masta.Status
	if status.InReplyToID != nil {
		ancestor, err = c.GetStatus(c.ctx, *status.InReplyToID)
		if err != nil {
			return
		}
	}

	var content string
	if c.s.UserID != status.Account.ID {
		content += "@" + status.Account.Acct + " "
	}
	for i := range status.Mentions {
		if status.Mentions[i].ID != c.s.UserID &&
			status.Mentions[i].ID != status.Account.ID {
			content += "@" + status.Mentions[i].Acct + " "
		}
	}

	var visibility string
	isDirect := status.Visibility == "direct"
	if isDirect || c.s.Settings.CopyScope {
		visibility = status.Visibility
	} else {
		visibility = c.s.Settings.DefaultVisibility
	}

	pctx := model.PostContext{
		DefaultVisibility: visibility,
		DefaultFormat:     c.s.Settings.DefaultFormat,
		Formats:           s.postFormats,
		ReplyContext: &model.ReplyContext{
			InReplyToID:        id,
			InReplyToName:      status.Account.Acct,
			QuickReply:         true,
			ReplyContent:       content,
			ReplySubjectHeader: status.SpoilerText,
			ForceVisibility:    isDirect,
		},
		UserAgent: ua.Parse(c.r.UserAgent()),
	}

	cdata := s.cdata(c, "post by "+status.Account.DisplayName, 0, 0, "")
	data := &renderer.QuickReplyData{
		Ancestor:    ancestor,
		Status:      status,
		PostContext: pctx,
		CommonData:  cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.QuickReplyPage, data)
}

func (s *service) LikedByPage(c *client, id string) (err error) {
	likers, err := c.GetFavouritedBy(c.ctx, id, nil)
	if err != nil {
		return
	}
	cdata := s.cdata(c, "likes", 0, 0, "")
	data := &renderer.LikedByData{
		CommonData: cdata,
		Users:      likers,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.LikedByPage, data)
}

func (s *service) RetweetedByPage(c *client, id string) (err error) {
	retweeters, err := c.GetRebloggedBy(c.ctx, id, nil)
	if err != nil {
		return
	}
	cdata := s.cdata(c, "retweets", 0, 0, "")
	data := &renderer.RetweetedByData{
		CommonData: cdata,
		Users:      retweeters,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.RetweetedByPage, data)
}

func (s *service) ReactionsPage(c *client, id string) (err error) {
	reactions, err := c.PlGetReactions(c.ctx, id, false)
	if err != nil {
		return
	}
	cdata := s.cdata(c, "reactions", 0, 0, "")
	data := &renderer.ReactionsData{
		CommonData: cdata,
		Reactions:  reactions,
	}

	return s.renderer.Render(c.rctx, c.w, renderer.ReactionsPage, data)
}

func (s *service) NotificationPage(c *client, maxID string, minID string) (err error) {
	var nextLink string
	var unreadCount int
	var readID string
	pg := masta.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	var filter masta.NotificationFilter
	if c.s.Settings.HideUnsupportedNotifs {
		// Explicitly include the supported types.
		// For now, only Pleroma supports this option, Mastadon
		// will simply ignore the unknown params.
		filter.Include = []string{"follow", "follow_request", "mention", "reblog", "favourite", "pleroma:emoji_reaction"}
	}
	if c.s.Settings.AntiDopamineMode {
		filter.Exclude = []string{"follow", "favourite", "reblog"}
	}

	notifications, err := c.GetNotificationsOf(c.ctx, filter, &pg)
	if err != nil {
		return
	}

	for i := range notifications {
		if notifications[i].Pleroma != nil && !notifications[i].Pleroma.IsSeen {
			unreadCount++
		}
	}

	if unreadCount > 0 {
		readID = notifications[0].ID
	}
	if len(notifications) == 20 && len(pg.MaxID) > 0 {
		nextLink = "/notifications?max_id=" + pg.MaxID
	}

	cdata := s.cdata(c, "notifications", unreadCount,
		c.s.Settings.NotificationInterval, "main")
	data := &renderer.NotificationData{
		Notifications: notifications,
		UnreadCount:   unreadCount,
		ReadID:        readID,
		NextLink:      nextLink,
		CommonData:    cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.NotificationPage, data)
}

func (s *service) UserPage(c *client, id string, pageType string, maxID string, minID string) (err error) {
	var nextLink string
	var statuses []*masta.Status
	var users []*masta.Account
	pg := masta.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	user, relationship, err := c.GetAccountWithRelationship(c.ctx, id)
	if err != nil {
		return
	}
	isCurrent := c.s.UserID == user.ID

	switch pageType {
	case "":
		statuses, err = c.GetAcctStatuses(c.ctx, id, masta.AcctStatusOpts{
			Pagination: &pg,
		})
		if err != nil {
			return
		}
		if len(statuses) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s?max_id=%s", id,
				pg.MaxID)
		}
	case "following":
		users, err = c.GetAccountFollowing(c.ctx, id, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/following?max_id=%s",
				id, pg.MaxID)
		}
	case "followers":
		users, err = c.GetAccountFollowers(c.ctx, id, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/followers?max_id=%s",
				id, pg.MaxID)
		}
	case "media":
		statuses, err = c.GetAcctStatuses(c.ctx, id, masta.AcctStatusOpts{
			OnlyMedia:  true,
			Pagination: &pg,
		})
		if err != nil {
			return
		}
		if len(statuses) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/media?max_id=%s",
				id, pg.MaxID)
		}
	case "bookmarks":
		if !isCurrent {
			return errInvalidArgument
		}
		statuses, err = c.GetBookmarks(c.ctx, &pg)
		if err != nil {
			return
		}
		if len(statuses) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/bookmarks?max_id=%s",
				id, pg.MaxID)
		}
	case "mutes":
		if !isCurrent {
			return errInvalidArgument
		}
		users, err = c.GetMutes(c.ctx, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/mutes?max_id=%s",
				id, pg.MaxID)
		}
	case "blocks":
		if !isCurrent {
			return errInvalidArgument
		}
		users, err = c.GetBlocks(c.ctx, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/blocks?max_id=%s",
				id, pg.MaxID)
		}
	case "likes":
		if !isCurrent {
			return errInvalidArgument
		}
		statuses, err = c.GetFavourites(c.ctx, &pg)
		if err != nil {
			return
		}
		if len(statuses) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/likes?max_id=%s",
				id, pg.MaxID)
		}
	case "requests":
		if !isCurrent {
			return errInvalidArgument
		}
		users, err = c.GetFollowRequests(c.ctx, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/requests?max_id=%s",
				id, pg.MaxID)
		}
	default:
		return errInvalidArgument
	}

	cdata := s.cdata(c, user.DisplayName+" @"+user.Acct, 0, 0, "")
	data := &renderer.UserData{
		User:         user,
		IsCurrent:    isCurrent,
		Type:         pageType,
		Users:        users,
		Statuses:     statuses,
		NextLink:     nextLink,
		CommonData:   cdata,
		Relationship: relationship,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.UserPage, data)
}

func (s *service) UserSearchPage(c *client, id string, q string, offset int) (err error) {
	var nextLink string
	title := "search"

	user, err := c.GetAccount(c.ctx, id)
	if err != nil {
		return
	}

	var results *masta.Results
	if len(q) > 0 {
		results, err = c.DoSearch(c.ctx, q,
			masta.SearchOpts{
				Type:      "statuses",
				Resolve:   true,
				Offset:    offset,
				AccountID: id,
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

	if len(results.Statuses) == 20 {
		offset += 20
		nextLink = fmt.Sprintf("/usersearch/%s?q=%s&offset=%d", id,
			q, offset)
	}

	if len(q) > 0 {
		title += " \"" + q + "\""
	}

	cdata := s.cdata(c, title, 0, 0, "")
	data := &renderer.UserSearchData{
		CommonData: cdata,
		User:       user,
		Q:          q,
		Statuses:   results.Statuses,
		NextLink:   nextLink,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.UserSearchPage, data)
}

func (s *service) MutePage(c *client, id string) (err error) {
	user, err := c.GetAccount(c.ctx, id)
	if err != nil {
		return
	}
	cdata := s.cdata(c, "Mute"+user.DisplayName+" @"+user.Acct, 0, 0, "")
	data := &renderer.UserData{
		User:       user,
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.MutePage, data)
}

func (s *service) AboutPage(c *client) (err error) {
	cdata := s.cdata(c, "about", 0, 0, "")
	data := &renderer.AboutData{
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.AboutPage, data)
}

func (s *service) EmojiPage(c *client) (err error) {
	emojis, err := c.GetInstanceEmojis(c.ctx)
	if err != nil {
		return
	}
	cdata := s.cdata(c, "emojis", 0, 0, "")
	data := &renderer.EmojiData{
		Emojis:     emojis,
		CommonData: cdata,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.EmojiPage, data)
}

func (s *service) SearchPage(c *client, q string, qType string, offset int) (err error) {
	var nextLink string
	title := "search"

	var results *masta.Results
	if len(q) > 0 {
		results, err = c.DoSearch(c.ctx, q,
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

	if (qType == "accounts" && len(results.Accounts) == 20) ||
		(qType == "statuses" && len(results.Statuses) == 20) {
		offset += 20
		nextLink = fmt.Sprintf("/search?q=%s&type=%s&offset=%d",
			q, qType, offset)
	}

	if len(q) > 0 {
		title += " \"" + q + "\""
	}

	cdata := s.cdata(c, title, 0, 0, "")
	data := &renderer.SearchData{
		CommonData: cdata,
		Q:          q,
		Type:       qType,
		Users:      results.Accounts,
		Statuses:   results.Statuses,
		NextLink:   nextLink,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.SearchPage, data)
}

func (s *service) SettingsPage(c *client) (err error) {
	cdata := s.cdata(c, "settings", 0, 0, "")
	data := &renderer.SettingsData{
		CommonData:  cdata,
		Settings:    &c.s.Settings,
		PostFormats: s.postFormats,
	}
	return s.renderer.Render(c.rctx, c.w, renderer.SettingsPage, data)
}

func (svc *service) FiltersPage(c *client) (err error) {
	filters, err := c.GetFilters(c.ctx)
	if err != nil {
		return
	}
	cdata := svc.cdata(c, "filters", 0, 0, "")
	data := &renderer.FiltersData{
		CommonData: cdata,
		Filters:    filters,
	}
	return svc.renderer.Render(c.rctx, c.w, renderer.FiltersPage, data)
}

func (s *service) SingleInstance() (instance string, ok bool) {
	if len(s.instance) > 0 {
		instance = s.instance
		ok = true
	}
	return
}

func (s *service) NewSession(c *client, instance string) (rurl string, sess *model.Session, err error) {
	var instanceURL string
	if strings.HasPrefix(instance, "https://") {
		instanceURL = instance
		instance = strings.TrimPrefix(instance, "https://")
	} else {
		instanceURL = "https://" + instance
	}

	sid, err := util.NewSessionID()
	if err != nil {
		return
	}
	csrf, err := util.NewCSRFToken()
	if err != nil {
		return
	}

	app, err := masta.RegisterApp(c.ctx, &masta.AppConfig{
		Server:       instanceURL,
		ClientName:   s.cname,
		Scopes:       s.cscope,
		Website:      s.cwebsite,
		RedirectURIs: s.cwebsite + "/oauth_callback",
	})
	if err != nil {
		return
	}
	sess = &model.Session{
		ID:           sid,
		Instance:     instance,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		CSRFToken:    csrf,
		Settings:     *model.NewSettings(),
	}

	u, err := url.Parse("/oauth/authorize")
	if err != nil {
		return
	}

	q := make(url.Values)
	q.Set("scope", "read write follow")
	q.Set("client_id", app.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", s.cwebsite+"/oauth_callback")
	u.RawQuery = q.Encode()

	rurl = instanceURL + u.String()
	return
}

func (s *service) Signin(c *client, code string) (err error) {
	if len(code) < 1 {
		err = errInvalidArgument
		return
	}
	err = c.AuthenticateToken(c.ctx, code, s.cwebsite+"/oauth_callback")
	if err != nil {
		return
	}
	u, err := c.GetAccountCurrentUser(c.ctx)
	if err != nil {
		return
	}

	c.s.AccessToken = c.Config.AccessToken
	c.s.UserID = u.ID
	return c.setSession(c.s)
}

func (s *service) Post(c *client, content string, replyToID string,
	format string, visibility string, subjectHeader string, isNSFW bool,
	files []*multipart.FileHeader,
) (id string, err error) {
	var mediaIDs []string
	for _, f := range files {
		var reader io.Reader
		reader, err = f.Open()
		if err != nil {
			return
		}
		a, err := c.UploadMediaFromReader(c.ctx, reader)
		if err != nil {
			return "", err
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
	st, err := c.PostStatus(c.ctx, tweet)
	if err != nil {
		return
	}
	return st.ID, nil
}

func (s *service) Like(c *client, id string) (count int64, err error) {
	st, err := c.Favourite(c.ctx, id)
	if err != nil {
		return
	}
	count = st.FavouritesCount
	return
}

func (s *service) UnLike(c *client, id string) (count int64, err error) {
	st, err := c.Unfavourite(c.ctx, id)
	if err != nil {
		return
	}
	count = st.FavouritesCount
	return
}

func (s *service) Retweet(c *client, id string) (count int64, err error) {
	st, err := c.Reblog(c.ctx, id)
	if err != nil {
		return
	}
	if st.Reblog != nil {
		count = st.Reblog.ReblogsCount
	}
	return
}

func (s *service) UnRetweet(c *client, id string) (count int64, err error) {
	st, err := c.Unreblog(c.ctx, id)
	if err != nil {
		return
	}
	count = st.ReblogsCount
	return
}

func (s *service) Vote(c *client, id string, choices ...int) error {
	_, err := c.PollVote(c.ctx, id, choices...)
	return err
}

func (s *service) Follow(c *client, id string, reblogs *bool) (err error) {
	_, err = c.AccountFollow(c.ctx, id)
	return
}

func (s *service) UnFollow(c *client, id string) (err error) {
	_, err = c.AccountUnfollow(c.ctx, id)
	return
}

func (s *service) Accept(c *client, id string) (err error) {
	return c.FollowRequestAuthorize(c.ctx, id)
}

func (s *service) Reject(c *client, id string) (err error) {
	return c.FollowRequestReject(c.ctx, id)
}

func (s *service) Mute(c *client, id string, notifications bool, duration int64) (err error) {
	_, err = c.AccountMuteWith(c.ctx, id, masta.AccountMuteOpts{
		Notifications: notifications,
		Duration:      duration,
	},
	)
	return
}

func (s *service) UnMute(c *client, id string) (err error) {
	_, err = c.AccountUnmute(c.ctx, id)
	return
}

func (s *service) Block(c *client, id string) (err error) {
	_, err = c.AccountBlock(c.ctx, id)
	return
}

func (s *service) UnBlock(c *client, id string) (err error) {
	_, err = c.AccountUnblock(c.ctx, id)
	return
}

func (s *service) Subscribe(c *client, id string) (err error) {
	_, err = c.PlAccountSubscribe(c.ctx, id)
	return
}

func (s *service) UnSubscribe(c *client, id string) (err error) {
	_, err = c.PlAccountUnsubscribe(c.ctx, id)
	return
}

func (s *service) SaveSettings(c *client, settings *model.Settings) (err error) {
	switch settings.NotificationInterval {
	case 0, 30, 60, 120, 300, 600:
	default:
		return errInvalidArgument
	}
	if len(settings.CSS) > 1<<20 {
		return errInvalidArgument
	}
	c.s.Settings = *settings
	return c.setSession(c.s)
}

func (s *service) MuteConversation(c *client, id string) (err error) {
	_, err = c.MuteConversation(c.ctx, id)
	return
}

func (s *service) UnMuteConversation(c *client, id string) (err error) {
	_, err = c.UnmuteConversation(c.ctx, id)
	return
}

func (s *service) Delete(c *client, id string) (err error) {
	return c.DeleteStatus(c.ctx, id)
}

func (s *service) ReadNotifications(c *client, maxID string) (err error) {
	return c.PlReadNotificationsTo(c.ctx, maxID)
}

func (s *service) Bookmark(c *client, id string) (err error) {
	_, err = c.Bookmark(c.ctx, id)
	return
}

func (s *service) UnBookmark(c *client, id string) (err error) {
	_, err = c.Unbookmark(c.ctx, id)
	return
}

func (svc *service) Filter(c *client, phrase string, wholeWord bool) (err error) {
	filter := &masta.Filter{
		Context:   []string{"home", "notifications", "public", "thread"},
		Phrase:    phrase,
		WholeWord: wholeWord,
	}
	_, err = c.CreateFilter(c.ctx, filter)
	return
}

func (svc *service) UnFilter(c *client, id string) (err error) {
	return c.DeleteFilter(c.ctx, id)
}
