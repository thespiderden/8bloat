package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"

	"bloat/model"
	"bloat/renderer"
	"bloat/util"
	"mastodon"
)

var (
	errInvalidArgument = errors.New("invalid argument")
)

type Service interface {
	ServeErrorPage(ctx context.Context, c *model.Client, err error)
	ServeSigninPage(ctx context.Context, c *model.Client) (err error)
	ServeTimelinePage(ctx context.Context, c *model.Client, tType string, maxID string, minID string) (err error)
	ServeThreadPage(ctx context.Context, c *model.Client, id string, reply bool) (err error)
	ServeLikedByPage(ctx context.Context, c *model.Client, id string) (err error)
	ServeRetweetedByPage(ctx context.Context, c *model.Client, id string) (err error)
	ServeFollowingPage(ctx context.Context, c *model.Client, id string, maxID string, minID string) (err error)
	ServeFollowersPage(ctx context.Context, c *model.Client, id string, maxID string, minID string) (err error)
	ServeNotificationPage(ctx context.Context, c *model.Client, maxID string, minID string) (err error)
	ServeUserPage(ctx context.Context, c *model.Client, id string, maxID string, minID string) (err error)
	ServeAboutPage(ctx context.Context, c *model.Client) (err error)
	ServeEmojiPage(ctx context.Context, c *model.Client) (err error)
	ServeSearchPage(ctx context.Context, c *model.Client, q string, qType string, offset int) (err error)
	ServeSettingsPage(ctx context.Context, c *model.Client) (err error)
	NewSession(ctx context.Context, instance string) (redirectUrl string, sessionID string, err error)
	Signin(ctx context.Context, c *model.Client, sessionID string, code string) (token string, err error)
	Post(ctx context.Context, c *model.Client, content string, replyToID string, format string,
		visibility string, isNSFW bool, files []*multipart.FileHeader) (id string, err error)
	Like(ctx context.Context, c *model.Client, id string) (count int64, err error)
	UnLike(ctx context.Context, c *model.Client, id string) (count int64, err error)
	Retweet(ctx context.Context, c *model.Client, id string) (count int64, err error)
	UnRetweet(ctx context.Context, c *model.Client, id string) (count int64, err error)
	Follow(ctx context.Context, c *model.Client, id string) (err error)
	UnFollow(ctx context.Context, c *model.Client, id string) (err error)
	SaveSettings(ctx context.Context, c *model.Client, settings *model.Settings) (err error)
}

type service struct {
	clientName    string
	clientScope   string
	clientWebsite string
	customCSS     string
	postFormats   []model.PostFormat
	renderer      renderer.Renderer
	sessionRepo   model.SessionRepo
	appRepo       model.AppRepo
}

func NewService(clientName string,
	clientScope string,
	clientWebsite string,
	customCSS string,
	postFormats []model.PostFormat,
	renderer renderer.Renderer,
	sessionRepo model.SessionRepo,
	appRepo model.AppRepo,
) Service {
	return &service{
		clientName:    clientName,
		clientScope:   clientScope,
		clientWebsite: clientWebsite,
		customCSS:     customCSS,
		postFormats:   postFormats,
		renderer:      renderer,
		sessionRepo:   sessionRepo,
		appRepo:       appRepo,
	}
}

