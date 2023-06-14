package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bloat/model"
	"bloat/renderer"

	"spiderden.org/masta"
)

type client struct {
	*masta.Client
	w    http.ResponseWriter
	r    *http.Request
	s    *model.Session
	csrf string
	ctx  context.Context
	rctx *renderer.Context
}

func (c *client) setSession(sess *model.Session) error {
	var sb strings.Builder
	bw := base64.NewEncoder(base64.URLEncoding, &sb)
	err := json.NewEncoder(bw).Encode(sess)
	bw.Close()
	if err != nil {
		return err
	}
	http.SetCookie(c.w, &http.Cookie{
		Name:    "session",
		Value:   sb.String(),
		Expires: time.Now().Add(365 * 24 * time.Hour),
	})
	return nil
}

func (c *client) getSession() (sess *model.Session, err error) {
	cookie, _ := c.r.Cookie("session")
	if cookie == nil {
		return nil, errInvalidSession
	}
	br := base64.NewDecoder(base64.URLEncoding, strings.NewReader(cookie.Value))
	err = json.NewDecoder(br).Decode(&sess)
	return
}

func (c *client) unsetSession() {
	http.SetCookie(c.w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now(),
	})
}

func (c *client) writeJson(data interface{}) error {
	return json.NewEncoder(c.w).Encode(map[string]interface{}{
		"data": data,
	})
}

func (c *client) redirect(url string) {
	c.w.Header().Add("Location", url)
	c.w.WriteHeader(http.StatusFound)
}

func (c *client) authenticate(t int) (err error) {
	csrf := c.r.FormValue("csrf_token")
	ref := c.r.URL.RequestURI()
	defer func() {
		if c.s == nil {
			c.s = &model.Session{
				Settings: *model.NewSettings(),
			}
		}
		c.rctx = &renderer.Context{
			HideAttachments:  c.s.Settings.HideAttachments,
			MaskNSFW:         c.s.Settings.MaskNSFW,
			ThreadInNewTab:   c.s.Settings.ThreadInNewTab,
			FluorideMode:     c.s.Settings.FluorideMode,
			DarkMode:         c.s.Settings.DarkMode,
			CSRFToken:        c.s.CSRFToken,
			UserID:           c.s.UserID,
			AntiDopamineMode: c.s.Settings.AntiDopamineMode,
			UserCSS:          c.s.Settings.CSS,
			Referrer:         ref,
		}
	}()
	if t < SESSION {
		return
	}
	sess, err := c.getSession()
	if err != nil {
		return err
	}
	c.s = sess
	c.Client = masta.NewClient(&masta.Config{
		Server:       "https://" + c.s.Instance,
		ClientID:     c.s.ClientID,
		ClientSecret: c.s.ClientSecret,
		AccessToken:  c.s.AccessToken,
	})
	if t >= CSRF && (len(csrf) < 1 || csrf != c.s.CSRFToken) {
		return errInvalidCSRFToken
	}
	return
}
