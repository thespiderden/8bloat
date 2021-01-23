package service

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/url"
	"strings"

	"bloat/mastodon"
	"bloat/model"
	"bloat/renderer"
	"bloat/util"
)

var (
	ctx                = context.Background()
	errInvalidArgument = errors.New("invalid argument")
)

type service struct {
	clientName     string
	clientScope    string
	clientWebsite  string
	customCSS      string
	postFormats    []model.PostFormat
	renderer       renderer.Renderer
	sessionRepo    model.SessionRepo
	appRepo        model.AppRepo
	singleInstance string
}

func NewService(clientName string,
	clientScope string,
	clientWebsite string,
	customCSS string,
	postFormats []model.PostFormat,
	renderer renderer.Renderer,
	sessionRepo model.SessionRepo,
	appRepo model.AppRepo,
	singleInstance string,
) *service {
	return &service{
		clientName:     clientName,
		clientScope:    clientScope,
		clientWebsite:  clientWebsite,
		customCSS:      customCSS,
		postFormats:    postFormats,
		renderer:       renderer,
		sessionRepo:    sessionRepo,
		appRepo:        appRepo,
		singleInstance: singleInstance,
	}
}

func getRendererContext(c *client) *renderer.Context {
	var settings model.Settings
	var session model.Session
	var referrer string
	if c != nil {
		settings = c.Session.Settings
		session = c.Session
		referrer = c.url()
	} else {
		settings = *model.NewSettings()
	}
	return &renderer.Context{
		HideAttachments:  settings.HideAttachments,
		MaskNSFW:         settings.MaskNSFW,
		ThreadInNewTab:   settings.ThreadInNewTab,
		FluorideMode:     settings.FluorideMode,
		DarkMode:         settings.DarkMode,
		CSRFToken:        session.CSRFToken,
		UserID:           session.UserID,
		AntiDopamineMode: settings.AntiDopamineMode,
		Referrer:         referrer,
	}
}

func addToReplyMap(m map[string][]mastodon.ReplyInfo, key interface{},
	val string, number int) {
	if key == nil {
		return
	}
	keyStr, ok := key.(string)
	if !ok {
		return
	}
	_, ok = m[keyStr]
	if !ok {
		m[keyStr] = []mastodon.ReplyInfo{}
	}
	m[keyStr] = append(m[keyStr], mastodon.ReplyInfo{val, number})
}

func (s *service) getCommonData(c *client, title string) (data *renderer.CommonData) {
	data = &renderer.CommonData{
		Title:     title + " - " + s.clientName,
		CustomCSS: s.customCSS,
	}
	if c != nil && c.Session.IsLoggedIn() {
		data.CSRFToken = c.Session.CSRFToken
	}
	return
}

func (s *service) ErrorPage(c *client, err error) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	commonData := s.getCommonData(nil, "error")
	data := &renderer.ErrorData{
		CommonData: commonData,
		Error:      errStr,
	}
	rCtx := getRendererContext(c)
	s.renderer.Render(rCtx, c, renderer.ErrorPage, data)
}

func (s *service) SigninPage(c *client) (err error) {
	commonData := s.getCommonData(nil, "signin")
	data := &renderer.SigninData{
		CommonData: commonData,
	}
	rCtx := getRendererContext(nil)
	return s.renderer.Render(rCtx, c, renderer.SigninPage, data)
}

func (s *service) RootPage(c *client) (err error) {
	data := &renderer.RootData{
		Title: s.clientName,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.RootPage, data)
}

func (s *service) NavPage(c *client) (err error) {
	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return
	}
	postContext := model.PostContext{
		DefaultVisibility: c.Session.Settings.DefaultVisibility,
		DefaultFormat:     c.Session.Settings.DefaultFormat,
		Formats:           s.postFormats,
	}
	commonData := s.getCommonData(c, "nav")
	commonData.Target = "main"
	data := &renderer.NavData{
		User:        u,
		CommonData:  commonData,
		PostContext: postContext,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.NavPage, data)
}

