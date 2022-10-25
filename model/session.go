package model

type Session struct {
	ID           string   `json:"id,omitempty"`
	UserID       string   `json:"uid,omitempty"`
	Instance     string   `json:"ins,omitempty"`
	ClientID     string   `json:"cid,omitempty"`
	ClientSecret string   `json:"cs,omitempty"`
	AccessToken  string   `json:"at,omitempty"`
	CSRFToken    string   `json:"csrf,omitempty"`
	Settings     Settings `json:"sett,omitempty"`
}

func (s Session) IsLoggedIn() bool {
	return len(s.AccessToken) > 0
}

type Settings struct {
	DefaultVisibility     string `json:"dv,omitempty"`
	DefaultFormat         string `json:"df,omitempty"`
	CopyScope             bool   `json:"cs,omitempty"`
	ThreadInNewTab        bool   `json:"tnt,omitempty"`
	HideAttachments       bool   `json:"ha,omitempty"`
	MaskNSFW              bool   `json:"mn,omitempty"`
	NotificationInterval  int    `json:"ni,omitempty"`
	FluorideMode          bool   `json:"fm,omitempty"`
	DarkMode              bool   `json:"dm,omitempty"`
	AntiDopamineMode      bool   `json:"adm,omitempty"`
	HideUnsupportedNotifs bool   `json:"hun,omitempty"`
	CSS                   string `json:"css,omitempty"`
}

func NewSettings() *Settings {
	return &Settings{
		DefaultVisibility:     "public",
		DefaultFormat:         "",
		CopyScope:             true,
		ThreadInNewTab:        false,
		HideAttachments:       false,
		MaskNSFW:              true,
		NotificationInterval:  0,
		FluorideMode:          false,
		DarkMode:              false,
		AntiDopamineMode:      false,
		HideUnsupportedNotifs: false,
		CSS:                   "",
	}
}
