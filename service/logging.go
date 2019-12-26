package service

import (
	"context"
	"io"
	"log"
	"mime/multipart"
	"time"
	"web/model"
)

type loggingService struct {
	logger *log.Logger
	Service
}

func NewLoggingService(logger *log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) GetAuthUrl(ctx context.Context, instance string) (
	redirectUrl string, sessionID string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, instance=%v, took=%v, err=%v\n",
			"GetAuthUrl", instance, time.Since(begin), err)
	}(time.Now())
	return s.Service.GetAuthUrl(ctx, instance)
}

func (s *loggingService) GetUserToken(ctx context.Context, sessionID string, c *model.Client,
	code string) (token string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, session_id=%v, code=%v, took=%v, err=%v\n",
			"GetUserToken", sessionID, code, time.Since(begin), err)
	}(time.Now())
	return s.Service.GetUserToken(ctx, sessionID, c, code)
}

func (s *loggingService) ServeHomePage(ctx context.Context, client io.Writer) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeHomePage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeHomePage(ctx, client)
}

func (s *loggingService) ServeErrorPage(ctx context.Context, client io.Writer, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, err=%v, took=%v\n",
			"ServeErrorPage", err, time.Since(begin))
	}(time.Now())
	s.Service.ServeErrorPage(ctx, client, err)
}

func (s *loggingService) ServeSigninPage(ctx context.Context, client io.Writer) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSigninPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSigninPage(ctx, client)
}

func (s *loggingService) ServeTimelinePage(ctx context.Context, client io.Writer,
	c *model.Client, timelineType string, maxID string, sinceID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, timeline_type=%v, max_id=%v, since_id=%v, min_id=%v, took=%v, err=%v\n",
			"ServeTimelinePage", timelineType, maxID, sinceID, minID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeTimelinePage(ctx, client, c, timelineType, maxID, sinceID, minID)
}

func (s *loggingService) ServeThreadPage(ctx context.Context, client io.Writer, c *model.Client, id string, reply bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, reply=%v, took=%v, err=%v\n",
			"ServeThreadPage", id, reply, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeThreadPage(ctx, client, c, id, reply)
}

func (s *loggingService) ServeNotificationPage(ctx context.Context, client io.Writer, c *model.Client, maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, max_id=%v, min_id=%v, took=%v, err=%v\n",
			"ServeNotificationPage", maxID, minID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeNotificationPage(ctx, client, c, maxID, minID)
}

func (s *loggingService) ServeUserPage(ctx context.Context, client io.Writer, c *model.Client, id string, maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, max_id=%v, min_id=%v, took=%v, err=%v\n",
			"ServeUserPage", id, maxID, minID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserPage(ctx, client, c, id, maxID, minID)
}

func (s *loggingService) ServeAboutPage(ctx context.Context, client io.Writer, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeAboutPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeAboutPage(ctx, client, c)
}

func (s *loggingService) ServeEmojiPage(ctx context.Context, client io.Writer, c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeEmojiPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeEmojiPage(ctx, client, c)
}

func (s *loggingService) ServeLikedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeLikedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeLikedByPage(ctx, client, c, id)
}

func (s *loggingService) ServeRetweetedByPage(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeRetweetedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeRetweetedByPage(ctx, client, c, id)
}

func (s *loggingService) ServeSearchPage(ctx context.Context, client io.Writer, c *model.Client, q string, qType string, offset int) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, q=%v, type=%v, offset=%v, took=%v, err=%v\n",
			"ServeSearchPage", q, qType, offset, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSearchPage(ctx, client, c, q, qType, offset)
}

func (s *loggingService) Like(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Like", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Like(ctx, client, c, id)
}

func (s *loggingService) UnLike(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnLike", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnLike(ctx, client, c, id)
}

func (s *loggingService) Retweet(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Retweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Retweet(ctx, client, c, id)
}

func (s *loggingService) UnRetweet(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnRetweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnRetweet(ctx, client, c, id)
}

func (s *loggingService) PostTweet(ctx context.Context, client io.Writer, c *model.Client, content string, replyToID string, format string, visibility string, isNSFW bool, files []*multipart.FileHeader) (id string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, content=%v, reply_to_id=%v, format=%v, visibility=%v, is_nsfw=%v, took=%v, err=%v\n",
			"PostTweet", content, replyToID, format, visibility, isNSFW, time.Since(begin), err)
	}(time.Now())
	return s.Service.PostTweet(ctx, client, c, content, replyToID, format, visibility, isNSFW, files)
}

func (s *loggingService) Follow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Follow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Follow(ctx, client, c, id)
}

func (s *loggingService) UnFollow(ctx context.Context, client io.Writer, c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnFollow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnFollow(ctx, client, c, id)
}