func (s *service) TimelinePage(c *client, tType string, instance string,
	maxID string, minID string) (err error) {

	var nextLink, prevLink, title string
	var statuses []*mastodon.Status
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	switch tType {
	default:
		return errInvalidArgument
	case "home":
		statuses, err = c.GetTimelineHome(ctx, &pg)
		title = "Timeline"
	case "direct":
		statuses, err = c.GetTimelineDirect(ctx, &pg)
		title = "Direct Timeline"
	case "local":
		statuses, err = c.GetTimelinePublic(ctx, true, "", &pg)
		title = "Local Timeline"
	case "remote":
		if len(instance) > 0 {
			statuses, err = c.GetTimelinePublic(ctx, false, instance, &pg)
		}
		title = "Remote Timeline"
	case "twkn":
		statuses, err = c.GetTimelinePublic(ctx, false, "", &pg)
		title = "The Whole Known Network"
	}
	if err != nil {
		return err
	}

	for i := range statuses {
		if statuses[i].Reblog != nil {
			statuses[i].Reblog.RetweetedByID = statuses[i].ID
		}
	}

	if (len(maxID) > 0 || len(minID) > 0) && len(statuses) > 0 {
		v := make(url.Values)
		v.Set("min_id", statuses[0].ID)
		if len(instance) > 0 {
			v.Set("instance", instance)
		}
		prevLink = "/timeline/" + tType + "?" + v.Encode()
	}

	if len(minID) > 0 || (len(pg.MaxID) > 0 && len(statuses) == 20) {
		v := make(url.Values)
		v.Set("max_id", pg.MaxID)
		if len(instance) > 0 {
			v.Set("instance", instance)
		}
		nextLink = "/timeline/" + tType + "?" + v.Encode()
	}

	commonData := s.getCommonData(c, tType+" timeline ")
	data := &renderer.TimelineData{
		Title:      title,
		Type:       tType,
		Instance:   instance,
		Statuses:   statuses,
		NextLink:   nextLink,
		PrevLink:   prevLink,
		CommonData: commonData,
	}

	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.TimelinePage, data)
}

func (s *service) ThreadPage(c *client, id string, reply bool) (err error) {
	var postContext model.PostContext

	status, err := c.GetStatus(ctx, id)
	if err != nil {
		return
	}

	if reply {
		var content string
		var visibility string
		if c.Session.UserID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for i := range status.Mentions {
			if status.Mentions[i].ID != c.Session.UserID &&
				status.Mentions[i].ID != status.Account.ID {
				content += "@" + status.Mentions[i].Acct + " "
			}
		}

		isDirect := status.Visibility == "direct"
		if isDirect || c.Session.Settings.CopyScope {
			visibility = status.Visibility
		} else {
			visibility = c.Session.Settings.DefaultVisibility
		}

		postContext = model.PostContext{
			DefaultVisibility: visibility,
			DefaultFormat:     c.Session.Settings.DefaultFormat,
			Formats:           s.postFormats,
			ReplyContext: &model.ReplyContext{
				InReplyToID:     id,
				InReplyToName:   status.Account.Acct,
				ReplyContent:    content,
				ForceVisibility: isDirect,
			},
		}
	}

	context, err := c.GetStatusContext(ctx, id)
	if err != nil {
		return
	}

	statuses := append(append(context.Ancestors, status), context.Descendants...)
	replies := make(map[string][]mastodon.ReplyInfo)
	idNumbers := make(map[string]int)

	for i := range statuses {
		statuses[i].ShowReplies = true

		statuses[i].IDNumbers = idNumbers
		idNumbers[statuses[i].ID] = i + 1

		statuses[i].IDReplies = replies
		addToReplyMap(replies, statuses[i].InReplyToID, statuses[i].ID, i+1)
	}

	commonData := s.getCommonData(c, "post by "+status.Account.DisplayName)
	data := &renderer.ThreadData{
		Statuses:    statuses,
		PostContext: postContext,
		ReplyMap:    replies,
		CommonData:  commonData,
	}

	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.ThreadPage, data)
}

func (s *service) LikedByPage(c *client, id string) (err error) {
	likers, err := c.GetFavouritedBy(ctx, id, nil)
	if err != nil {
		return
	}
	commonData := s.getCommonData(c, "likes")
	data := &renderer.LikedByData{
		CommonData: commonData,
		Users:      likers,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.LikedByPage, data)
}

