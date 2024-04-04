package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"net/http"
	"spiderden.org/8bloat/internal/conf"
	"strings"
	"time"

	"spiderden.org/8bloat/internal/render"

	"spiderden.org/masta"
)

type Transaction struct {
	*masta.Client
	h       *http.Client
	R       *http.Request
	Conf    *conf.Configuration
	Session *Session
	Rctx    *render.Context
	sfnode  *snowflake.Node
	Ctx     context.Context
	W       http.ResponseWriter
	Vars    map[string]string
	Qry     map[string]string
}

func (c *Transaction) setSession(sess *Session) error {
	var sb strings.Builder
	bw := base64.NewEncoder(base64.URLEncoding, &sb)
	err := json.NewEncoder(bw).Encode(sess)
	bw.Close()
	if err != nil {
		return err
	}
	http.SetCookie(c.W, &http.Cookie{
		Name:    "session",
		Value:   sb.String(),
		Expires: time.Now().Add(365 * 24 * time.Hour),
	})
	return nil
}

func (c *Transaction) getSession() (sess *Session, err error) {
	cookie, _ := c.R.Cookie("session")
	if cookie == nil {
		return nil, errInvalidSession
	}
	br := base64.NewDecoder(base64.URLEncoding, strings.NewReader(cookie.Value))
	err = json.NewDecoder(br).Decode(&sess)
	return
}

func (c *Transaction) unsetSession() {
	c.RevokeToken(c.Ctx)
	http.SetCookie(c.W, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now(),
	})
}

func (c *Transaction) writeJson(data interface{}) error {
	return json.NewEncoder(c.W).Encode(map[string]interface{}{
		"data": data,
	})
}

func (c *Transaction) redirect(url string) {
	c.W.Header().Add("Location", url)
	c.W.WriteHeader(http.StatusFound)
}

type authMode int

const (
	authAnon     authMode = 0
	authSessCSRF authMode = 1
	authSess     authMode = 2
)

func (t *Transaction) authenticate(am authMode) (err error) {
	ref := t.R.URL.RequestURI()
	defer func() {
		if t.Session == nil {
			t.Session = &Session{
				Settings: *render.NewSettings(),
			}
		}
		t.Rctx = &render.Context{
			CSRFToken: t.Session.CSRFToken,
			UserID:    t.Session.UserID,
			Referrer:  ref,
			Settings:  t.Session.Settings,
		}
	}()

	sess, err := t.getSession()
	if err != nil {
		if am == authAnon {
			t.Session = nil
			return nil
		}

		return err
	}

	t.Session = sess

	if am == authAnon {
		return
	}

	t.Session = sess
	t.Client = masta.NewClient(&masta.Config{
		Server:       "https://" + t.Session.Instance,
		ClientID:     t.Session.ClientID,
		ClientSecret: t.Session.ClientSecret,
		AccessToken:  t.Session.AccessToken,
	})

	t.Client.UserAgent = t.Conf.UserAgent
	t.Client.Client = *t.h

	if am != authSessCSRF {
		return
	}

	if token := t.R.FormValue("csrf_token"); token != t.Session.CSRFToken {
		return errInvalidCSRFToken
	}

	return
}

func newSession(t *Transaction, instance string) (rurl string, sess *Session, err error) {
	var instanceURL string
	if strings.HasPrefix(instance, "https://") {
		instanceURL = instance
		instance = strings.TrimPrefix(instance, "https://")
	} else {
		instanceURL = "https://" + instance
	}

	csrf, err := NewCSRFToken()
	if err != nil {
		return
	}

	app, err := masta.RegisterApp(t.Ctx, &masta.AppConfig{
		Client:       *t.h,
		Server:       instanceURL,
		ClientName:   t.Conf.ClientName,
		Scopes:       t.Conf.ClientScope,
		Website:      t.Conf.ClientWebsite,
		RedirectURIs: t.Conf.ClientWebsite + "/oauth_callback",
	})
	if err != nil {
		return
	}
	sess = &Session{
		Instance:     instance,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		CSRFToken:    csrf,
		Settings:     *render.NewSettings(),
	}

	rurl = app.AuthURI
	return
}

type Session struct {
	UserID       string          `json:"uid,omitempty"`
	Instance     string          `json:"ins,omitempty"`
	ClientID     string          `json:"cid,omitempty"`
	ClientSecret string          `json:"cs,omitempty"`
	AccessToken  string          `json:"at,omitempty"`
	CSRFToken    string          `json:"csrf,omitempty"`
	Settings     render.Settings `json:"sett,omitempty"`
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}
