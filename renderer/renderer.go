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
	RenderErrorPage(ctx context.Context, writer io.Writer, err error)
	RenderHomePage(ctx context.Context, writer io.Writer) (err error)
	RenderSigninPage(ctx context.Context, writer io.Writer) (err error)
	RenderTimelinePage(ctx context.Context, writer io.Writer, data *TimelinePageTemplateData) (err error)
	RenderThreadPage(ctx context.Context, writer io.Writer, data *ThreadPageTemplateData) (err error)
	RenderNotificationPage(ctx context.Context, writer io.Writer, data *NotificationPageTemplateData) (err error)
}

type renderer struct {
	template *template.Template
}

func NewRenderer(templateGlobPattern string) (r *renderer, err error) {
	t := template.New("default")
	t, err = t.Funcs(template.FuncMap{
		"WithEmojis":              WithEmojis,
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

func (r *renderer) RenderErrorPage(ctx context.Context, writer io.Writer, err error) {
	r.template.ExecuteTemplate(writer, "error.tmpl", err)
	return
}

func (r *renderer) RenderHomePage(ctx context.Context, writer io.Writer) (err error) {
	return r.template.ExecuteTemplate(writer, "homepage.tmpl", nil)
}

func (r *renderer) RenderSigninPage(ctx context.Context, writer io.Writer) (err error) {
	return r.template.ExecuteTemplate(writer, "signin.tmpl", nil)
}

func (r *renderer) RenderTimelinePage(ctx context.Context, writer io.Writer, data *TimelinePageTemplateData) (err error) {
	return r.template.ExecuteTemplate(writer, "timeline.tmpl", data)
}

func (r *renderer) RenderThreadPage(ctx context.Context, writer io.Writer, data *ThreadPageTemplateData) (err error) {
	return r.template.ExecuteTemplate(writer, "thread.tmpl", data)
}

func (r *renderer) RenderNotificationPage(ctx context.Context, writer io.Writer, data *NotificationPageTemplateData) (err error) {
	return r.template.ExecuteTemplate(writer, "notification.tmpl", data)
}

func WithEmojis(content string, emojis []mastodon.Emoji) string {
	var emojiNameContentPair []string
	for _, e := range emojis {
		emojiNameContentPair = append(emojiNameContentPair, ":"+e.ShortCode+":", "<img class=\"status-emoji\" src=\""+e.URL+"\" alt=\""+e.ShortCode+"\" />")
	}
	return strings.NewReplacer(emojiNameContentPair...).Replace(content)
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

	y := m / 12
	return strconv.Itoa(int(y)) + "y"
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}
