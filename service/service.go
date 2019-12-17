package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"mastodon"
	"web/model"
	"web/renderer"
	"web/util"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidClient   = errors.New("invalid client")
)

type Service interface {
	ServeHomePage(ctx context.Context, client io.Writer) (err error)
	GetAuthUrl(ctx context.Context, instance string) (url string, sessionID string, err error)
	GetUserToken(ctx context.Context, sessionID string, c *mastodon.Client, token string) (accessToken string, err error)
	ServeErrorPage(ctx context.Context, client io.Writer, err error)
	ServeSigninPage(ctx context.Context, client io.Writer) (err error)
	ServeTimelinePage(ctx context.Context, client io.Writer, c *mastodon.Client, maxID string, sinceID string, minID string) (err error)
	ServeThreadPage(ctx context.Context, client io.Writer, c *mastodon.Client, id string, reply bool) (err error)
	ServeNotificationPage(ctx context.Context, client io.Writer, c *mastodon.Client, maxID string, minID string) (err error)
	Like(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error)
	UnLike(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error)
	Retweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error)
	UnRetweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error)
	PostTweet(ctx context.Context, client io.Writer, c *mastodon.Client, content string, replyToID string, files []*multipart.FileHeader) (id string, err error)
}

type service struct {
	clientName    string
	clientScope   string
	clientWebsite string
	renderer      renderer.Renderer
	sessionRepo   model.SessionRepository
	appRepo       model.AppRepository
}

func NewService(clientName string, clientScope string, clientWebsite string,
	renderer renderer.Renderer, sessionRepo model.SessionRepository,
	appRepo model.AppRepository) Service {
	return &service{
		clientName:    clientName,
		clientScope:   clientScope,
		clientWebsite: clientWebsite,
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
	err = svc.sessionRepo.Add(model.Session{
		ID:             sessionID,
		InstanceDomain: instance,
	})
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

func (svc *service) GetUserToken(ctx context.Context, sessionID string, c *mastodon.Client,
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
	err = svc.renderer.RenderHomePage(ctx, client)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeErrorPage(ctx context.Context, client io.Writer, err error) {
	svc.renderer.RenderErrorPage(ctx, client, err)
}

func (svc *service) ServeSigninPage(ctx context.Context, client io.Writer) (err error) {
	err = svc.renderer.RenderSigninPage(ctx, client)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeTimelinePage(ctx context.Context, client io.Writer,
	c *mastodon.Client, maxID string, sinceID string, minID string) (err error) {

	var hasNext, hasPrev bool
	var nextLink, prevLink string

	var pg = mastodon.Pagination{
		MaxID: maxID,
		MinID: minID,
		Limit: 20,
	}

	statuses, err := c.GetTimelineHome(ctx, &pg)
	if err != nil {
		return err
	}

	if len(maxID) > 0 && len(statuses) > 0 {
		hasPrev = true
		prevLink = "/timeline?min_id=" + statuses[0].ID
	}
	if len(minID) > 0 && len(pg.MinID) > 0 {
		newStatuses, err := c.GetTimelineHome(ctx, &mastodon.Pagination{MinID: pg.MinID, Limit: 20})
		if err != nil {
			return err
		}
		newStatusesLen := len(newStatuses)
		if newStatusesLen == 20 {
			hasPrev = true
			prevLink = "/timeline?min_id=" + pg.MinID
		} else {
			i := 20 - newStatusesLen - 1
			if len(statuses) > i {
				hasPrev = true
				prevLink = "/timeline?min_id=" + statuses[i].ID
			}
		}
	}
	if len(pg.MaxID) > 0 {
		hasNext = true
		nextLink = "/timeline?max_id=" + pg.MaxID
	}

	navbarData, err := svc.getNavbarTemplateData(ctx, client, c)
	if err != nil {
		return
	}

	data := renderer.NewTimelinePageTemplateData(statuses, hasNext, nextLink, hasPrev, prevLink, navbarData)
	err = svc.renderer.RenderTimelinePage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeThreadPage(ctx context.Context, client io.Writer, c *mastodon.Client, id string, reply bool) (err error) {
	status, err := c.GetStatus(ctx, id)
	if err != nil {
		return
	}

	context, err := c.GetStatusContext(ctx, id)
	if err != nil {
		return
	}

	u, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		return
	}

	var content string
	if reply {
		if u.ID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for _, m := range status.Mentions {
			if u.ID != m.ID {
				content += "@" + m.Acct + " "
			}
		}
	}

	navbarData, err := svc.getNavbarTemplateData(ctx, client, c)
	if err != nil {
		return
	}

	data := renderer.NewThreadPageTemplateData(status, context, reply, id, content, navbarData)
	err = svc.renderer.RenderThreadPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) ServeNotificationPage(ctx context.Context, client io.Writer, c *mastodon.Client, maxID string, minID string) (err error) {
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
		switch notifications[i].Type {
		case "reblog", "favourite":
			if notifications[i].Status != nil {
				notifications[i].Status.Account.ID = ""
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

	navbarData, err := svc.getNavbarTemplateData(ctx, client, c)
	if err != nil {
		return
	}

	data := renderer.NewNotificationPageTemplateData(notifications, hasNext, nextLink, navbarData)
	err = svc.renderer.RenderNotificationPage(ctx, client, data)
	if err != nil {
		return
	}

	return
}

func (svc *service) getNavbarTemplateData(ctx context.Context, client io.Writer, c *mastodon.Client) (data *renderer.NavbarTemplateData, err error) {
	notifications, err := c.GetNotifications(ctx, nil)
	if err != nil {
		return
	}

	var notificationCount int
	for i := range notifications {
		if notifications[i].Pleroma != nil && !notifications[i].Pleroma.IsSeen {
			notificationCount++
		}
	}

	data = renderer.NewNavbarTemplateData(notificationCount)

	return
}

func (svc *service) Like(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	_, err = c.Favourite(ctx, id)
	return
}

func (svc *service) UnLike(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	_, err = c.Unfavourite(ctx, id)
	return
}

func (svc *service) Retweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	_, err = c.Reblog(ctx, id)
	return
}

func (svc *service) UnRetweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	_, err = c.Unreblog(ctx, id)
	return
}

func (svc *service) PostTweet(ctx context.Context, client io.Writer, c *mastodon.Client, content string, replyToID string, files []*multipart.FileHeader) (id string, err error) {
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
	}

	s, err := c.PostStatus(ctx, tweet)
	if err != nil {
		return
	}

	return s.ID, nil
}
