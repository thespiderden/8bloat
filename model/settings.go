package model

type Settings struct {
	DefaultVisibility        string `json:"default_visibility"`
	CopyScope                bool   `json:"copy_scope"`
	ThreadInNewTab           bool   `json:"thread_in_new_tab"`
	HideAttachments          bool   `json:"hide_attachments"`
	MaskNSFW                 bool   `json:"mask_nfsw"`
	AutoRefreshNotifications bool   `json:"auto_refresh_notifications"`
	FluorideMode             bool   `json:"fluoride_mode"`
	DarkMode                 bool   `json:"dark_mode"`
	AntiDopamineMode         bool 	`json:"anti_dopamine_mode"`
}

func NewSettings() *Settings {
	return &Settings{
		DefaultVisibility:        "public",
		CopyScope:                true,
		ThreadInNewTab:           false,
		HideAttachments:          false,
		MaskNSFW:                 true,
		AutoRefreshNotifications: false,
		FluorideMode:             false,
		DarkMode:                 false,
		AntiDopamineMode:         false,
	}
}
