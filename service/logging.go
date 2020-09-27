package service

import (
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

func (s *ls) ServeErrorPage(c *model.Client, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, err=%v, took=%v\n",
			"ServeErrorPage", err, time.Since(begin))
	}(time.Now())
	s.Service.ServeErrorPage(c, err)
}

func (s *ls) ServeSigninPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSigninPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSigninPage(c)
}

func (s *ls) ServeRootPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeRootPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeRootPage(c)
}

func (s *ls) ServeNavPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeNavPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeNavPage(c)
}

func (s *ls) ServeTimelinePage(c *model.Client, tType string,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, type=%v, took=%v, err=%v\n",
			"ServeTimelinePage", tType, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeTimelinePage(c, tType, maxID, minID)
}

func (s *ls) ServeThreadPage(c *model.Client, id string,
	reply bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeThreadPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeThreadPage(c, id, reply)
}

func (s *ls) ServeLikedByPage(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeLikedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeLikedByPage(c, id)
}

func (s *ls) ServeRetweetedByPage(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"ServeRetweetedByPage", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeRetweetedByPage(c, id)
}

func (s *ls) ServeNotificationPage(c *model.Client,
	maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeNotificationPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeNotificationPage(c, maxID, minID)
}

func (s *ls) ServeUserPage(c *model.Client, id string,
	pageType string, maxID string, minID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, type=%v, took=%v, err=%v\n",
			"ServeUserPage", id, pageType, time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserPage(c, id, pageType, maxID, minID)
}

func (s *ls) ServeAboutPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeAboutPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeAboutPage(c)
}

func (s *ls) ServeEmojiPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeEmojiPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeEmojiPage(c)
}

func (s *ls) ServeSearchPage(c *model.Client, q string,
	qType string, offset int) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSearchPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSearchPage(c, q, qType, offset)
}

func (s *ls) ServeUserSearchPage(c *model.Client,
	id string, q string, offset int) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeUserSearchPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeUserSearchPage(c, id, q, offset)
}

func (s *ls) ServeSettingsPage(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"ServeSettingsPage", time.Since(begin), err)
	}(time.Now())
	return s.Service.ServeSettingsPage(c)
}

func (s *ls) NewSession(instance string) (redirectUrl string,
	sessionID string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, instance=%v, took=%v, err=%v\n",
			"NewSession", instance, time.Since(begin), err)
	}(time.Now())
	return s.Service.NewSession(instance)
}

func (s *ls) Signin(c *model.Client, sessionID string,
	code string) (token string, userID string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, session_id=%v, took=%v, err=%v\n",
			"Signin", sessionID, time.Since(begin), err)
	}(time.Now())
	return s.Service.Signin(c, sessionID, code)
}

func (s *ls) Signout(c *model.Client) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"Signout", time.Since(begin), err)
	}(time.Now())
	return s.Service.Signout(c)
}

func (s *ls) Post(c *model.Client, content string,
	replyToID string, format string, visibility string, isNSFW bool,
	files []*multipart.FileHeader) (id string, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"Post", time.Since(begin), err)
	}(time.Now())
	return s.Service.Post(c, content, replyToID, format,
		visibility, isNSFW, files)
}

func (s *ls) Like(c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Like", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Like(c, id)
}

func (s *ls) UnLike(c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnLike", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnLike(c, id)
}

func (s *ls) Retweet(c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Retweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Retweet(c, id)
}

func (s *ls) UnRetweet(c *model.Client, id string) (count int64, err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnRetweet", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnRetweet(c, id)
}

func (s *ls) Vote(c *model.Client, id string, choices []string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Vote", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Vote(c, id, choices)
}

func (s *ls) Follow(c *model.Client, id string, reblogs *bool) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Follow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Follow(c, id, reblogs)
}

func (s *ls) UnFollow(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnFollow", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnFollow(c, id)
}

func (s *ls) Mute(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Mute", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Mute(c, id)
}

func (s *ls) UnMute(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnMute", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnMute(c, id)
}

func (s *ls) Block(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Block", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Block(c, id)
}

func (s *ls) UnBlock(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnBlock", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnBlock(c, id)
}

func (s *ls) Subscribe(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Subscribe", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Subscribe(c, id)
}

func (s *ls) UnSubscribe(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnSubscribe", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnSubscribe(c, id)
}

func (s *ls) SaveSettings(c *model.Client, settings *model.Settings) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, took=%v, err=%v\n",
			"SaveSettings", time.Since(begin), err)
	}(time.Now())
	return s.Service.SaveSettings(c, settings)
}

func (s *ls) MuteConversation(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"MuteConversation", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.MuteConversation(c, id)
}

func (s *ls) UnMuteConversation(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnMuteConversation", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnMuteConversation(c, id)
}

func (s *ls) Delete(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Delete", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Delete(c, id)
}

func (s *ls) ReadNotifications(c *model.Client, maxID string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, max_id=%v, took=%v, err=%v\n",
			"ReadNotifications", maxID, time.Since(begin), err)
	}(time.Now())
	return s.Service.ReadNotifications(c, maxID)
}

func (s *ls) Bookmark(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"Bookmark", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.Bookmark(c, id)
}

func (s *ls) UnBookmark(c *model.Client, id string) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf("method=%v, id=%v, took=%v, err=%v\n",
			"UnBookmark", id, time.Since(begin), err)
	}(time.Now())
	return s.Service.UnBookmark(c, id)
}
