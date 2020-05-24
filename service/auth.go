package service

import (
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

func (s *as) authenticateClient(c *model.Client) (err error) {
	if len(c.Ctx.SessionID) < 1 {
		return errInvalidSession
	}
	session, err := s.sessionRepo.Get(c.Ctx.SessionID)
	if err != nil {
		return errInvalidSession
	}
	app, err := s.appRepo.Get(session.InstanceDomain)
	if err != nil {
		return
	}
	mc := mastodon.NewClient(&mastodon.Config{
		Server:       app.InstanceURL,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		AccessToken:  session.AccessToken,
	})
	c.Client = mc
	c.Session = session
	return nil
}

func checkCSRF(c *model.Client) (err error) {
	if c.Ctx.CSRFToken != c.Session.CSRFToken {
		return errInvalidCSRFToken
	}
	return nil
}

func (s *as) ServeErrorPage(c *model.Client, err error) {
	s.authenticateClient(c)
	s.Service.ServeErrorPage(c, err)
}

func (s *as) ServeSigninPage(c *model.Client) (err error) {
	return s.Service.ServeSigninPage(c)
}

func (s *as) ServeRootPage(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeRootPage(c)
}

func (s *as) ServeNavPage(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeNavPage(c)
}

func (s *as) ServeTimelinePage(c *model.Client, tType string,
	maxID string, minID string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeTimelinePage(c, tType, maxID, minID)
}

func (s *as) ServeThreadPage(c *model.Client, id string, reply bool) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeThreadPage(c, id, reply)
}

func (s *as) ServeLikedByPage(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeLikedByPage(c, id)
}

func (s *as) ServeRetweetedByPage(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeRetweetedByPage(c, id)
}

func (s *as) ServeNotificationPage(c *model.Client,
	maxID string, minID string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeNotificationPage(c, maxID, minID)
}

func (s *as) ServeUserPage(c *model.Client, id string,
	pageType string, maxID string, minID string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeUserPage(c, id, pageType, maxID, minID)
}

func (s *as) ServeAboutPage(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeAboutPage(c)
}

func (s *as) ServeEmojiPage(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeEmojiPage(c)
}

func (s *as) ServeSearchPage(c *model.Client, q string,
	qType string, offset int) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeSearchPage(c, q, qType, offset)
}

func (s *as) ServeUserSearchPage(c *model.Client,
	id string, q string, offset int) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeUserSearchPage(c, id, q, offset)
}

func (s *as) ServeSettingsPage(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	return s.Service.ServeSettingsPage(c)
}

func (s *as) NewSession(instance string) (redirectUrl string,
	sessionID string, err error) {
	return s.Service.NewSession(instance)
}

func (s *as) Signin(c *model.Client, sessionID string,
	code string) (token string, userID string, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}

	token, userID, err = s.Service.Signin(c, c.Session.ID, code)
	if err != nil {
		return
	}

	c.Session.AccessToken = token
	c.Session.UserID = userID

	err = s.sessionRepo.Add(c.Session)
	if err != nil {
		return
	}

	return
}

func (s *as) Signout(c *model.Client) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	s.Service.Signout(c)
	return
}

func (s *as) Post(c *model.Client, content string,
	replyToID string, format string, visibility string, isNSFW bool,
	files []*multipart.FileHeader) (id string, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Post(c, content, replyToID, format, visibility, isNSFW, files)
}

func (s *as) Like(c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Like(c, id)
}

func (s *as) UnLike(c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnLike(c, id)
}

func (s *as) Retweet(c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Retweet(c, id)
}

func (s *as) UnRetweet(c *model.Client, id string) (count int64, err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnRetweet(c, id)
}

func (s *as) Vote(c *model.Client, id string,
	choices []string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Vote(c, id, choices)
}

func (s *as) Follow(c *model.Client, id string, reblogs *bool) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Follow(c, id, reblogs)
}

func (s *as) UnFollow(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnFollow(c, id)
}

func (s *as) Mute(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Mute(c, id)
}

func (s *as) UnMute(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnMute(c, id)
}

func (s *as) Block(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Block(c, id)
}

func (s *as) UnBlock(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnBlock(c, id)
}

func (s *as) Subscribe(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Subscribe(c, id)
}

func (s *as) UnSubscribe(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnSubscribe(c, id)
}

func (s *as) SaveSettings(c *model.Client, settings *model.Settings) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.SaveSettings(c, settings)
}

func (s *as) MuteConversation(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.MuteConversation(c, id)
}

func (s *as) UnMuteConversation(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.UnMuteConversation(c, id)
}

func (s *as) Delete(c *model.Client, id string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.Delete(c, id)
}

func (s *as) ReadNotifications(c *model.Client,
	maxID string) (err error) {
	err = s.authenticateClient(c)
	if err != nil {
		return
	}
	err = checkCSRF(c)
	if err != nil {
		return
	}
	return s.Service.ReadNotifications(c, maxID)
}