func (s *service) RetweetedByPage(c *client, id string) (err error) {
	retweeters, err := c.GetRebloggedBy(ctx, id, nil)
	if err != nil {
		return
	}
	commonData := s.getCommonData(c, "retweets")
	data := &renderer.RetweetedByData{
		CommonData: commonData,
		Users:      retweeters,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.RetweetedByPage, data)
}

func (s *service) NotificationPage(c *client, maxID string,
	minID string) (err error) {

	var nextLink string
	var unreadCount int
	var readID string
	var excludes []string
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	if c.Session.Settings.AntiDopamineMode {
		excludes = []string{"follow", "favourite", "reblog"}
	}

	notifications, err := c.GetNotifications(ctx, &pg, excludes)
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

	commonData := s.getCommonData(c, "notifications")
	commonData.RefreshInterval = c.Session.Settings.NotificationInterval
	commonData.Target = "main"
	commonData.Count = unreadCount
	data := &renderer.NotificationData{
		Notifications: notifications,
		UnreadCount:   unreadCount,
		ReadID:        readID,
		NextLink:      nextLink,
		CommonData:    commonData,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.NotificationPage, data)
}

func (s *service) UserPage(c *client, id string, pageType string,
	maxID string, minID string) (err error) {

	var nextLink string
	var statuses []*mastodon.Status
	var users []*mastodon.Account
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	user, err := c.GetAccount(ctx, id)
	if err != nil {
		return
	}
	isCurrent := c.Session.UserID == user.ID

	switch pageType {
	case "":
		statuses, err = c.GetAccountStatuses(ctx, id, false, &pg)
		if err != nil {
			return
		}
		if len(statuses) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s?max_id=%s", id,
				pg.MaxID)
		}
	case "following":
		users, err = c.GetAccountFollowing(ctx, id, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/following?max_id=%s",
				id, pg.MaxID)
		}
	case "followers":
		users, err = c.GetAccountFollowers(ctx, id, &pg)
		if err != nil {
			return
		}
		if len(users) == 20 && len(pg.MaxID) > 0 {
			nextLink = fmt.Sprintf("/user/%s/followers?max_id=%s",
				id, pg.MaxID)
		}
	case "media":
		statuses, err = c.GetAccountStatuses(ctx, id, true, &pg)
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
		statuses, err = c.GetBookmarks(ctx, &pg)
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
		users, err = c.GetMutes(ctx, &pg)
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
		users, err = c.GetBlocks(ctx, &pg)
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
		statuses, err = c.GetFavourites(ctx, &pg)
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
		users, err = c.GetFollowRequests(ctx, &pg)
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

	for i := range statuses {
		if statuses[i].Reblog != nil {
			statuses[i].Reblog.RetweetedByID = statuses[i].ID
		}
	}

	commonData := s.getCommonData(c, user.DisplayName)
	data := &renderer.UserData{
		User:       user,
		IsCurrent:  isCurrent,
		Type:       pageType,
		Users:      users,
		Statuses:   statuses,
		NextLink:   nextLink,
		CommonData: commonData,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.UserPage, data)
}

func (s *service) UserSearchPage(c *client,
	id string, q string, offset int) (err error) {

	var nextLink string
	var title = "search"

	user, err := c.GetAccount(ctx, id)
	if err != nil {
		return
	}

	var results *mastodon.Results
	if len(q) > 0 {
		results, err = c.Search(ctx, q, "statuses", 20, true, offset, id)
		if err != nil {
			return err
		}
	} else {
		results = &mastodon.Results{}
	}

	if len(results.Statuses) == 20 {
		offset += 20
		nextLink = fmt.Sprintf("/usersearch/%s?q=%s&offset=%d", id,
			url.QueryEscape(q), offset)
	}

	qq := template.HTMLEscapeString(q)
	if len(q) > 0 {
		title += " \"" + qq + "\""
	}

	commonData := s.getCommonData(c, title)
	data := &renderer.UserSearchData{
		CommonData: commonData,
		User:       user,
		Q:          qq,
		Statuses:   results.Statuses,
		NextLink:   nextLink,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.UserSearchPage, data)
}

func (s *service) AboutPage(c *client) (err error) {
	commonData := s.getCommonData(c, "about")
	data := &renderer.AboutData{
		CommonData: commonData,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.AboutPage, data)
}

