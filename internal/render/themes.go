package render

import (
	"embed"
	"errors"
	"io"
	"net/url"
	"spiderden.org/8bloat/internal/conf"
	ptemplate "text/template"
)

//go:embed themes/*.css
var themeFS embed.FS
var themes = ptemplate.Must(
	ptemplate.New("themes").ParseFS(themeFS, "themes/*.css"),
)

type Theme struct {
	template *ptemplate.Template
	Name     string
	UIName   string
}

var themeRegistry = make(map[string]Theme)
var themeList []*Theme

func Themes() []*Theme {
	return themeList
}

func RenderTheme(name string, config conf.Configuration, w io.Writer) error {
	theme, ok := themeRegistry[name]
	if !ok {
		return errors.New("template does not exist")
	}

	return theme.template.Execute(w, themer(config.ClientWebsite))
}

func LookupTheme(name string) (uiName string, ok bool) {
	var t Theme
	t, ok = themeRegistry[name]
	uiName = t.UIName
	return
}

func registerTheme(name string, uiName string, template string) {
	// We test the Theme, and panic if invalid.
	var t themer = "bloat.example.com"
	templ := themes.Lookup(template)
	if templ == nil {
		panic("Theme template does not exist: " + template)
	}

	err := templ.Execute(io.Discard, t)
	if err != nil {
		panic(err)
	}

	theme := Theme{template: templ, Name: name, UIName: uiName}
	themeRegistry[name] = theme
	themeList = append(themeList, &theme)
}

type themer string

func (t themer) Import(theme string) string {
	url, err := url.JoinPath(string(t), "/theme/", theme)
	if err != nil {
		panic(err)
	}

	return `@import url("` + url + `");`
}
