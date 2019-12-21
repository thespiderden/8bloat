package service

import (
	"context"
	"errors"
	"io"
	"mastodon"
	"mime/multipart"
	"web/model"
)

var (
	ErrInvalidSession = errors.New("invalid session")
)

type authService struct {
	sessionRepo model.SessionRepository
	appRepo     model.AppRepository
	Service
}

func NewAuthService(sessionRepo model.SessionRepository, appRepo model.AppRepository, s Service) Service {
	return &authService{sessionRepo, appRepo, s}
}

func getSessionID(ctx context.Context) (sessionID string, err error) {
	sessionID, ok := ctx.Value("session_id").(string)
	if !ok || len(sessionID) < 1 {
		return "", ErrInvalidSession
	}
	return sessionID, nil
}

func (s *authService) getClient(ctx context.Context) (c *mastodon.Client, err error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return nil, ErrInvalidSession
	}
	session, err := s.sessionRepo.Get(sessionID)
	if err != nil {
		return nil, ErrInvalidSession
	}
	client, err := s.appRepo.Get(session.InstanceDomain)
	if err != nil {
		return
	}
	c = mastodon.NewClient(&mastodon.Config{
		Server:       client.InstanceURL,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		AccessToken:  session.AccessToken,
	})
	return c, nil
}

func (s *authService) GetAuthUrl(ctx context.Context, instance string) (
	redirectUrl string, sessionID string, err error) {
	return s.Service.GetAuthUrl(ctx, instance)
}

func (s *authService) GetUserToken(ctx context.Context, sessionID string, c *mastodon.Client,
	code string) (token string, err error) {
	sessionID, err = getSessionID(ctx)
	if err != nil {
		return
	}
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}

	token, err = s.Service.GetUserToken(ctx, sessionID, c, code)
	if err != nil {
		return
	}

	err = s.sessionRepo.Update(sessionID, token)
	if err != nil {
		return
	}

	return
}

func (s *authService) ServeHomePage(ctx context.Context, client io.Writer) (err error) {
	return s.Service.ServeHomePage(ctx, client)
}

func (s *authService) ServeErrorPage(ctx context.Context, client io.Writer, err error) {
	s.Service.ServeErrorPage(ctx, client, err)
}

func (s *authService) ServeSigninPage(ctx context.Context, client io.Writer) (err error) {
	return s.Service.ServeSigninPage(ctx, client)
}

func (s *authService) ServeTimelinePage(ctx context.Context, client io.Writer,
	c *mastodon.Client, maxID string, sinceID string, minID string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.ServeTimelinePage(ctx, client, c, maxID, sinceID, minID)
}

func (s *authService) ServeThreadPage(ctx context.Context, client io.Writer, c *mastodon.Client, id string, reply bool) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.ServeThreadPage(ctx, client, c, id, reply)
}

func (s *authService) ServeNotificationPage(ctx context.Context, client io.Writer, c *mastodon.Client, maxID string, minID string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.ServeNotificationPage(ctx, client, c, maxID, minID)
}

func (s *authService) ServeUserPage(ctx context.Context, client io.Writer, c *mastodon.Client, id string, maxID string, minID string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.ServeUserPage(ctx, client, c, id, maxID, minID)
}

func (s *authService) ServeAboutPage(ctx context.Context, client io.Writer, c *mastodon.Client) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.ServeAboutPage(ctx, client, c)
}

func (s *authService) Like(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.Like(ctx, client, c, id)
}

func (s *authService) UnLike(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.UnLike(ctx, client, c, id)
}

func (s *authService) Retweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.Retweet(ctx, client, c, id)
}

func (s *authService) UnRetweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.UnRetweet(ctx, client, c, id)
}

func (s *authService) PostTweet(ctx context.Context, client io.Writer, c *mastodon.Client, content string, replyToID string, files []*multipart.FileHeader) (id string, err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.PostTweet(ctx, client, c, content, replyToID, files)
}

func (s *authService) Follow(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.Follow(ctx, client, c, id)
}

func (s *authService) UnFollow(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	c, err = s.getClient(ctx)
	if err != nil {
		return
	}
	return s.Service.UnFollow(ctx, client, c, id)
}
