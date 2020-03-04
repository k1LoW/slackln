package client

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	dur "github.com/k1LoW/duration"
	"github.com/acaloiaro/slack"
)

var uMentionRe = regexp.MustCompile(`<@U[0-9A-Z]+>`)
var subtMentionRe = regexp.MustCompile(`<\!subteam\^S[0-9A-Z]+\|([^>]+)>`)
var urlRe = regexp.MustCompile(`<(https?://[^>]+)>`)

type Client struct {
	api          *slack.Client
	channelCache map[string]*slack.Channel
	userCache    map[string]*slack.User
	urlCache     string
}

func New(token string) (*Client, error) {
	return &Client{
		api:          slack.New(token),
		channelCache: make(map[string]*slack.Channel),
		userCache:    make(map[string]*slack.User),
	}, nil
}

func (c *Client) GetChannelIDByName(ctx context.Context, channel string) (string, error) {
	if cc, ok := c.channelCache[channel]; ok {
		return cc.ID, nil
	}
	var (
		nc  string
		err error
		cID string
	)
L:
	for {
		var ch []slack.Channel
		p := &slack.GetConversationsParameters{
			Limit:  1000,
			Cursor: nc,
		}
		ch, nc, err = c.api.GetConversationsContext(ctx, p)
		if err != nil {
			return "", err
		}
		for _, cc := range ch {
			if cc.Name == channel {
				cID = cc.ID
				c.channelCache[channel] = &cc
				break L
			}
		}
		if nc == "" {
			break
		}
	}

	return cID, nil
}

func (c *Client) GetHistory(ctx context.Context, msgChan chan slack.Msg, channel, duration, latest, oldest string) error {
	defer close(msgChan)

	var (
		nc string
	)

	cID, err := c.GetChannelIDByName(ctx, channel)
	if err != nil {
		return err
	}

	l, o, err := detectTimeRangeOfMessage(duration, latest, oldest)
	if err != nil {
		return err
	}

	for {
		p := &slack.GetConversationHistoryParameters{
			ChannelID: cID,
			Cursor:    nc,
			Inclusive: false,
			Latest:    l,
			Limit:     1000,
			Oldest:    o,
		}
		res, err := c.api.GetConversationHistoryContext(ctx, p)
		if err != nil {
			return err
		}

		for _, m := range res.Messages {
			msgChan <- m.Msg
		}

		if !res.HasMore {
			break
		}
		nc = res.ResponseMetaData.NextCursor
	}
	return nil
}

func (c *Client) GetUserNameByID(ctx context.Context, uID string) string {
	if uID == "" {
		return ""
	}
	if u, ok := c.userCache[uID]; ok {
		return u.Name
	}
	u, err := c.api.GetUserInfoContext(ctx, uID)
	if err != nil {
		return uID
	}
	c.userCache[uID] = u
	return u.Name
}

func (c *Client) HumanizeMessage(ctx context.Context, in string) string {
	out := strings.Replace(in, "<!here>", "@here", -1)
	out = strings.Replace(out, "<!channel>", "@channel", -1)
	out = uMentionRe.ReplaceAllStringFunc(out, func(in string) string {
		return "@" + c.GetUserNameByID(ctx, strings.Trim(in, "<@>"))
	})
	out = subtMentionRe.ReplaceAllString(out, "$1")
	out = urlRe.ReplaceAllString(out, "$1")
	return out
}

func (c *Client) HumanizeTimestamp(in string) string {
	u := strings.Split(in, ".")
	if len(u) != 2 {
		return in
	}
	b, err := strconv.Atoi(u[0])
	if err != nil {
		return in
	}
	a, err := strconv.Atoi(u[1])
	if err != nil {
		return in
	}
	return time.Unix(int64(b), int64(a)).Format(time.RFC3339Nano)
}

func detectTimeRangeOfMessage(duration, latest, oldest string) (l, o string, err error) {
	var (
		d  time.Duration
		lt time.Time
		ot time.Time
	)

	loc, err := time.LoadLocation("Local")
	if err != nil {
		return "", "", err
	}

	if duration != "" {
		d, err = dur.Parse(duration)
		if err != nil {
			return "", "", err
		}
	}

	switch {
	case latest == "" && oldest == "":
		now := time.Now().UTC()
		lt = now
		ot = now.Add(-d)
	case latest != "" && oldest == "":
		now := time.Now().UTC()
		ll, err := dateparse.ParseFormat(latest)
		if err != nil {
			return "", "", err
		}
		lt, err = time.ParseInLocation(ll, latest, loc)
		if err != nil {
			return "", "", err
		}
		ot = now.Add(-d)
	case latest == "" && oldest != "":
		ol, err := dateparse.ParseFormat(oldest)
		if err != nil {
			return "", "", err
		}
		ot, err = time.ParseInLocation(ol, oldest, loc)
		if err != nil {
			return "", "", err
		}
		lt = ot.Add(d)
	case latest != "" && oldest != "":
		ll, err := dateparse.ParseFormat(latest)
		if err != nil {
			return "", "", err
		}
		lt, err = time.ParseInLocation(ll, latest, loc)
		if err != nil {
			return "", "", err
		}
		ol, err := dateparse.ParseFormat(oldest)
		if err != nil {
			return "", "", err
		}
		ot, err = time.ParseInLocation(ol, oldest, loc)
		if err != nil {
			return "", "", err
		}
	}
	l = fmt.Sprintf("%.6f", float64(lt.UnixNano())/1000000000.0)
	o = fmt.Sprintf("%.6f", float64(ot.UnixNano())/1000000000.0)

	return l, o, err
}

func (c *Client) GetPermalink(ctx context.Context, channel, messageTs string) (string, error) {
	cID, err := c.GetChannelIDByName(ctx, channel)
	if err != nil {
		return "", err
	}
	p := &slack.PermalinkParameters{
		Channel: cID,
		Ts:      messageTs,
	}
	l, err := c.api.GetPermalinkContext(ctx, p)
	if err != nil {
		return "", err
	}
	return l, nil
}

func (c *Client) CreateParmalink(ctx context.Context, channel, messageTs string) (string, error) {
	if c.urlCache == "" {
		l, err := c.GetPermalink(ctx, channel, messageTs)
		if err != nil {
			return "", err
		}
		c.urlCache = l
		return l, err
	}
	cID, err := c.GetChannelIDByName(ctx, channel)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(c.urlCache)
	if err != nil {
		return "", err
	}
	us := u.Scheme + "://" + strings.Join([]string{u.Host, "archives", cID, "p" + strings.Replace(messageTs, ".", "", -1)}, "/")
	return us, nil
}
