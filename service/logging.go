package service

import (
	"context"
	"io"
	"log"
	"mastodon"
	"mime/multipart"
	"time"
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

func (s *loggingService) GetUserToken(ctx context.Context, sessionID string, c *mastodon.Client,
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
	c *mastodon.Client, maxID string, sinceID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, max_id=%v, since_id=%v, min_id=%v, took=%v, err=%v\n",
			"ServeTimelinePage", maxID, sinceID, minID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeTimelinePage(ctx, client, c, maxID, sinceID, minID)
}

func (s *loggingService) ServeThreadPage(ctx context.Context, client io.Writer, c *mastodon.Client, id string, reply bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, reply=%v, took=%v, err=%v\n",
			"ServeThreadPage", id, reply, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeThreadPage(ctx, client, c, id, reply)
}

func (s *loggingService) Like(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Like", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Like(ctx, client, c, id)
}

func (s *loggingService) UnLike(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnLike", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnLike(ctx, client, c, id)
}

func (s *loggingService) Retweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Retweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Retweet(ctx, client, c, id)
}

func (s *loggingService) UnRetweet(ctx context.Context, client io.Writer, c *mastodon.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnRetweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnRetweet(ctx, client, c, id)
}

func (s *loggingService) PostTweet(ctx context.Context, client io.Writer, c *mastodon.Client, content string, replyToID string, files []*multipart.FileHeader) (id string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, content=%v, reply_to_id=%v, took=%v, err=%v\n",
			"PostTweet", content, replyToID, time.Since(begin), err)
	}(time.Now())
	return s.Service.PostTweet(ctx, client, c, content, replyToID, files)
}
