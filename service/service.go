package service

import (
	"bytes"
	"context"
	"encoding/json"
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
	"mastodon"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidClient   = errors.New("invalid client")
	ErrInvalidTimeline = errors.New("invalid timeline")
)

type Service interface {
	ServeHomePage(ctx context.Context, client io.Writer) (err error)
	GetAuthUrl(ctx context.Context, instance string) (url string, sessionID string, err error)
	GetUserToken(ctx context.Context, sessionID string, c *model.Client, token string) (accessToken string, err error)
	ServeErrorPage(ctx context.Context, client io.Writer, err error)
	ServeSigninPage(ctx context.Context, client io.Writer) (err error)
	ServeTimelinePage(ctx context.Context, client io.Writer, c *model.Client, timelineType string, maxID string, sinceID string, minID string) (err error)
	ServeThreadPage(ctx context.Context, client io.Writer, c *model.Client, id string, reply bool) (err error)
	ServeNotificationPage(ctx context.Context, client io.Writer, c *model.Client, maxID string, minID string) (err error)
	ServeUserPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error)
	ServeAboutPage(ctx context.Context, client io.Writer, c *model.Client) (err error)
	ServeEmojiPage(ctx context.Context, client io.Writer, c *model.Client) (err error)
	ServeLikedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error)
	ServeRetweetedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error)
	ServeFollowingPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error)
	ServeFollowersPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error)
	ServeSearchPage(ctx context.Context, client io.Writer, c *model.Client, q string, qType string, offset int) (err error)
	ServeSettingsPage(ctx context.Context, client io.Writer, c *model.Client) (err error)
	SaveSettings(ctx context.Context, client io.Writer, c *model.Client, settings *model.Settings) (err error)
	Like(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error)
	UnLike(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error)
	Retweet(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error)
	UnRetweet(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error)
	PostTweet(ctx context.Context, client io.Writer, c *model.Client, content string, replyToID string, format string, visibility string, isNSFW bool, files []*multipart.FileHeader) (id string, err error)
	Follow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error)
	UnFollow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error)
}

type service struct {
	clientName    string
	clientScope   string
	clientWebsite string
	customCSS     string
	postFormats   []model.PostFormat
	renderer      renderer.Renderer
	sessionRepo   model.SessionRepository
	appRepo       model.AppRepository
}

