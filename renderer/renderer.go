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

type TemplateData struct {
	Data interface{}
	Ctx  *Context
}

type Renderer interface {
	RenderSigninPage(ctx *Context, writer io.Writer, data *SigninData) (err error)
	RenderErrorPage(ctx *Context, writer io.Writer, data *ErrorData)
	RenderRootPage(ctx *Context, writer io.Writer, data *RootData) (err error)
	RenderNavPage(ctx *Context, writer io.Writer, data *NavData) (err error)
	RenderTimelinePage(ctx *Context, writer io.Writer, data *TimelineData) (err error)
	RenderThreadPage(ctx *Context, writer io.Writer, data *ThreadData) (err error)
	RenderNotificationPage(ctx *Context, writer io.Writer, data *NotificationData) (err error)
	RenderUserPage(ctx *Context, writer io.Writer, data *UserData) (err error)
	RenderUserSearchPage(ctx *Context, writer io.Writer, data *UserSearchData) (err error)
	RenderAboutPage(ctx *Context, writer io.Writer, data *AboutData) (err error)
	RenderEmojiPage(ctx *Context, writer io.Writer, data *EmojiData) (err error)
	RenderLikedByPage(ctx *Context, writer io.Writer, data *LikedByData) (err error)
	RenderRetweetedByPage(ctx *Context, writer io.Writer, data *RetweetedByData) (err error)
	RenderSearchPage(ctx *Context, writer io.Writer, data *SearchData) (err error)
	RenderSettingsPage(ctx *Context, writer io.Writer, data *SettingsData) (err error)
}

type renderer struct {
	template *template.Template
}

func NewRenderer(templateGlobPattern string) (r *renderer, err error) {
	t := template.New("default")
	t, err = t.Funcs(template.FuncMap{
		"EmojiFilter":             EmojiFilter,
		"StatusContentFilter":     StatusContentFilter,
		"DisplayInteractionCount": DisplayInteractionCount,
		"TimeSince":               TimeSince,
		"TimeUntil":               TimeUntil,
		"FormatTimeRFC3339":       FormatTimeRFC3339,
		"FormatTimeRFC822":        FormatTimeRFC822,
		"WithContext":             WithContext,
	}).ParseGlob(templateGlobPattern)
	if err != nil {
		return
	}
	return &renderer{
		template: t,
	}, nil
}

func (r *renderer) RenderSigninPage(ctx *Context, writer io.Writer,
	signinData *SigninData) (err error) {
	return r.template.ExecuteTemplate(writer, "signin.tmpl", WithContext(signinData, ctx))
}

func (r *renderer) RenderErrorPage(ctx *Context, writer io.Writer,
	errorData *ErrorData) {
	r.template.ExecuteTemplate(writer, "error.tmpl", WithContext(errorData, ctx))
	return
}

func (r *renderer) RenderNavPage(ctx *Context, writer io.Writer,
	data *NavData) (err error) {
	return r.template.ExecuteTemplate(writer, "nav.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderRootPage(ctx *Context, writer io.Writer,
	data *RootData) (err error) {
	return r.template.ExecuteTemplate(writer, "root.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderTimelinePage(ctx *Context, writer io.Writer,
	data *TimelineData) (err error) {
	return r.template.ExecuteTemplate(writer, "timeline.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderThreadPage(ctx *Context, writer io.Writer,
	data *ThreadData) (err error) {
	return r.template.ExecuteTemplate(writer, "thread.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderNotificationPage(ctx *Context, writer io.Writer,
	data *NotificationData) (err error) {
	return r.template.ExecuteTemplate(writer, "notification.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderUserPage(ctx *Context, writer io.Writer,
	data *UserData) (err error) {
	return r.template.ExecuteTemplate(writer, "user.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderUserSearchPage(ctx *Context, writer io.Writer,
	data *UserSearchData) (err error) {
	return r.template.ExecuteTemplate(writer, "usersearch.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderAboutPage(ctx *Context, writer io.Writer,
	data *AboutData) (err error) {
	return r.template.ExecuteTemplate(writer, "about.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderEmojiPage(ctx *Context, writer io.Writer,
	data *EmojiData) (err error) {
	return r.template.ExecuteTemplate(writer, "emoji.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderLikedByPage(ctx *Context, writer io.Writer,
	data *LikedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "likedby.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderRetweetedByPage(ctx *Context, writer io.Writer,
	data *RetweetedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "retweetedby.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderSearchPage(ctx *Context, writer io.Writer,
	data *SearchData) (err error) {
	return r.template.ExecuteTemplate(writer, "search.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderSettingsPage(ctx *Context, writer io.Writer,
	data *SettingsData) (err error) {
	return r.template.ExecuteTemplate(writer, "settings.tmpl", WithContext(data, ctx))
}

func EmojiFilter(content string, emojis []mastodon.Emoji) string {
	var replacements []string
	var r string
	for _, e := range emojis {
		r = fmt.Sprintf("<img class=\"status-emoji\" src=\"%s\" alt=\"%s\" title=\"%s\" />",
			e.URL, e.ShortCode, e.ShortCode)
		replacements = append(replacements, ":"+e.ShortCode+":", r)
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

func StatusContentFilter(spoiler string, content string,
	emojis []mastodon.Emoji, mentions []mastodon.Mention) string {

	var replacements []string
	var r string
	if len(spoiler) > 0 {
		content = spoiler + "<br />" + content
	}
	for _, e := range emojis {
		r = fmt.Sprintf("<img class=\"status-emoji\" src=\"%s\" alt=\"%s\" title=\"%s\" />",
			e.URL, e.ShortCode, e.ShortCode)
		replacements = append(replacements, ":"+e.ShortCode+":", r)
	}
	for _, m := range mentions {
		replacements = append(replacements, "\""+m.URL+"\"", "\"/user/"+m.ID+"\"")
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

func DisplayInteractionCount(c int64) string {
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
	if m < 60 {
		return strconv.Itoa(int(m)) + "m"
	}
	h := dur.Hours()
	if h < 24 {
		return strconv.Itoa(int(h)) + "h"
	}
	d := h / 24
	if d < 30 {
		return strconv.Itoa(int(d)) + "d"
	}
	mo := d / 30
	if mo < 12 {
		return strconv.Itoa(int(mo)) + "mo"
	}
	y := mo / 12
	return strconv.Itoa(int(y)) + "y"
}

func TimeSince(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	return DurToStr(d)
}

func TimeUntil(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		d = 0
	}
	return DurToStr(d)
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func FormatTimeRFC822(t time.Time) string {
	return t.Format(time.RFC822)
}

func WithContext(data interface{}, ctx *Context) TemplateData {
	return TemplateData{data, ctx}
}
