package render

import (
	"bytes"
	"embed"
	"html/template"
	"regexp"
	"spiderden.org/8b/internal/conf"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"spiderden.org/masta"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var tmpl *template.Template = template.Must(template.New("default").Funcs(
	template.FuncMap{
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
		"wrapRawStatus":           wrapRawStatus,
		"version":                 conf.Version,
		"dbool":                   func(b *bool) bool { return *b },
		"kvf":                     func(k any, v any, f any) kvf { return kvf{k: k, v: v, f: f} }, // triplet tuple
		"themes":                  Themes,
		"themeUIName":             func(name string) string { return themeRegistry[name].UIName },
	}).ParseFS(templateFS, "templates/*.tmpl"),
)

func render(ctx *Context, page string, data interface{}) (err error) {
	return tmpl.ExecuteTemplate(ctx.W, page, withContext(data, ctx))
}

type Page string

type TemplateData struct {
	Data interface{}
	Ctx  *Context
}

type kvf struct {
	k any
	v any
	f any
}

func (k kvf) K() any {
	return k.k
}

func (k kvf) V() any {
	return k.v
}

func (k kvf) F() any {
	return k.f
}

func emojiHTML(e masta.Emoji, height string) string {
	esc := template.HTMLEscapeString
	return `<img class="emoji" src="` + esc(e.URL) + `" alt=":` + esc(e.ShortCode) + `:" title=":` + esc(e.ShortCode) + `:" height="` + esc(height) + `"/>`
}

func emojiFilter(content string, emojis []masta.Emoji) string {
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", emojiHTML(e, "24"))
	}
	return strings.NewReplacer(replacements...).Replace(content)
}

// This is to make it so that links always open in a new tab.
// This isn't meant to be a secure solution, since CSP will
// catch attempts to open in a frame.
// TODO: More granular location detection, to allow hosting under
// a shared domain.
func linkFilter(content string) string {
	node, err := html.Parse(bytes.NewBuffer([]byte(content)))
	if err != nil {
		// This is not for security, just to avoid annoyance.
		// A secure solution would return an error.
		return content
	}

	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			var reli int = -1
			var hrefi int = -1
			var targeti int = -1
			for i, v := range node.Attr {
				if v.Key == "rel" {
					if reli == -1 {
						reli = i
					} else {
						node.Attr[i].Val = ""
					}
				}

				if v.Key == "href" {
					if hrefi == -1 {
						hrefi = i
					} else {
						node.Attr[i].Val = ""
					}
				}

				if v.Key == "target" {
					if hrefi == -1 {
						hrefi = i
					} else {
						node.Attr[i].Val = ""
					}
				}
			}

			if hrefi != -1 {
				href := node.Attr[hrefi].Val
				if !strings.HasPrefix(href, "/") {
					if reli != -1 {
						node.Attr[reli].Val = "noreferer noopener"
					} else {
						node.Attr = append(node.Attr, html.Attribute{Key: "rel", Val: "noreferer noopener"})
					}

					if targeti != -1 {
						node.Attr[targeti].Val = "_blank"
					} else {
						node.Attr = append(node.Attr, html.Attribute{Key: "target", Val: "_blank"})
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(node)

	buf := &bytes.Buffer{}
	err = html.Render(buf, node)
	if err != nil {
		// This is not for security, just to avoid annoyance.
		// A secure solution would return an error.
		return content
	}

	return buf.String()
}

var quoteRE = regexp.MustCompile("(?mU)(^|> *|\n)(&gt;.*)(<br|$)")

func statusContentFilter(content string, emojis []masta.Emoji, mentions []masta.Mention) string {
	content = quoteRE.ReplaceAllString(content, `$1<span class="quote">$2</span>$3`)
	var replacements []string
	for _, e := range emojis {
		replacements = append(replacements, ":"+e.ShortCode+":", emojiHTML(e, "32"))
	}
	for _, m := range mentions {
		replacements = append(replacements, `"`+m.URL+`"`, `"/user/`+m.ID+`" title="@`+m.Acct+`"`)
	}
	return linkFilter(strings.NewReplacer(replacements...).Replace(content))
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

func wrapRawStatus(status *masta.Status) StatusData {
	return StatusData{
		Status: status,
	}
}
