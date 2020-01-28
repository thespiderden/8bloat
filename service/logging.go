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

func (s *ls) ServeFollowingPage(ctx context.Context, c *model.Client, id string,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeFollowingPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeFollowingPage(ctx, c, id, maxID, minID)
}

func (s *ls) ServeFollowersPage(ctx context.Context, c *model.Client, id string,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeFollowersPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeFollowersPage(ctx, c, id, maxID, minID)
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
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeUserPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserPage(ctx, c, id, maxID, minID)
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
	code string) (token string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, session_id=%v, took=%v, err=%v\n",
			"Signin", sessionID, time.Since(begin), err)
	}(time.Now())
	return s.Service.Signin(ctx, c, sessionID, code)
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

func (s *ls) Follow(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Follow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Follow(ctx, c, id)
}

func (s *ls) UnFollow(ctx context.Context, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnFollow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnFollow(ctx, c, id)
}

func (s *ls) SaveSettings(ctx context.Context, c *model.Client, settings *model.Settings) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"SaveSettings", time.Since(begin), err)
	}(time.Now())
	return s.Service.SaveSettings(ctx, c, settings)
}
