package model

type Settings struct {
	DefaultVisibility string `json:"default_visibility"`
	CopyScope         bool   `json:"copy_scope"`
	ThreadInNewTab    bool   `json:"thread_in_new_tab"`
	MaskNSFW          bool   `json:"mask_nfsw"`
	FluorideMode      bool   `json:"fluoride_mode"`
	DarkMode          bool   `json:"dark_mode"`
}

func NewSettings() *Settings {
	return &Settings{
		DefaultVisibility: "public",
		CopyScope:         true,
		ThreadInNewTab:    false,
		MaskNSFW:          true,
		FluorideMode:      false,
		DarkMode:          false,
	}
}