func getRendererContext(c *model.Client) *renderer.Context {
	var settings model.Settings
	var session model.Session
	if c != nil {
		settings = c.Session.Settings
		session = c.Session
	} else {
		settings = *model.NewSettings()
	}
	return &renderer.Context{
		MaskNSFW:       settings.MaskNSFW,
		ThreadInNewTab: settings.ThreadInNewTab,
		FluorideMode:   settings.FluorideMode,
		DarkMode:       settings.DarkMode,
		CSRFToken:      session.CSRFToken,
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

func (svc *service) getCommonData(ctx context.Context, c *model.Client,
	title string) (data *renderer.CommonData, err error) {

	data = new(renderer.CommonData)
	data.HeaderData = &renderer.HeaderData{
		Title:             title + " - " + svc.clientName,
		NotificationCount: 0,
		CustomCSS:         svc.customCSS,
	}

	if c == nil || !c.Session.IsLoggedIn() {
		return
	}

	notifications, err := c.GetNotifications(ctx, nil)
	if err != nil {
		return nil, err
	}

	var notificationCount int
	for i := range notifications {
		if notifications[i].Pleroma != nil &&
			!notifications[i].Pleroma.IsSeen {
			notificationCount++
		}
	}

	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	data.NavbarData = &renderer.NavbarData{
		User:              u,
		NotificationCount: notificationCount,
	}

	data.HeaderData.NotificationCount = notificationCount
	data.HeaderData.CSRFToken = c.Session.CSRFToken

	return
}

func (svc *service) ServeErrorPage(ctx context.Context, c *model.Client, err error) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	commonData, err := svc.getCommonData(ctx, nil, "error")
	if err != nil {
		return
	}

	data := &renderer.ErrorData{
		CommonData: commonData,
		Error:      errStr,
	}

	rCtx := getRendererContext(c)
	svc.renderer.RenderErrorPage(rCtx, c.Writer, data)
}

func (svc *service) ServeSigninPage(ctx context.Context, c *model.Client) (
	err error) {

	commonData, err := svc.getCommonData(ctx, nil, "signin")
	if err != nil {
		return
	}

	data := &renderer.SigninData{
		CommonData: commonData,
	}

	rCtx := getRendererContext(nil)
	return svc.renderer.RenderSigninPage(rCtx, c.Writer, data)
}

func (svc *service) ServeTimelinePage(ctx context.Context, c *model.Client,
	tType string, maxID string, minID string) (err error) {

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
	case "local":
		statuses, err = c.GetTimelinePublic(ctx, true, &pg)
		title = "Local Timeline"
	case "twkn":
		statuses, err = c.GetTimelinePublic(ctx, false, &pg)
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

	if len(maxID) > 0 && len(statuses) > 0 {
		prevLink = fmt.Sprintf("/timeline/%s?min_id=%s", tType,
			statuses[0].ID)
	}

	if len(minID) > 0 && len(pg.MinID) > 0 {
		newPg := &mastodon.Pagination{MinID: pg.MinID, Limit: 20}
		newStatuses, err := c.GetTimelineHome(ctx, newPg)
		if err != nil {
			return err
		}
		newLen := len(newStatuses)
		if newLen == 20 {
			prevLink = fmt.Sprintf("/timeline/%s?min_id=%s",
				tType, pg.MinID)
		} else {
			i := 20 - newLen - 1
			if len(statuses) > i {
				prevLink = fmt.Sprintf("/timeline/%s?min_id=%s",
					tType, statuses[i].ID)
			}
		}
	}

	if len(pg.MaxID) > 0 {
		nextLink = fmt.Sprintf("/timeline/%s?max_id=%s", tType, pg.MaxID)
	}

	postContext := model.PostContext{
		DefaultVisibility: c.Session.Settings.DefaultVisibility,
		Formats:           svc.postFormats,
	}

	commonData, err := svc.getCommonData(ctx, c, tType+" timeline ")
	if err != nil {
		return
	}

	data := &renderer.TimelineData{
		Title:       title,
		Statuses:    statuses,
		NextLink:    nextLink,
		PrevLink:    prevLink,
		PostContext: postContext,
		CommonData:  commonData,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderTimelinePage(rCtx, c.Writer, data)
}

func (svc *service) ServeThreadPage(ctx context.Context, c *model.Client,
	id string, reply bool) (err error) {

	var postContext model.PostContext

	status, err := c.GetStatus(ctx, id)
	if err != nil {
		return
	}

	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return
	}

	if reply {
		var content string
		var visibility string
		if u.ID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for i := range status.Mentions {
			if status.Mentions[i].ID != u.ID &&
				status.Mentions[i].ID != status.Account.ID {
				content += "@" + status.Mentions[i].Acct + " "
			}
		}

		if c.Session.Settings.CopyScope {
			s, err := c.GetStatus(ctx, id)
			if err != nil {
				return err
			}
			visibility = s.Visibility
		} else {
			visibility = c.Session.Settings.DefaultVisibility
		}

		postContext = model.PostContext{
			DefaultVisibility: visibility,
			Formats:           svc.postFormats,
			ReplyContext: &model.ReplyContext{
				InReplyToID:   id,
				InReplyToName: status.Account.Acct,
				ReplyContent:  content,
			},
			DarkMode: c.Session.Settings.DarkMode,
		}
	}

	context, err := c.GetStatusContext(ctx, id)
	if err != nil {
		return
	}

	statuses := append(append(context.Ancestors, status), context.Descendants...)
	replies := make(map[string][]mastodon.ReplyInfo)

	for i := range statuses {
		statuses[i].ShowReplies = true
		statuses[i].ReplyMap = replies
		addToReplyMap(replies, statuses[i].InReplyToID, statuses[i].ID, i+1)
	}

	commonData, err := svc.getCommonData(ctx, c, "post by "+status.Account.DisplayName)
	if err != nil {
		return
	}

	data := &renderer.ThreadData{
		Statuses:    statuses,
		PostContext: postContext,
		ReplyMap:    replies,
		CommonData:  commonData,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderThreadPage(rCtx, c.Writer, data)
}

func (svc *service) ServeLikedByPage(ctx context.Context, c *model.Client,
	id string) (err error) {

	likers, err := c.GetFavouritedBy(ctx, id, nil)
	if err != nil {
		return
	}

	commonData, err := svc.getCommonData(ctx, c, "likes")
	if err != nil {
		return
	}

	data := &renderer.LikedByData{
		CommonData: commonData,
		Users:      likers,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderLikedByPage(rCtx, c.Writer, data)
}

func (svc *service) ServeRetweetedByPage(ctx context.Context, c *model.Client,
	id string) (err error) {

	retweeters, err := c.GetRebloggedBy(ctx, id, nil)
	if err != nil {
		return
	}

	commonData, err := svc.getCommonData(ctx, c, "retweets")
	if err != nil {
		return
	}

	data := &renderer.RetweetedByData{
		CommonData: commonData,
		Users:      retweeters,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderRetweetedByPage(rCtx, c.Writer, data)
}

func (svc *service) ServeFollowingPage(ctx context.Context, c *model.Client,
	id string, maxID string, minID string) (err error) {

	var nextLink string
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	followings, err := c.GetAccountFollowing(ctx, id, &pg)
	if err != nil {
		return
	}

	if len(followings) == 20 && len(pg.MaxID) > 0 {
		nextLink = "/following/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, c, "following")
	if err != nil {
		return
	}

	data := &renderer.FollowingData{
		CommonData: commonData,
		Users:      followings,
		NextLink:   nextLink,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderFollowingPage(rCtx, c.Writer, data)
}

func (svc *service) ServeFollowersPage(ctx context.Context, c *model.Client,
	id string, maxID string, minID string) (err error) {

	var nextLink string
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	followers, err := c.GetAccountFollowers(ctx, id, &pg)
	if err != nil {
		return
	}

	if len(followers) == 20 && len(pg.MaxID) > 0 {
		nextLink = "/followers/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, c, "followers")
	if err != nil {
		return
	}

	data := &renderer.FollowersData{
		CommonData: commonData,
		Users:      followers,
		NextLink:   nextLink,
	}
	rCtx := getRendererContext(c)
	return svc.renderer.RenderFollowersPage(rCtx, c.Writer, data)
}

func (svc *service) ServeNotificationPage(ctx context.Context, c *model.Client,
	maxID string, minID string) (err error) {

	var nextLink string
	var unreadCount int
	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	notifications, err := c.GetNotifications(ctx, &pg)
	if err != nil {
		return
	}

	for i := range notifications {
		if notifications[i].Status != nil {
			notifications[i].Status.CreatedAt = notifications[i].CreatedAt
			switch notifications[i].Type {
			case "reblog", "favourite":
				notifications[i].Status.HideAccountInfo = true
			}
		}
		if notifications[i].Pleroma != nil && !notifications[i].Pleroma.IsSeen {
			unreadCount++
		}
	}

	if unreadCount > 0 {
		err := c.ReadNotifications(ctx, notifications[0].ID)
		if err != nil {
			return err
		}
	}

	if len(pg.MaxID) > 0 {
		nextLink = "/notifications?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, c, "notifications")
	if err != nil {
		return
	}

	data := &renderer.NotificationData{
		Notifications: notifications,
		NextLink:      nextLink,
		CommonData:    commonData,
	}
	rCtx := getRendererContext(c)
	return svc.renderer.RenderNotificationPage(rCtx, c.Writer, data)
}

func (svc *service) ServeUserPage(ctx context.Context, c *model.Client,
	id string, maxID string, minID string) (err error) {

	var nextLink string

	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	user, err := c.GetAccount(ctx, id)
	if err != nil {
		return
	}

	statuses, err := c.GetAccountStatuses(ctx, id, &pg)
	if err != nil {
		return
	}

	if len(pg.MaxID) > 0 {
		nextLink = "/user/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, c, user.DisplayName)
	if err != nil {
		return
	}

	data := &renderer.UserData{
		User:       user,
		Statuses:   statuses,
		NextLink:   nextLink,
		CommonData: commonData,
	}
	rCtx := getRendererContext(c)
	return svc.renderer.RenderUserPage(rCtx, c.Writer, data)
}

func (svc *service) ServeAboutPage(ctx context.Context, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, c, "about")
	if err != nil {
		return
	}

	data := &renderer.AboutData{
		CommonData: commonData,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderAboutPage(rCtx, c.Writer, data)
}

func (svc *service) ServeEmojiPage(ctx context.Context, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, c, "emojis")
	if err != nil {
		return
	}

	emojis, err := c.GetInstanceEmojis(ctx)
	if err != nil {
		return
	}

	data := &renderer.EmojiData{
		Emojis:     emojis,
		CommonData: commonData,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderEmojiPage(rCtx, c.Writer, data)
}

func (svc *service) ServeSearchPage(ctx context.Context, c *model.Client,
	q string, qType string, offset int) (err error) {

	var nextLink string
	var title = "search"

	results, err := c.Search(ctx, q, qType, 20, true, offset)
	if err != nil {
		return
	}

	if (qType == "accounts" && len(results.Accounts) == 20) ||
		(qType == "statuses" && len(results.Statuses) == 20) {
		offset += 20
		nextLink = fmt.Sprintf("/search?q=%s&type=%s&offset=%d", q, qType, offset)
	}

	if len(q) > 0 {
		title += " \"" + q + "\""
	}

	commonData, err := svc.getCommonData(ctx, c, title)
	if err != nil {
		return
	}

	data := &renderer.SearchData{
		CommonData: commonData,
		Q:          q,
		Type:       qType,
		Users:      results.Accounts,
		Statuses:   results.Statuses,
		NextLink:   nextLink,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderSearchPage(rCtx, c.Writer, data)
}

func (svc *service) ServeSettingsPage(ctx context.Context, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, c, "settings")
	if err != nil {
		return
	}

	data := &renderer.SettingsData{
		CommonData: commonData,
		Settings:   &c.Session.Settings,
	}

	rCtx := getRendererContext(c)
	return svc.renderer.RenderSettingsPage(rCtx, c.Writer, data)
}

func (svc *service) NewSession(ctx context.Context, instance string) (
	redirectUrl string, sessionID string, err error) {

	var instanceURL string
	if strings.HasPrefix(instance, "https://") {
		instanceURL = instance
		instance = strings.TrimPrefix(instance, "https://")
	} else {
		instanceURL = "https://" + instance
	}

	sessionID, err = util.NewSessionID()
	if err != nil {
		return
	}

	csrfToken, err := util.NewCSRFToken()
	if err != nil {
		return
	}

	session := model.Session{
		ID:             sessionID,
		InstanceDomain: instance,
		CSRFToken:      csrfToken,
		Settings:       *model.NewSettings(),
	}

	err = svc.sessionRepo.Add(session)
	if err != nil {
		return
	}

	app, err := svc.appRepo.Get(instance)
	if err != nil {
		if err != model.ErrAppNotFound {
			return
		}

		mastoApp, err := mastodon.RegisterApp(ctx, &mastodon.AppConfig{
			Server:       instanceURL,
			ClientName:   svc.clientName,
			Scopes:       svc.clientScope,
			Website:      svc.clientWebsite,
			RedirectURIs: svc.clientWebsite + "/oauth_callback",
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

		err = svc.appRepo.Add(app)
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
	q.Set("redirect_uri", svc.clientWebsite+"/oauth_callback")
	u.RawQuery = q.Encode()

	redirectUrl = instanceURL + u.String()

	return
}

func (svc *service) Signin(ctx context.Context, c *model.Client,
	sessionID string, code string) (token string, err error) {

	if len(code) < 1 {
		err = errInvalidArgument
		return
	}

	err = c.AuthenticateToken(ctx, code, svc.clientWebsite+"/oauth_callback")
	if err != nil {
		return
	}
	token = c.GetAccessToken(ctx)

	return
}

func (svc *service) Post(ctx context.Context, c *model.Client, content string,
	replyToID string, format string, visibility string, isNSFW bool,
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

	s, err := c.PostStatus(ctx, tweet)
	if err != nil {
		return
	}

	return s.ID, nil
}

func (svc *service) Like(ctx context.Context, c *model.Client, id string) (
	count int64, err error) {
	s, err := c.Favourite(ctx, id)
	if err != nil {
		return
	}
	count = s.FavouritesCount
	return
}

func (svc *service) UnLike(ctx context.Context, c *model.Client, id string) (
	count int64, err error) {
	s, err := c.Unfavourite(ctx, id)
	if err != nil {
		return
	}
	count = s.FavouritesCount
	return
}

func (svc *service) Retweet(ctx context.Context, c *model.Client, id string) (
	count int64, err error) {
	s, err := c.Reblog(ctx, id)
	if err != nil {
		return
	}
	if s.Reblog != nil {
		count = s.Reblog.ReblogsCount
	}
	return
}

func (svc *service) UnRetweet(ctx context.Context, c *model.Client, id string) (
	count int64, err error) {
	s, err := c.Unreblog(ctx, id)
	if err != nil {
		return
	}
	count = s.ReblogsCount
	return
}

func (svc *service) Follow(ctx context.Context, c *model.Client, id string) (err error) {
	_, err = c.AccountFollow(ctx, id)
	return
}

func (svc *service) UnFollow(ctx context.Context, c *model.Client, id string) (err error) {
	_, err = c.AccountUnfollow(ctx, id)
	return
}

func (svc *service) SaveSettings(ctx context.Context, c *model.Client,
	settings *model.Settings) (err error) {

	session, err := svc.sessionRepo.Get(c.Session.ID)
	if err != nil {
		return
	}

	session.Settings = *settings
	return svc.sessionRepo.Add(session)
}
