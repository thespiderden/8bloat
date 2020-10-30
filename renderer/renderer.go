package renderer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
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
	ThreadPage       = "thread.tmpl"
	NotificationPage = "notification.tmpl"
	UserPage         = "user.tmpl"
	UserSearchPage   = "usersearch.tmpl"
	AboutPage        = "about.tmpl"
	EmojiPage        = "emoji.tmpl"
	LikedByPage      = "likedby.tmpl"
	RetweetedByPage  = "retweetedby.tmpl"
	SearchPage       = "search.tmpl"
	SettingsPage     = "settings.tmpl"
)

type TemplateData struct {
	Data interface{}
	Ctx  *Context
}

func emojiFilter(content string, emojis []mastodon.Emoji) string {
	var replacements []string
	var r string
	for _, e := range emojis {
		r = fmt.Sprintf("<img class=\"emoji\" src=\"%s\" alt=\":%s:\" title=\":%s:\" height=\"24\" />",
			e.URL, e.ShortCode, e.ShortCode)
		replacements = append(replacements, ":"+e.ShortCode+":", r)
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

func statusContentFilter(spoiler string, content string,
	emojis []mastodon.Emoji, mentions []mastodon.Mention) string {

	var replacements []string
	var r string
	if len(spoiler) > 0 {
		content = spoiler + "<br />" + content
	}
	for _, e := range emojis {
		r = fmt.Sprintf("<img class=\"emoji\" src=\"%s\" alt=\":%s:\" title=\":%s:\" height=\"32\" />",
			e.URL, e.ShortCode, e.ShortCode)
		replacements = append(replacements, ":"+e.ShortCode+":", r)
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

func DurToStr(dur time.Duration) string {
	s := dur.Seconds()
	if s < 60 {
		return strconv.Itoa(int(s)) + "s"
	}
	m := dur.Minutes()
	if m < 60*2 {
		return strconv.Itoa(int(m)) + "m"
	}
	h := dur.Hours()
	if h < 24*2 {
		return strconv.Itoa(int(h)) + "h"
	}
	d := h / 24
	if d < 30*2 {
		return strconv.Itoa(int(d)) + "d"
	}
	mo := d / 30
	if mo < 12*2 {
		return strconv.Itoa(int(mo)) + "mo"
	}
	y := mo / 12
	return strconv.Itoa(int(y)) + "y"
}

func timeSince(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	return DurToStr(d)
}

func timeUntil(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		d = 0
	}
	return DurToStr(d)
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
