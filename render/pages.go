package render

import (
	"fmt"
	"net/http"

	"spiderden.org/8b/conf"
	"spiderden.org/masta"
)

const (
	SigninPageTmpl       = "signin.tmpl"
	ErrorPageTmpl        = "error.tmpl"
	NavPageTmpl          = "nav.tmpl"
	RootPageTmpl         = "root.tmpl"
	TimelinePageTmpl     = "timeline.tmpl"
	ListsPageTmpl        = "lists.tmpl"
	ListPageTmpl         = "list.tmpl"
	ThreadPageTmpl       = "thread.tmpl"
	QuickReplyPageTmpl   = "quickreply.tmpl"
	NotificationPageTmpl = "notification.tmpl"
	UserPageTmpl         = "user.tmpl"
	UserSearchPageTmpl   = "usersearch.tmpl"
	AboutPageTmpl        = "about.tmpl"
	EmojiPageTmpl        = "emoji.tmpl"
	LikedByPageTmpl      = "likedby.tmpl"
	RetweetedByPageTmpl  = "retweetedby.tmpl"
	ReactionsPageTmpl    = "reactions.tmpl"
	SearchPageTmpl       = "search.tmpl"
	SettingsPageTmpl     = "settings.tmpl"
	FiltersPageTmpl      = "filters.tmpl"
	MutePageTmpl         = "mute.tmpl"
)

func SigninPage(rctx *Context) error {
	return render(rctx, SigninPageTmpl, nil)
}

func ListsPage(rctx *Context, lists []*masta.List) error {
	data := &ListsData{
		Context: rctx,
		Lists:   lists,
	}

	return render(rctx, ListsPageTmpl, data)
}

func ListPage(rctx *Context, data *ListData) error {
	return render(rctx, ListPageTmpl, data)
}

func NavPage(rctx *Context, user *masta.Account) (err error) {
	rctx.target = "main"

	return render(rctx, NavPageTmpl, &NavData{
		User: user,
	})
}

func RootPage(rctx *Context) (err error) {
	rctx.title = "8bloat"

	return render(rctx, RootPageTmpl, rctx)
}

func ThreadPage(rctx *Context, status *masta.Status, context *masta.Context, mutate bool, src *masta.Source) (err error) {
	var pctx PostContext

	// If we are mutating, and there is no source status, then
	// it is a reply.
	if mutate && src != nil {
		pctx = PostContext{
			DefaultVisibility: status.Visibility,
			DefaultFormat:     rctx.Settings.DefaultFormat,
			Formats:           conf.PostFormats,
			EditContext: &EditContext{
				Source: src,
				Status: status,
			},
		}
	} else if mutate {
		var content string
		var visibility string
		if rctx.UserID != status.Account.ID {
			content += "@" + status.Account.Acct + " "
		}
		for i := range status.Mentions {
			if status.Mentions[i].ID != rctx.UserID &&
				status.Mentions[i].ID != status.Account.ID {
				content += "@" + status.Mentions[i].Acct + " "
			}
		}

		isDirect := status.Visibility == "direct"
		if isDirect || rctx.Settings.CopyScope {
			visibility = status.Visibility
		} else {
			visibility = rctx.Settings.DefaultVisibility
		}

		pctx = PostContext{
			DefaultVisibility: visibility,
			DefaultFormat:     rctx.Settings.DefaultFormat,
			Formats:           conf.PostFormats,
			ReplyContext: &ReplyContext{
				InReplyToID:        status.ID,
				InReplyToName:      status.Account.Acct,
				ReplyContent:       content,
				ReplySubjectHeader: status.SpoilerText,
				ForceVisibility:    isDirect,
			},
		}
	}

	replymap := make(map[masta.ID][]ThreadReplyData)
	nomap := make(map[masta.ID]int)

	statuses := append(append(context.Ancestors, status), context.Descendants...)
	statusdata := make([]*StatusData, len(statuses))

	for i, status := range statuses {
		no := i + 1
		nomap[status.ID] = no

		data := StatusData{
			No:          &no,
			Status:      status,
			Replies:     []ThreadReplyData{},
			ShowReplies: true,
		}

		statusdata[i] = &data

		if replyee := status.InReplyToID; replyee != nil {
			replyee := *replyee
			replydata := ThreadReplyData{
				No: no,
				ID: status.ID,
			}

			_, ok := replymap[replyee]
			if !ok {
				replymap[replyee] = []ThreadReplyData{replydata}
			} else {
				replyarr := replymap[replyee]
				replymap[replyee] = append(replyarr, replydata)
			}
		}

	}

	for i, v := range statusdata {
		v.Replies = replymap[v.Status.ID]

		if id := statuses[i].InReplyToID; id != nil {
			no := nomap[*id]
			v.InReplyToNo = &no
		}
	}

	data := &ThreadData{
		Statuses:    statusdata,
		PostContext: pctx,
	}

	return render(rctx, ThreadPageTmpl, data)
}

