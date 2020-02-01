package service

import (
	"context"
	"errors"
	"mime/multipart"

	"bloat/mastodon"
	"bloat/model"
)

var (
	errInvalidSession   = errors.New("invalid session")
	errInvalidCSRFToken = errors.New("invalid csrf token")
)

type as struct {
	sessionRepo model.SessionRepo
	appRepo     model.AppRepo
	Service
}

func NewAuthService(sessionRepo model.SessionRepo, appRepo model.AppRepo, s Service) Service {
	return &as{sessionRepo, appRepo, s}
}

func (s *as) authenticateClient(ctx context.Context, c *model.Client) (err error) {
	sessionID, ok := ctx.Value("session_id").(string)
	if !ok || len(sessionID) < 1 {
		return errInvalidSession
	}
	session, err := s.sessionRepo.Get(sessionID)
	if err != nil {
		return errInvalidSession
	}
	client, err := s.appRepo.Get(session.InstanceDomain)
	if err != nil {
		return
	}
	mc := mastodon.NewClient(&mastodon.Config{
		Server:       client.InstanceURL,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		AccessToken:  session.AccessToken,
	})
	if c == nil {
		c = &model.Client{}
	}
	c.Client = mc
	c.Session = session
	return nil
}

func checkCSRF(ctx context.Context, c *model.Client) (err error) {
	token, ok := ctx.Value("csrf_token").(string)
	if !ok || token != c.Session.CSRFToken {
		return errInvalidCSRFToken
	}
	return nil
}

func (s *as) ServeErrorPage(ctx context.Context, c *model.Client, err error) {
	s.authenticateClient(ctx, c)
	s.Service.ServeErrorPage(ctx, c, err)
}

func (s *as) ServeSigninPage(ctx context.Context, c *model.Client) (err error) {
	return s.Service.ServeSigninPage(ctx, c)
}

func (s *as) ServeTimelinePage(ctx context.Context, c *model.Client, tType string,
	maxID string, minID string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeTimelinePage(ctx, c, tType, maxID, minID)
}

func (s *as) ServeThreadPage(ctx context.Context, c *model.Client, id string, reply bool) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeThreadPage(ctx, c, id, reply)
}

func (s *as) ServeLikedByPage(ctx context.Context, c *model.Client, id string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeLikedByPage(ctx, c, id)
}

func (s *as) ServeRetweetedByPage(ctx context.Context, c *model.Client, id string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeRetweetedByPage(ctx, c, id)
}

func (s *as) ServeNotificationPage(ctx context.Context, c *model.Client,
	maxID string, minID string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeNotificationPage(ctx, c, maxID, minID)
}

func (s *as) ServeUserPage(ctx context.Context, c *model.Client, id string,
	pageType string, maxID string, minID string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeUserPage(ctx, c, id, pageType, maxID, minID)
}

func (s *as) ServeAboutPage(ctx context.Context, c *model.Client) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeAboutPage(ctx, c)
}

func (s *as) ServeEmojiPage(ctx context.Context, c *model.Client) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeEmojiPage(ctx, c)
}

func (s *as) ServeSearchPage(ctx context.Context, c *model.Client, q string,
	qType string, offset int) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeSearchPage(ctx, c, q, qType, offset)
}

func (s *as) ServeUserSearchPage(ctx context.Context, c *model.Client,
	id string, q string, offset int) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeUserSearchPage(ctx, c, id, q, offset)
}

func (s *as) ServeSettingsPage(ctx context.Context, c *model.Client) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	return s.Service.ServeSettingsPage(ctx, c)
}

func (s *as) NewSession(ctx context.Context, instance string) (redirectUrl string,
	sessionID string, err error) {
	return s.Service.NewSession(ctx, instance)
}

func (s *as) Signin(ctx context.Context, c *model.Client, sessionID string,
	code string) (token string, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}

	token, err = s.Service.Signin(ctx, c, c.Session.ID, code)
	if err != nil {
		return
	}

	c.Session.AccessToken = token
	err = s.sessionRepo.Add(c.Session)
	if err != nil {
		return
	}

	return
}

func (s *as) Post(ctx context.Context, c *model.Client, content string,
	replyToID string, format string, visibility string, isNSFW bool,
	files []*multipart.FileHeader) (id string, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.Post(ctx, c, content, replyToID, format, visibility, isNSFW, files)
}

func (s *as) Like(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.Like(ctx, c, id)
}

func (s *as) UnLike(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.UnLike(ctx, c, id)
}

func (s *as) Retweet(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.Retweet(ctx, c, id)
}

func (s *as) UnRetweet(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.UnRetweet(ctx, c, id)
}

func (s *as) Follow(ctx context.Context, c *model.Client, id string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.Follow(ctx, c, id)
}

func (s *as) UnFollow(ctx context.Context, c *model.Client, id string) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.UnFollow(ctx, c, id)
}

func (s *as) SaveSettings(ctx context.Context, c *model.Client, settings *model.Settings) (err error) {
	err = s.authenticateClient(ctx, c)
	if err != nil {
		return
	}
	err = checkCSRF(ctx, c)
	if err != nil {
		return
	}
	return s.Service.SaveSettings(ctx, c, settings)
}