func NewService(clientName string, clientScope string, clientWebsite string,
	customCSS string, postFormats []model.PostFormat, renderer renderer.Renderer,
	sessionRepo model.SessionRepository, appRepo model.AppRepository) Service {
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

func (svc *service) GetAuthUrl(ctx context.Context, instance string) (
	redirectUrl string, sessionID string, err error) {
	var instanceURL string
	if strings.HasPrefix(instance, "https://") {
		instanceURL = instance
		instance = strings.TrimPrefix(instance, "https://")
	} else {
		instanceURL = "https://" + instance
	}

	sessionID = util.NewSessionId()
	session := model.Session{
		ID:             sessionID,
		InstanceDomain: instance,
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

		var mastoApp *mastodon.Application
		mastoApp, err = mastodon.RegisterApp(ctx, &mastodon.AppConfig{
			Server:       instanceURL,
			ClientName:   svc.clientName,
			Scopes:       svc.clientScope,
			Website:      svc.clientWebsite,
			RedirectURIs: svc.clientWebsite + "/oauth_callback",
		})
		if err != nil {
			return
		}

		app = model.App{
			InstanceDomain: instance,
			InstanceURL:    instanceURL,
			ClientID:       mastoApp.ClientID,
			ClientSecret:   mastoApp.ClientSecret,
		}

		err = svc.appRepo.Add(app)
		if err != nil {
			return
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

func (svc *service) GetUserToken(ctx context.Context, sessionID string, c *model.Client,
	code string) (token string, err error) {
	if len(code) < 1 {
		err = ErrInvalidArgument
		return
	}

	session, err := svc.sessionRepo.Get(sessionID)
	if err != nil {
		return
	}

	app, err := svc.appRepo.Get(session.InstanceDomain)
	if err != nil {
		return
	}

	data := &bytes.Buffer{}
	err = json.NewEncoder(data).Encode(map[string]string{
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  svc.clientWebsite + "/oauth_callback",
	})
	if err != nil {
		return
	}

	resp, err := http.Post(app.InstanceURL+"/oauth/token", "application/json", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return
	}
	/*
		err = c.AuthenticateToken(ctx, code, svc.clientWebsite+"/oauth_callback")
		if err != nil {
			return
		}
		err = svc.sessionRepo.Update(sessionID, c.GetAccessToken(ctx))
	*/

	return res.AccessToken, nil
}

func (svc *service) ServeHomePage(ctx context.Context, client io.Writer) (err error) {
	commonData, err := svc.getCommonData(ctx, client, nil, "home")
	if err != nil {
		return
	}

	data := &renderer.HomePageData{
		CommonData: commonData,
	}

	return svc.renderer.RenderHomePage(ctx, client, data)
}

func (svc *service) ServeErrorPage(ctx context.Context, client io.Writer, err error) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	commonData, err := svc.getCommonData(ctx, client, nil, "error")
	if err != nil {
		return
	}

	data := &renderer.ErrorData{
		CommonData: commonData,
		Error:      errStr,
	}

	svc.renderer.RenderErrorPage(ctx, client, data)
}

func (svc *service) ServeSigninPage(ctx context.Context, client io.Writer) (err error) {
	commonData, err := svc.getCommonData(ctx, client, nil, "signin")
	if err != nil {
		return
	}

	data := &renderer.SigninData{
		CommonData: commonData,
	}

	return svc.renderer.RenderSigninPage(ctx, client, data)
}

func (svc *service) ServeTimelinePage(ctx context.Context, client io.Writer,
	c *model.Client, timelineType string, maxID string, sinceID string, minID string) (err error) {

	var hasNext, hasPrev bool
	var nextLink, prevLink string

	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	var statuses []*mastodon.Status
	var title string
	switch timelineType {
	default:
		return ErrInvalidTimeline
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
		statuses[i].ThreadInNewTab = c.Session.Settings.ThreadInNewTab
		statuses[i].MaskNSFW = c.Session.Settings.MaskNSFW
		if statuses[i].Reblog != nil {
			statuses[i].Reblog.RetweetedByID = statuses[i].ID
			statuses[i].Reblog.ThreadInNewTab = c.Session.Settings.ThreadInNewTab
			statuses[i].Reblog.MaskNSFW = c.Session.Settings.MaskNSFW
		}
	}

	if len(maxID) > 0 && len(statuses) > 0 {
		hasPrev = true
		prevLink = fmt.Sprintf("/timeline/$s?min_id=%s", timelineType, statuses[0].ID)
	}
	if len(minID) > 0 && len(pg.MinID) > 0 {
		newStatuses, err := c.GetTimelineHome(ctx, &mastodon.Pagination{MinID: pg.MinID, Limit: 20})
		if err != nil {
			return err
		}
		newStatusesLen := len(newStatuses)
		if newStatusesLen == 20 {
			hasPrev = true
			prevLink = fmt.Sprintf("/timeline/%s?min_id=%s", timelineType, pg.MinID)
		} else {
			i := 20 - newStatusesLen - 1
			if len(statuses) > i {
				hasPrev = true
				prevLink = fmt.Sprintf("/timeline/%s?min_id=%s", timelineType, statuses[i].ID)
			}
		}
	}
	if len(pg.MaxID) > 0 {
		hasNext = true
		nextLink = fmt.Sprintf("/timeline/%s?max_id=%s", timelineType, pg.MaxID)
	}

	postContext := model.PostContext{
		DefaultVisibility: c.Session.Settings.DefaultVisibility,
		Formats:           svc.postFormats,
	}

	commonData, err := svc.getCommonData(ctx, client, c, timelineType+" timeline ")
	if err != nil {
		return
	}

	data := &renderer.TimelineData{
		Title:       title,
		Statuses:    statuses,
		HasNext:     hasNext,
		NextLink:    nextLink,
		HasPrev:     hasPrev,
		PrevLink:    prevLink,
		PostContext: postContext,
		CommonData:  commonData,
	}

	err = svc.renderer.RenderTimelinePage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeThreadPage(ctx context.Context, client io.Writer, c *model.Client, id string, reply bool) (err error) {
	status, err := c.GetStatus(ctx, id)
	if err != nil {
		return
	}

	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return
	}

	var postContext model.PostContext
	if reply {
		var content string
		if u.ID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for i := range status.Mentions {
			if status.Mentions[i].ID != u.ID && status.Mentions[i].ID != status.Account.ID {
				content += "@" + status.Mentions[i].Acct + " "
			}
		}

		var visibility string
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
		}
	}

	context, err := c.GetStatusContext(ctx, id)
	if err != nil {
		return
	}

	statuses := append(append(context.Ancestors, status), context.Descendants...)

	replyMap := make(map[string][]mastodon.ReplyInfo)

	for i := range statuses {
		statuses[i].ShowReplies = true
		statuses[i].ReplyMap = replyMap
		statuses[i].MaskNSFW = c.Session.Settings.MaskNSFW
		addToReplyMap(replyMap, statuses[i].InReplyToID, statuses[i].ID, i+1)
	}

	commonData, err := svc.getCommonData(ctx, client, c, "post by "+status.Account.DisplayName)
	if err != nil {
		return
	}

	data := &renderer.ThreadData{
		Statuses:    statuses,
		PostContext: postContext,
		ReplyMap:    replyMap,
		CommonData:  commonData,
	}

	err = svc.renderer.RenderThreadPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeNotificationPage(ctx context.Context, client io.Writer, c *model.Client, maxID string, minID string) (err error) {
	var hasNext bool
	var nextLink string

	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	notifications, err := c.GetNotifications(ctx, &pg)
	if err != nil {
		return
	}

	var unreadCount int
	for i := range notifications {
		if notifications[i].Status != nil {
			notifications[i].Status.CreatedAt = notifications[i].CreatedAt
			notifications[i].Status.MaskNSFW = c.Session.Settings.MaskNSFW
			switch notifications[i].Type {
			case "reblog", "favourite":
				notifications[i].Status.HideAccountInfo = true
			}
		}
		if notifications[i].Pleroma != nil && notifications[i].Pleroma.IsSeen {
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
		hasNext = true
		nextLink = "/notifications?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, client, c, "notifications")
	if err != nil {
		return
	}

	data := &renderer.NotificationData{
		Notifications: notifications,
		HasNext:       hasNext,
		NextLink:      nextLink,
		CommonData:    commonData,
	}
	err = svc.renderer.RenderNotificationPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeUserPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error) {
	user, err := c.GetAccount(ctx, id)
	if err != nil {
		return
	}

	var hasNext bool
	var nextLink string

	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	statuses, err := c.GetAccountStatuses(ctx, id, &pg)
	if err != nil {
		return
	}

	for i := range statuses {
		statuses[i].MaskNSFW = c.Session.Settings.MaskNSFW
		if statuses[i].Reblog != nil {
			statuses[i].Reblog.MaskNSFW = c.Session.Settings.MaskNSFW
		}
	}

	if len(pg.MaxID) > 0 {
		hasNext = true
		nextLink = "/user/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, client, c, user.DisplayName)
	if err != nil {
		return
	}

	data := &renderer.UserData{
		User:       user,
		Statuses:   statuses,
		HasNext:    hasNext,
		NextLink:   nextLink,
		CommonData: commonData,
	}

	err = svc.renderer.RenderUserPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeAboutPage(ctx context.Context, client io.Writer, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, client, c, "about")
	if err != nil {
		return
	}

	data := &renderer.AboutData{
		CommonData: commonData,
	}
	err = svc.renderer.RenderAboutPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeEmojiPage(ctx context.Context, client io.Writer, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, client, c, "emojis")
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

	err = svc.renderer.RenderEmojiPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeLikedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	likers, err := c.GetFavouritedBy(ctx, id, nil)
	if err != nil {
		return
	}

	commonData, err := svc.getCommonData(ctx, client, c, "likes")
	if err != nil {
		return
	}

	data := &renderer.LikedByData{
		CommonData: commonData,
		Users:      likers,
	}

	err = svc.renderer.RenderLikedByPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeRetweetedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	retweeters, err := c.GetRebloggedBy(ctx, id, nil)
	if err != nil {
		return
	}

	commonData, err := svc.getCommonData(ctx, client, c, "retweets")
	if err != nil {
		return
	}

	data := &renderer.RetweetedByData{
		CommonData: commonData,
		Users:      retweeters,
	}

	err = svc.renderer.RenderRetweetedByPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeFollowingPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error) {
	var hasNext bool
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
		hasNext = true
		nextLink = "/following/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, client, c, "following")
	if err != nil {
		return
	}

	data := &renderer.FollowingData{
		CommonData: commonData,
		Users:      followings,
		HasNext:    hasNext,
		NextLink:   nextLink,
	}

	err = svc.renderer.RenderFollowingPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeFollowersPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error) {
	var hasNext bool
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
		hasNext = true
		nextLink = "/followers/" + id + "?max_id=" + pg.MaxID
	}

	commonData, err := svc.getCommonData(ctx, client, c, "followers")
	if err != nil {
		return
	}

	data := &renderer.FollowersData{
		CommonData: commonData,
		Users:      followers,
		HasNext:    hasNext,
		NextLink:   nextLink,
	}

	err = svc.renderer.RenderFollowersPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeSearchPage(ctx context.Context, client io.Writer, c *model.Client, q string, qType string, offset int) (err error) {
	var hasNext bool
	var nextLink string

	results, err := c.Search(ctx, q, qType, 20, true, offset)
	if err != nil {
		return
	}

	switch qType {
	case "accounts":
		hasNext = len(results.Accounts) == 20
	case "statuses":
		hasNext = len(results.Statuses) == 20
		for i := range results.Statuses {
			results.Statuses[i].MaskNSFW = c.Session.Settings.MaskNSFW
		}

	}

	if hasNext {
		offset += 20
		nextLink = fmt.Sprintf("/search?q=%s&type=%s&offset=%d", q, qType, offset)
	}

	var title = "search"
	if len(q) > 0 {
		title += " \"" + q + "\""
	}
	commonData, err := svc.getCommonData(ctx, client, c, title)
	if err != nil {
		return
	}

	data := &renderer.SearchData{
		CommonData: commonData,
		Q:          q,
		Type:       qType,
		Users:      results.Accounts,
		Statuses:   results.Statuses,
		HasNext:    hasNext,
		NextLink:   nextLink,
	}

	err = svc.renderer.RenderSearchPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeSettingsPage(ctx context.Context, client io.Writer, c *model.Client) (err error) {
	commonData, err := svc.getCommonData(ctx, client, c, "settings")
	if err != nil {
		return
	}

	data := &renderer.SettingsData{
		CommonData: commonData,
		Settings:   &c.Session.Settings,
	}

	err = svc.renderer.RenderSettingsPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) SaveSettings(ctx context.Context, client io.Writer, c *model.Client, settings *model.Settings) (err error) {
	session, err := svc.sessionRepo.Get(c.Session.ID)
	if err != nil {
		return
	}

	session.Settings = *settings
	err = svc.sessionRepo.Add(session)
	if err != nil {
		return
	}

	return
}

func (svc *service) getCommonData(ctx context.Context, client io.Writer, c *model.Client, title string) (data *renderer.CommonData, err error) {
	data = new(renderer.CommonData)

	data.HeaderData = &renderer.HeaderData{
		Title:             title + " - " + svc.clientName,
		NotificationCount: 0,
		CustomCSS:         svc.customCSS,
	}

	if c != nil && c.Session.IsLoggedIn() {
		notifications, err := c.GetNotifications(ctx, nil)
		if err != nil {
			return nil, err
		}

		var notificationCount int
		for i := range notifications {
			if notifications[i].Pleroma != nil && !notifications[i].Pleroma.IsSeen {
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
		data.HeaderData.FluorideMode = c.Session.Settings.FluorideMode
	}

	return
}

func (svc *service) Like(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error) {
	s, err := c.Favourite(ctx, id)
	if err != nil {
		return
	}
	count = s.FavouritesCount
	return
}

func (svc *service) UnLike(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error) {
	s, err := c.Unfavourite(ctx, id)
	if err != nil {
		return
	}
	count = s.FavouritesCount
	return
}

func (svc *service) Retweet(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error) {
	s, err := c.Reblog(ctx, id)
	if err != nil {
		return
	}
	if s.Reblog != nil {
		count = s.Reblog.ReblogsCount
	}
	return
}

func (svc *service) UnRetweet(ctx context.Context, client io.Writer, c *model.Client, id string) (count int64, err error) {
	s, err := c.Unreblog(ctx, id)
	if err != nil {
		return
	}
	count = s.ReblogsCount
	return
}

func (svc *service) PostTweet(ctx context.Context, client io.Writer, c *model.Client, content string, replyToID string, format string, visibility string, isNSFW bool, files []*multipart.FileHeader) (id string, err error) {
	var mediaIds []string
	for _, f := range files {
		a, err := c.UploadMediaFromMultipartFileHeader(ctx, f)
		if err != nil {
			return "", err
		}
		mediaIds = append(mediaIds, a.ID)
	}

	tweet := &mastodon.Toot{
		Status:      content,
		InReplyToID: replyToID,
		MediaIDs:    mediaIds,
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

func (svc *service) Follow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	_, err = c.AccountFollow(ctx, id)
	return
}

func (svc *service) UnFollow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	_, err = c.AccountUnfollow(ctx, id)
	return
}

func addToReplyMap(m map[string][]mastodon.ReplyInfo, key interface{}, val string, number int) {
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