// We don't abstract this stuff away that much because
// there's too many transport-level details and too little
// data reshuffling for the templating.
func TimelinePage(rctx *Context, data *TimelineData) error {
	return render(rctx, TimelinePageTmpl, data)
}

func QuickReplyPage(rctx *Context, replyee *masta.Status, parent *masta.Status) (err error) {
	var content string
	if rctx.UserID != replyee.Account.ID {
		content += "@" + replyee.Account.Acct + " "
	}
	for _, mention := range replyee.Mentions {
		if mention.ID != rctx.UserID && mention.ID != replyee.Account.ID {
			content += "@" + mention.Acct + " "
		}
	}

	var visibility string
	isDirect := replyee.Visibility == "direct"
	if isDirect || rctx.Settings.CopyScope {
		visibility = replyee.Visibility
	} else {
		visibility = rctx.Settings.DefaultVisibility
	}

	pctx := PostContext{
		DefaultVisibility: visibility,
		DefaultFormat:     rctx.Settings.DefaultFormat,
		Formats:           conf.PostFormats,
		ReplyContext: &ReplyContext{
			InReplyToID:        replyee.ID,
			InReplyToName:      replyee.Account.Acct,
			QuickReply:         true,
			ReplyContent:       content,
			ReplySubjectHeader: replyee.SpoilerText,
			ForceVisibility:    isDirect,
		},
	}

	data := &QuickReplyData{
		Ancestor:    parent,
		Status:      replyee,
		PostContext: pctx,
	}
	return render(rctx, QuickReplyPageTmpl, data)
}

func LikedByPage(rctx *Context, likers []*masta.Account) (err error) {
	data := &LikedByData{
		Users: likers,
	}
	return render(rctx, LikedByPageTmpl, data)
}

func RetweetedByPage(rctx *Context, retweeters []*masta.Account) (err error) {
	data := &RetweetedByData{
		Users: retweeters,
	}
	return render(rctx, RetweetedByPageTmpl, data)
}

func ReactionsPage(rctx *Context, reactions []masta.EmojiReaction) (err error) {
	data := &ReactionsData{
		Reactions: reactions,
	}

	return render(rctx, ReactionsPageTmpl, data)
}

func NotificationPage(rctx *Context, notifs []*masta.Notification) (err error) {
	data := &NotificationData{
		Notifications: notifs,
	}

	rctx.title = "8b | notifications"
	rctx.refreshInterval = rctx.Settings.NotificationInterval
	rctx.target = "main"

	for _, notif := range notifs {
		if notif != nil && notif.Pleroma != nil && !notif.Pleroma.IsSeen {
			data.UnmarkedCount++
			rctx.count++
		}
	}

	if data.UnmarkedCount > 0 {
		data.ReadID = notifs[0].ID
	}

	if len(notifs) > conf.MaxPagination {
		data.NextLink = "/notifications?max_id=" + notifs[len(notifs)-1].ID
	}

	return render(rctx, NotificationPageTmpl, data)
}

