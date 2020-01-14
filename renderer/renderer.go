package renderer

import (
	"io"
	"strconv"
	"strings"
	"text/template"
	"time"

	"mastodon"
)

var (
	icons = map[string]string{
		"envelope":          "/static/icons/envelope.png",
		"dark-envelope":     "/static/icons/dark-envelope.png",
		"globe":             "/static/icons/globe.png",
		"dark-globe":        "/static/icons/dark-globe.png",
		"liked":             "/static/icons/liked.png",
		"dark-liked":        "/static/icons/liked.png",
		"link":              "/static/icons/link.png",
		"dark-link":         "/static/icons/dark-link.png",
		"lock":              "/static/icons/lock.png",
		"dark-lock":         "/static/icons/dark-lock.png",
		"mail-forward":      "/static/icons/mail-forward.png",
		"dark-mail-forward": "/static/icons/dark-mail-forward.png",
		"reply":             "/static/icons/reply.png",
		"dark-reply":        "/static/icons/dark-reply.png",
		"retweet":           "/static/icons/retweet.png",
		"dark-retweet":      "/static/icons/dark-retweet.png",
		"retweeted":         "/static/icons/retweeted.png",
		"dark-retweeted":    "/static/icons/retweeted.png",
		"smile-o":           "/static/icons/smile-o.png",
		"dark-smile-o":      "/static/icons/dark-smile-o.png",
		"star-o":            "/static/icons/star-o.png",
		"dark-star-o":       "/static/icons/dark-star-o.png",
		"star":              "/static/icons/star.png",
		"dark-star":         "/static/icons/dark-star.png",
		"unlock-alt":        "/static/icons/unlock-alt.png",
		"dark-unlock-alt":   "/static/icons/dark-unlock-alt.png",
		"user-plus":         "/static/icons/user-plus.png",
		"dark-user-plus":    "/static/icons/dark-user-plus.png",
	}
)

type TemplateData struct {
	Data interface{}
	Ctx  *Context
}

type Renderer interface {
	RenderSigninPage(ctx *Context, writer io.Writer, data *SigninData) (err error)
	RenderErrorPage(ctx *Context, writer io.Writer, data *ErrorData)
	RenderTimelinePage(ctx *Context, writer io.Writer, data *TimelineData) (err error)
	RenderThreadPage(ctx *Context, writer io.Writer, data *ThreadData) (err error)
	RenderNotificationPage(ctx *Context, writer io.Writer, data *NotificationData) (err error)
	RenderUserPage(ctx *Context, writer io.Writer, data *UserData) (err error)
	RenderAboutPage(ctx *Context, writer io.Writer, data *AboutData) (err error)
	RenderEmojiPage(ctx *Context, writer io.Writer, data *EmojiData) (err error)
	RenderLikedByPage(ctx *Context, writer io.Writer, data *LikedByData) (err error)
	RenderRetweetedByPage(ctx *Context, writer io.Writer, data *RetweetedByData) (err error)
	RenderFollowingPage(ctx *Context, writer io.Writer, data *FollowingData) (err error)
	RenderFollowersPage(ctx *Context, writer io.Writer, data *FollowersData) (err error)
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
		"FormatTimeRFC3339":       FormatTimeRFC3339,
		"FormatTimeRFC822":        FormatTimeRFC822,
		"GetIcon":                 GetIcon,
		"WithContext":             WithContext,
	}).ParseGlob(templateGlobPattern)
	if err != nil {
		return
	}
	return &renderer{
		template: t,
	}, nil
}

func (r *renderer) RenderSigninPage(ctx *Context, writer io.Writer, signinData *SigninData) (err error) {
	return r.template.ExecuteTemplate(writer, "signin.tmpl", WithContext(signinData, ctx))
}

func (r *renderer) RenderErrorPage(ctx *Context, writer io.Writer, errorData *ErrorData) {
	r.template.ExecuteTemplate(writer, "error.tmpl", WithContext(errorData, ctx))
	return
}

func (r *renderer) RenderTimelinePage(ctx *Context, writer io.Writer, data *TimelineData) (err error) {
	return r.template.ExecuteTemplate(writer, "timeline.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderThreadPage(ctx *Context, writer io.Writer, data *ThreadData) (err error) {
	return r.template.ExecuteTemplate(writer, "thread.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderNotificationPage(ctx *Context, writer io.Writer, data *NotificationData) (err error) {
	return r.template.ExecuteTemplate(writer, "notification.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderUserPage(ctx *Context, writer io.Writer, data *UserData) (err error) {
	return r.template.ExecuteTemplate(writer, "user.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderAboutPage(ctx *Context, writer io.Writer, data *AboutData) (err error) {
	return r.template.ExecuteTemplate(writer, "about.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderEmojiPage(ctx *Context, writer io.Writer, data *EmojiData) (err error) {
	return r.template.ExecuteTemplate(writer, "emoji.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderLikedByPage(ctx *Context, writer io.Writer, data *LikedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "likedby.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderRetweetedByPage(ctx *Context, writer io.Writer, data *RetweetedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "retweetedby.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderFollowingPage(ctx *Context, writer io.Writer, data *FollowingData) (err error) {
	return r.template.ExecuteTemplate(writer, "following.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderFollowersPage(ctx *Context, writer io.Writer, data *FollowersData) (err error) {
	return r.template.ExecuteTemplate(writer, "followers.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderSearchPage(ctx *Context, writer io.Writer, data *SearchData) (err error) {
	return r.template.ExecuteTemplate(writer, "search.tmpl", WithContext(data, ctx))
}

func (r *renderer) RenderSettingsPage(ctx *Context, writer io.Writer, data *SettingsData) (err error) {
	return r.template.ExecuteTemplate(writer, "settings.tmpl", WithContext(data, ctx))
}

func EmojiFilter(content string, emojis []mastodon.Emoji) string {
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", "<img class=\"status-emoji\" src=\""+e.URL+"\" alt=\""+e.ShortCode+"\" title=\""+e.ShortCode+"\" />")
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

func StatusContentFilter(spoiler string, content string, emojis []mastodon.Emoji, mentions []mastodon.Mention) string {
	if len(spoiler) > 0 {
		content = spoiler + "<br />" + content
	}
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", "<img class=\"status-emoji\" src=\""+e.URL+"\" alt=\""+e.ShortCode+"\" title=\""+e.ShortCode+"\" />")
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

func TimeSince(t time.Time) string {
	dur := time.Since(t)

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

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func FormatTimeRFC822(t time.Time) string {
	return t.Format(time.RFC822)
}

func GetIcon(name string, darkMode bool) (icon string) {
	if darkMode {
		name = "dark-" + name
	}
	icon, _ = icons[name]
	return
}

func WithContext(data interface{}, ctx *Context) TemplateData {
	return TemplateData{data, ctx}
}