func (s *service) EmojiPage(c *client) (err error) {
	emojis, err := c.GetInstanceEmojis(ctx)
	if err != nil {
		return
	}
	commonData := s.getCommonData(c, "emojis")
	data := &renderer.EmojiData{
		Emojis:     emojis,
		CommonData: commonData,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.EmojiPage, data)
}

func (s *service) SearchPage(c *client,
	q string, qType string, offset int) (err error) {

	var nextLink string
	var title = "search"

	var results *mastodon.Results
	if len(q) > 0 {
		results, err = c.Search(ctx, q, qType, 20, true, offset, "")
		if err != nil {
			return err
		}
	} else {
		results = &mastodon.Results{}
	}

	if (qType == "accounts" && len(results.Accounts) == 20) ||
		(qType == "statuses" && len(results.Statuses) == 20) {
		offset += 20
		nextLink = fmt.Sprintf("/search?q=%s&type=%s&offset=%d",
			url.QueryEscape(q), qType, offset)
	}

	qq := template.HTMLEscapeString(q)
	if len(q) > 0 {
		title += " \"" + qq + "\""
	}

	commonData := s.getCommonData(c, title)
	data := &renderer.SearchData{
		CommonData: commonData,
		Q:          qq,
		Type:       qType,
		Users:      results.Accounts,
		Statuses:   results.Statuses,
		NextLink:   nextLink,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.SearchPage, data)
}

func (s *service) SettingsPage(c *client) (err error) {
	commonData := s.getCommonData(c, "settings")
	data := &renderer.SettingsData{
		CommonData:  commonData,
		Settings:    &c.Session.Settings,
		PostFormats: s.postFormats,
	}
	rCtx := getRendererContext(c)
	return s.renderer.Render(rCtx, c, renderer.SettingsPage, data)
}

func (s *service) SingleInstance() (instance string, ok bool) {
	if len(s.singleInstance) > 0 {
		instance = s.singleInstance
		ok = true
	}
	return
}

func (s *service) NewSession(instance string) (rurl string, sid string, err error) {
	var instanceURL string
	if strings.HasPrefix(instance, "https://") {
		instanceURL = instance
		instance = strings.TrimPrefix(instance, "https://")
	} else {
		instanceURL = "https://" + instance
	}

	sid, err = util.NewSessionID()
	if err != nil {
		return
	}
	csrfToken, err := util.NewCSRFToken()
	if err != nil {
		return
	}

	session := model.Session{
		ID:             sid,
		InstanceDomain: instance,
		CSRFToken:      csrfToken,
		Settings:       *model.NewSettings(),
	}
	err = s.sessionRepo.Add(session)
	if err != nil {
		return
	}

	app, err := s.appRepo.Get(instance)
	if err != nil {
		if err != model.ErrAppNotFound {
			return
		}
		mastoApp, err := mastodon.RegisterApp(ctx, &mastodon.AppConfig{
			Server:       instanceURL,
			ClientName:   s.clientName,
			Scopes:       s.clientScope,
			Website:      s.clientWebsite,
			RedirectURIs: s.clientWebsite + "/oauth_callback",
		})
		if err != nil {
			return "", "", err
		}
		app = model.App{
			InstanceDomain: instance,
			InstanceURL:    instanceURL,
			ClientID:       mastoApp.ClientID,
			ClientSecret:   mastoApp.ClientSecret,
		}
		err = s.appRepo.Add(app)
		if err != nil {
			return "", "", err
		}
	}

	u, err := url.Parse("/oauth/authorize")
	if err != nil {
		return
	}

	q := make(url.Values)
	q.Set("scope", "read write follow")
	q.Set("client_id", app.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", s.clientWebsite+"/oauth_callback")
	u.RawQuery = q.Encode()

	rurl = instanceURL + u.String()
	return
}

func (s *service) Signin(c *client, code string) (token string,
	userID string, err error) {

	if len(code) < 1 {
		err = errInvalidArgument
		return
	}
	err = c.AuthenticateToken(ctx, code, s.clientWebsite+"/oauth_callback")
	if err != nil {
		return
	}
	token = c.GetAccessToken(ctx)

	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return
	}
	userID = u.ID
	return
}

func (s *service) Signout(c *client) (err error) {
	s.sessionRepo.Remove(c.Session.ID)
	return
}

