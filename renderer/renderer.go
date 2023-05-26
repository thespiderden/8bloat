package renderer

import (
	"html/template"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bloat/mastodon"
)

type Page string

const (
	SigninPage       = "signin.tmpl"
	ErrorPage        = "error.tmpl"
	NavPage          = "nav.tmpl"
	RootPage         = "root.tmpl"
	TimelinePage     = "timeline.tmpl"
	ListsPage        = "lists.tmpl"
	ListPage         = "list.tmpl"
	ThreadPage       = "thread.tmpl"
	QuickReplyPage   = "quickreply.tmpl"
	NotificationPage = "notification.tmpl"
	UserPage         = "user.tmpl"
	UserSearchPage   = "usersearch.tmpl"
	AboutPage        = "about.tmpl"
	EmojiPage        = "emoji.tmpl"
	LikedByPage      = "likedby.tmpl"
	RetweetedByPage  = "retweetedby.tmpl"
	ReactionsPage    = "reactions.tmpl"
	SearchPage       = "search.tmpl"
	SettingsPage     = "settings.tmpl"
	FiltersPage      = "filters.tmpl"
	MutePage         = "mute.tmpl"
)

type TemplateData struct {
	Data interface{}
	Ctx  *Context
}

func emojiHTML(e mastodon.Emoji, height string) string {
	return `<img class="emoji" src="` + e.URL + `" alt=":` + e.ShortCode + `:" title=":` + e.ShortCode + `:" height="` + height + `"/>`
}

func emojiFilter(content string, emojis []mastodon.Emoji) string {
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", emojiHTML(e, "24"))
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

var quoteRE = regexp.MustCompile("(?mU)(^|> *|\n)(&gt;.*)(<br|$)")

func statusContentFilter(content string, emojis []mastodon.Emoji, mentions []mastodon.Mention) string {
	content = quoteRE.ReplaceAllString(content, `$1<span class="quote">$2</span>$3`)
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", emojiHTML(e, "32"))
	}
	for _, m := range mentions {
		replacements = append(replacements, `"`+m.URL+`"`, `"/user/`+m.ID+`" title="@`+m.Acct+`"`)
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

func displayInteractionCount(c int64) string {
	if c > 0 {
		return strconv.Itoa(int(c))
	}
	return ""
}

func durUnit(s int64) (dur int64, unit string) {
	if s < 60 {
		if s < 0 {
			s = 0
		}
		return s, "s"
	}
	m := s / 60
	h := m / 60
	if h < 2 {
		return m, "m"
	}
	d := h / 24
	if d < 2 {
		return h, "h"
	}
	mo := d / 30
	if mo < 2 {
		return d, "d"
	}
	y := d / 365
	if y < 2 {
		return mo, "mo"
	}
	return y, "y"
}

func timeSince(t time.Time) string {
	d, u := durUnit(time.Now().Unix() - t.Unix())
	return strconv.FormatInt(d, 10) + u
}

func timeUntil(t time.Time) string {
	d, u := durUnit(t.Unix() - time.Now().Unix())
	return strconv.FormatInt(d, 10) + u
}

func formatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatTimeRFC822(t time.Time) string {
	return t.Format(time.RFC822)
}

func withContext(data interface{}, ctx *Context) TemplateData {
	return TemplateData{data, ctx}
}

func raw(s string) template.HTML {
	return template.HTML(s)
}

func rawCSS(s string) template.CSS {
	return template.CSS(s)
}

type Renderer interface {
	Render(ctx *Context, writer io.Writer, page string, data interface{}) (err error)
}

type renderer struct {
	template *template.Template
}

func NewRenderer(templateGlobPattern string) (r *renderer, err error) {
	t := template.New("default")
	t, err = t.Funcs(template.FuncMap{
		"EmojiFilter":             emojiFilter,
		"StatusContentFilter":     statusContentFilter,
		"DisplayInteractionCount": displayInteractionCount,
		"TimeSince":               timeSince,
		"TimeUntil":               timeUntil,
		"FormatTimeRFC3339":       formatTimeRFC3339,
		"FormatTimeRFC822":        formatTimeRFC822,
		"WithContext":             withContext,
		"HTML":                    template.HTMLEscapeString,
		"Raw":                     raw,
		"RawCSS":                  rawCSS,
	}).ParseGlob(templateGlobPattern)
	if err != nil {
		return
	}
	return &renderer{
		template: t,
	}, nil
}

func (r *renderer) Render(ctx *Context, writer io.Writer,
	page string, data interface{}) (err error) {
	return r.template.ExecuteTemplate(writer, page, withContext(data, ctx))
}