type userPageEntry interface {
	[]*masta.Status | []*masta.Account
}

type userPageType string

const (
	UserPageStatuses  userPageType = "statuses"
	UserPagePinned    userPageType = "pinned"
	UserPageMedia     userPageType = "media"
	UserPageFollowers userPageType = "followers"
	UserPageFollowing userPageType = "following"
	UserPageBookmarks userPageType = "bookmarks"
	UserPageMutes     userPageType = "mutes"
	UserPageBlocks    userPageType = "blocks"
	UserPageLikes     userPageType = "likes"
	UserPageRequests  userPageType = "requests"
)

func UserPage[up userPageEntry](rctx *Context, user *masta.Account, rel *masta.Relationship, pdata up, page userPageType) (err error) {
	data := &UserData{
		User:         user,
		IsCurrent:    (user.ID == rctx.UserID),
		Type:         string(page),
		Relationship: rel,
	}

	if pg := rctx.Pagination; pg != nil {
		data.NextLink = fmt.Sprintf("/user/%s/%s?max_id=%s", page, pg.MinID, pg.MaxID)
	}

	switch d := interface{}(pdata).(type) {
	case []*masta.Status:
		data.Statuses = d
	case []*masta.Account:
		data.Users = d
	}

	return render(rctx, UserPageTmpl, data)
}

func UserSearchPage(rctx *Context, offset int, res *masta.Results, acct *masta.Account, query string) (err error) {
	if len(res.Statuses) == conf.MaxPagination {
		rctx.next = fmt.Sprintf("/usersearch/%s?q=%s&offset=%d", acct.ID, query, offset+conf.MaxPagination)
	}

	data := &UserSearchData{
		User:     acct,
		Q:        query,
		Statuses: res.Statuses,
	}

	return render(rctx, UserSearchPageTmpl, data)
}

func MutePage(rctx *Context, acct *masta.Account) (err error) {
	return render(rctx, MutePageTmpl, &MuteData{
		User: acct,
	})
}

func AboutPage(rctx *Context) (err error) {
	return render(rctx, AboutPageTmpl, nil)
}

func EmojiPage(rctx *Context, ems []*masta.Emoji) (err error) {
	return render(rctx, EmojiPageTmpl, &EmojiData{
		Emojis: ems,
	})
}

func SearchPage(rctx *Context, results *masta.Results, q string, qType string, offset int) (err error) {
	var nextLink string

	if (qType == "accounts" && len(results.Accounts) == 20) ||
		(qType == "statuses" && len(results.Statuses) == 20) {
		offset += 20
		nextLink = fmt.Sprintf("/search?q=%s&type=%s&offset=%d",
			q, qType, offset)
	}

	data := &SearchData{
		Q:        q,
		Type:     qType,
		Users:    results.Accounts,
		Statuses: results.Statuses,
		NextLink: nextLink,
	}
	return render(rctx, SearchPageTmpl, data)
}

func SettingsPage(rctx *Context) (err error) {
	return render(rctx, SettingsPageTmpl, &SettingsData{
		Settings:    &rctx.Settings,
		PostFormats: conf.PostFormats,
	})
}

func FiltersPage(rctx *Context, filters []*masta.Filter) (err error) {
	return render(rctx, FiltersPageTmpl, &FiltersData{
		Filters: filters,
	})
}

func ErrorPage(rctx *Context, err error, retry bool) error {
	var errStr string
	var sessionErr bool
	if err != nil {
		errStr = err.Error()
		if me, ok := err.(*masta.APIError); ok {
			switch me.Code {
			case http.StatusForbidden, http.StatusUnauthorized:
				sessionErr = true
			}
		}
	}

	return render(rctx, ErrorPageTmpl, &ErrorData{
		Context:    rctx,
		Err:        errStr,
		Retry:      retry,
		SessionErr: sessionErr,
	})
}
