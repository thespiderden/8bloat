package render

// Built-in themes
func init() {
	registerTheme("slate", "Slate", "slate.css")
	registerTheme("slate-dark", "Slate (Dark)", "slate-dark.css")
	registerTheme("foil", "Foil (Experimental)", "foil.css")
	registerTheme("foil-dark", "Foil (Dark) (Experimental)", "foil-dark.css")
}
