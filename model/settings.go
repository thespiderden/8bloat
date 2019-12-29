package model

type Settings struct {
	DefaultVisibility string `json:"default_visibility"`
	CopyScope         bool   `json:"copy_scope"`
	ThreadInNewTab    bool   `json:"thread_in_new_tab"`
}

func NewSettings() *Settings {
	return &Settings{
		DefaultVisibility: "public",
		CopyScope:         true,
		ThreadInNewTab:    false,
	}
}