func (s *service) Post(c *client, content string, replyToID string,
	format string, visibility string, isNSFW bool,
	files []*multipart.FileHeader) (id string, err error) {

	var mediaIDs []string
	for _, f := range files {
		a, err := c.UploadMediaFromMultipartFileHeader(ctx, f)
		if err != nil {
			return "", err
		}
		mediaIDs = append(mediaIDs, a.ID)
	}

	tweet := &mastodon.Toot{
		Status:      content,
		InReplyToID: replyToID,
		MediaIDs:    mediaIDs,
		ContentType: format,
		Visibility:  visibility,
		Sensitive:   isNSFW,
	}
	st, err := c.PostStatus(ctx, tweet)
	if err != nil {
		return
	}
	return st.ID, nil
}

func (s *service) Like(c *client, id string) (count int64, err error) {
	st, err := c.Favourite(ctx, id)
	if err != nil {
		return
	}
	count = st.FavouritesCount
	return
}

func (s *service) UnLike(c *client, id string) (count int64, err error) {
	st, err := c.Unfavourite(ctx, id)
	if err != nil {
		return
	}
	count = st.FavouritesCount
	return
}

func (s *service) Retweet(c *client, id string) (count int64, err error) {
	st, err := c.Reblog(ctx, id)
	if err != nil {
		return
	}
	if st.Reblog != nil {
		count = st.Reblog.ReblogsCount
	}
	return
}

func (s *service) UnRetweet(c *client, id string) (
	count int64, err error) {
	st, err := c.Unreblog(ctx, id)
	if err != nil {
		return
	}
	count = st.ReblogsCount
	return
}

func (s *service) Vote(c *client, id string, choices []string) (err error) {
	_, err = c.Vote(ctx, id, choices)
	return
}

func (s *service) Follow(c *client, id string, reblogs *bool) (err error) {
	_, err = c.AccountFollow(ctx, id, reblogs)
	return
}

func (s *service) UnFollow(c *client, id string) (err error) {
	_, err = c.AccountUnfollow(ctx, id)
	return
}

func (s *service) Accept(c *client, id string) (err error) {
	return c.FollowRequestAuthorize(ctx, id)
}

func (s *service) Reject(c *client, id string) (err error) {
	return c.FollowRequestReject(ctx, id)
}

func (s *service) Mute(c *client, id string) (err error) {
	_, err = c.AccountMute(ctx, id)
	return
}

func (s *service) UnMute(c *client, id string) (err error) {
	_, err = c.AccountUnmute(ctx, id)
	return
}

func (s *service) Block(c *client, id string) (err error) {
	_, err = c.AccountBlock(ctx, id)
	return
}

func (s *service) UnBlock(c *client, id string) (err error) {
	_, err = c.AccountUnblock(ctx, id)
	return
}

func (s *service) Subscribe(c *client, id string) (err error) {
	_, err = c.Subscribe(ctx, id)
	return
}

func (s *service) UnSubscribe(c *client, id string) (err error) {
	_, err = c.UnSubscribe(ctx, id)
	return
}

func (s *service) SaveSettings(c *client, settings *model.Settings) (err error) {
	switch settings.NotificationInterval {
	case 0, 30, 60, 120, 300, 600:
	default:
		return errInvalidArgument
	}
	session, err := s.sessionRepo.Get(c.Session.ID)
	if err != nil {
		return
	}
	session.Settings = *settings
	return s.sessionRepo.Add(session)
}

func (s *service) MuteConversation(c *client, id string) (err error) {
	_, err = c.MuteConversation(ctx, id)
	return
}

func (s *service) UnMuteConversation(c *client, id string) (err error) {
	_, err = c.UnmuteConversation(ctx, id)
	return
}

func (s *service) Delete(c *client, id string) (err error) {
	return c.DeleteStatus(ctx, id)
}

func (s *service) ReadNotifications(c *client, maxID string) (err error) {
	return c.ReadNotifications(ctx, maxID)
}

func (s *service) Bookmark(c *client, id string) (err error) {
	_, err = c.Bookmark(ctx, id)
	return
}

func (s *service) UnBookmark(c *client, id string) (err error) {
	_, err = c.Unbookmark(ctx, id)
	return
}
