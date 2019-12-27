package renderer

import (
	"context"
	"io"
	"strconv"
	"strings"
	"text/template"
	"time"

	"mastodon"
)

type Renderer interface {
	RenderErrorPage(ctx context.Context, writer io.Writer, data *ErrorData)
	RenderHomePage(ctx context.Context, writer io.Writer, data *HomePageData) (err error)
	RenderSigninPage(ctx context.Context, writer io.Writer, data *SigninData) (err error)
	RenderTimelinePage(ctx context.Context, writer io.Writer, data *TimelineData) (err error)
	RenderThreadPage(ctx context.Context, writer io.Writer, data *ThreadData) (err error)
	RenderNotificationPage(ctx context.Context, writer io.Writer, data *NotificationData) (err error)
	RenderUserPage(ctx context.Context, writer io.Writer, data *UserData) (err error)
	RenderAboutPage(ctx context.Context, writer io.Writer, data *AboutData) (err error)
	RenderEmojiPage(ctx context.Context, writer io.Writer, data *EmojiData) (err error)
	RenderLikedByPage(ctx context.Context, writer io.Writer, data *LikedByData) (err error)
	RenderRetweetedByPage(ctx context.Context, writer io.Writer, data *RetweetedByData) (err error)
	RenderSearchPage(ctx context.Context, writer io.Writer, data *SearchData) (err error)
	RenderSettingsPage(ctx context.Context, writer io.Writer, data *SettingsData) (err error)
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
	}).ParseGlob(templateGlobPattern)
	if err != nil {
		return
	}
	return &renderer{
		template: t,
	}, nil
}

func (r *renderer) RenderErrorPage(ctx context.Context, writer io.Writer, errorData *ErrorData) {
	r.template.ExecuteTemplate(writer, "error.tmpl", errorData)
	return
}

func (r *renderer) RenderHomePage(ctx context.Context, writer io.Writer, homePageData *HomePageData) (err error) {
	return r.template.ExecuteTemplate(writer, "homepage.tmpl", homePageData)
}

func (r *renderer) RenderSigninPage(ctx context.Context, writer io.Writer, signinData *SigninData) (err error) {
	return r.template.ExecuteTemplate(writer, "signin.tmpl", signinData)
}

func (r *renderer) RenderTimelinePage(ctx context.Context, writer io.Writer, data *TimelineData) (err error) {
	return r.template.ExecuteTemplate(writer, "timeline.tmpl", data)
}

func (r *renderer) RenderThreadPage(ctx context.Context, writer io.Writer, data *ThreadData) (err error) {
	return r.template.ExecuteTemplate(writer, "thread.tmpl", data)
}

func (r *renderer) RenderNotificationPage(ctx context.Context, writer io.Writer, data *NotificationData) (err error) {
	return r.template.ExecuteTemplate(writer, "notification.tmpl", data)
}

func (r *renderer) RenderUserPage(ctx context.Context, writer io.Writer, data *UserData) (err error) {
	return r.template.ExecuteTemplate(writer, "user.tmpl", data)
}

func (r *renderer) RenderAboutPage(ctx context.Context, writer io.Writer, data *AboutData) (err error) {
	return r.template.ExecuteTemplate(writer, "about.tmpl", data)
}

func (r *renderer) RenderEmojiPage(ctx context.Context, writer io.Writer, data *EmojiData) (err error) {
	return r.template.ExecuteTemplate(writer, "emoji.tmpl", data)
}

func (r *renderer) RenderLikedByPage(ctx context.Context, writer io.Writer, data *LikedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "likedby.tmpl", data)
}

func (r *renderer) RenderRetweetedByPage(ctx context.Context, writer io.Writer, data *RetweetedByData) (err error) {
	return r.template.ExecuteTemplate(writer, "retweetedby.tmpl", data)
}

func (r *renderer) RenderSearchPage(ctx context.Context, writer io.Writer, data *SearchData) (err error) {
	return r.template.ExecuteTemplate(writer, "search.tmpl", data)
}

func (r *renderer) RenderSettingsPage(ctx context.Context, writer io.Writer, data *SettingsData) (err error) {
	return r.template.ExecuteTemplate(writer, "settings.tmpl", data)
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
