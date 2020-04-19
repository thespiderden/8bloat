package service

import (
	"context"
	"log"
	"mime/multipart"
	"time"

	"bloat/model"
)

type ls struct {
	logger *log.Logger
	Service
}

func NewLoggingService(logger *log.Logger, s Service) Service {
	return &ls{logger, s}
}

func (s *ls) ServeErrorPage(ctx context.Context, c *model.Client, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, err=%v, took=%v\n",
			"ServeErrorPage", err, time.Since(begin))
	}(time.Now())
	s.Service.ServeErrorPage(ctx, c, err)
}

func (s *ls) ServeSigninPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSigninPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSigninPage(ctx, c)
}

func (s *ls) ServeRootPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeRootPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeRootPage(ctx, c)
}

func (s *ls) ServeNavPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeNavPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeNavPage(ctx, c)
}

func (s *ls) ServeTimelinePage(ctx context.Context, c *model.Client, tType string,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, type=%v, took=%v, err=%v\n",
			"ServeTimelinePage", tType, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeTimelinePage(ctx, c, tType, maxID, minID)
}

func (s *ls) ServeThreadPage(ctx context.Context, c *model.Client, id string,
	reply bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeThreadPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeThreadPage(ctx, c, id, reply)
}

func (s *ls) ServeLikedByPage(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeLikedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeLikedByPage(ctx, c, id)
}

func (s *ls) ServeRetweetedByPage(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeRetweetedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeRetweetedByPage(ctx, c, id)
}

func (s *ls) ServeNotificationPage(ctx context.Context, c *model.Client,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeNotificationPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeNotificationPage(ctx, c, maxID, minID)
}

func (s *ls) ServeUserPage(ctx context.Context, c *model.Client, id string,
	pageType string, maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, type=%v, took=%v, err=%v\n",
			"ServeUserPage", id, pageType, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserPage(ctx, c, id, pageType, maxID, minID)
}

func (s *ls) ServeAboutPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeAboutPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeAboutPage(ctx, c)
}

func (s *ls) ServeEmojiPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeEmojiPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeEmojiPage(ctx, c)
}

func (s *ls) ServeSearchPage(ctx context.Context, c *model.Client, q string,
	qType string, offset int) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSearchPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSearchPage(ctx, c, q, qType, offset)
}

func (s *ls) ServeUserSearchPage(ctx context.Context, c *model.Client,
	id string, q string, offset int) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeUserSearchPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserSearchPage(ctx, c, id, q, offset)
}

func (s *ls) ServeSettingsPage(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSettingsPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSettingsPage(ctx, c)
}

func (s *ls) NewSession(ctx context.Context, instance string) (redirectUrl string,
	sessionID string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, instance=%v, took=%v, err=%v\n",
			"NewSession", instance, time.Since(begin), err)
	}(time.Now())
	return s.Service.NewSession(ctx, instance)
}

func (s *ls) Signin(ctx context.Context, c *model.Client, sessionID string,
	code string) (token string, userID string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, session_id=%v, took=%v, err=%v\n",
			"Signin", sessionID, time.Since(begin), err)
	}(time.Now())
	return s.Service.Signin(ctx, c, sessionID, code)
}

func (s *ls) Signout(ctx context.Context, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"Signout", time.Since(begin), err)
	}(time.Now())
	return s.Service.Signout(ctx, c)
}

func (s *ls) Post(ctx context.Context, c *model.Client, content string,
	replyToID string, format string, visibility string, isNSFW bool,
	files []*multipart.FileHeader) (id string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"Post", time.Since(begin), err)
	}(time.Now())
	return s.Service.Post(ctx, c, content, replyToID, format,
		visibility, isNSFW, files)
}

func (s *ls) Like(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Like", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Like(ctx, c, id)
}

func (s *ls) UnLike(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnLike", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnLike(ctx, c, id)
}

func (s *ls) Retweet(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Retweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Retweet(ctx, c, id)
}

func (s *ls) UnRetweet(ctx context.Context, c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnRetweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnRetweet(ctx, c, id)
}

func (s *ls) Vote(ctx context.Context, c *model.Client, id string, choices []string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Vote", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Vote(ctx, c, id, choices)
}

func (s *ls) Follow(ctx context.Context, c *model.Client, id string, reblogs *bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Follow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Follow(ctx, c, id, reblogs)
}

func (s *ls) UnFollow(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnFollow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnFollow(ctx, c, id)
}

func (s *ls) Mute(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Mute", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Mute(ctx, c, id)
}

func (s *ls) UnMute(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnMute", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnMute(ctx, c, id)
}

func (s *ls) Block(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Block", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Block(ctx, c, id)
}

func (s *ls) UnBlock(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnBlock", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnBlock(ctx, c, id)
}

func (s *ls) Subscribe(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Subscribe", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Subscribe(ctx, c, id)
}

func (s *ls) UnSubscribe(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnSubscribe", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnSubscribe(ctx, c, id)
}

func (s *ls) SaveSettings(ctx context.Context, c *model.Client, settings *model.Settings) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"SaveSettings", time.Since(begin), err)
	}(time.Now())
	return s.Service.SaveSettings(ctx, c, settings)
}

func (s *ls) MuteConversation(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"MuteConversation", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.MuteConversation(ctx, c, id)
}

func (s *ls) UnMuteConversation(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnMuteConversation", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnMuteConversation(ctx, c, id)
}

func (s *ls) Delete(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Delete", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Delete(ctx, c, id)
}

func (s *ls) ReadNotifications(ctx context.Context, c *model.Client,
	maxID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, max_id=%v, took=%v, err=%v\n",
			"ReadNotifications", maxID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ReadNotifications(ctx, c, maxID)
}
