package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/eduncan911/podcast"
)

const (
	rumbleHost = "rumble.com"
	dateLayout = "2006-01-02T15:04:05-07:00"

	httpClientTimeout      = 10 * time.Second
	httpServerReadTimeout  = 5 * time.Second
	httpServerWriteTimeout = 300 * time.Second
)

type Request struct {
	Channel     string
	ChannelPath string
}
type Item struct {
	Title           string
	Description     string
	Duration        string
	PublishTime     string
	ThumbnailSrc    string
	Link            string
	IsLiveBroadcast bool
}

func FeedHandler(w http.ResponseWriter, r *http.Request) {

	var req Request

	link := r.URL.Query().Get("link")

	if link == "" {
		http.Error(w, "'link' parameter not found", http.StatusBadRequest)
		return
	}
	url, err := url.Parse(link)
	if err != nil {
		http.Error(w, "could not parse link", http.StatusBadRequest)
		return
	}

	if url.Scheme == "" {
		link = "https://" + link
		url, err = url.Parse(link)
		if err != nil {
			http.Error(w, "could not parse link", http.StatusBadRequest)
			return
		}
	}

	slog.Debug("url", "url", fmt.Sprintf("%#v", url))

	if url.Host != rumbleHost {
		http.Error(w, "link must use host "+rumbleHost, http.StatusBadRequest)
		return
	}

	// Trim anything from link after channel name
	bits := strings.Split(url.Path, "/")
	switch {
	case len(bits) == 2:
		req.Channel = bits[1]
		req.ChannelPath = "/" + bits[1]
	case len(bits) > 2:
		if bits[1] == "c" {
			req.Channel = bits[2]
			req.ChannelPath = strings.Join(bits[:3], "/")
		} else {
			req.Channel = bits[1]
			req.ChannelPath = "/" + bits[1]
		}
	}

	if req.ChannelPath == "" {
		http.Error(w, "channel name could not be found in link", http.StatusBadRequest)
		return
	}

	feed, err := GetFeed(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("there was an error fetching the feed: %s", err), http.StatusBadRequest)
		return
	}

	err = feed.Encode(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetFeed(ctx context.Context, r Request) (*podcast.Podcast, error) {

	ctx2, cancel2 := context.WithTimeout(ctx, httpClientTimeout)
	defer cancel2()

	channelLink := "https://" + rumbleHost + r.ChannelPath
	req, err := http.NewRequestWithContext(ctx2, "GET", channelLink, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("rumble.com returned unexpected status %q", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	items := []Item{}

	feedHeader := doc.Find("div.channel-header--content")
	feedTitle := feedHeader.Find("div.channel-header--title h1").Text()
	feedThumb, _ := feedHeader.Find("div.channel-header--thumb img").Attr("src")

	doc.Find("section.channel-listing__container div.videostream.thumbnail__grid--item").Each(func(i int, s *goquery.Selection) {

		if maxItemCount > 0 && len(items) == maxItemCount {
			return
		}

		item := Item{}
		info := s.Find("div.videostream__info")
		item.Duration = strings.TrimSpace(info.Find(".videostream__status--duration").Text())
		live := info.Find(".videostream__status--live")

		// AFAIK there is no way to flag a podcast as live in iTunes RSS, but may be handy in future
		if len(live.Nodes) > 0 {
			item.IsLiveBroadcast = true
		}

		item.Title = strings.TrimSpace(s.Find("h3.thumbnail__title").Text())
		if item.Title == "" {
			item.Title = "unknown title"
		}
		item.Description = strings.TrimSpace(s.Find("div.videostream__description").Text())
		if item.Description == "" {
			item.Description = "unknown description"
		}

		if maxTextLength > 0 {
			// trim title and description lengths
			if len(item.Title) > maxTextLength {
				item.Title = item.Title[:maxTextLength] + "..."
			}
			if len(item.Description) > maxTextLength {
				item.Description = item.Description[:maxTextLength] + "..."
			}
		}

		publishTimeEl := s.Find("div.videostream__data time")
		item.PublishTime, _ = publishTimeEl.Attr("datetime")

		item.Link = "https://" + rumbleHost
		link := s.Find("a.videostream__link")
		href, _ := link.Attr("href")
		if href != "" {
			item.Link += href
		}

		item.ThumbnailSrc, _ = s.Find("img.thumbnail__image").Attr("src")

		items = append(items, item)
	})

	now := time.Now()

	p := podcast.New(
		feedTitle,
		channelLink,
		"",   // TODO fix empty feed description
		&now, // pubDate
		&now, // lastBuildDate
	)

	if feedThumb != "" {
		p.AddImage(feedThumb)
	}

	for _, i := range items {
		publishTime := time.Time{}
		if err != nil {
			return nil, err
		}
		if i.PublishTime != "" {
			publishTime, err = time.Parse(dateLayout, i.PublishTime)
			if err != nil {
				return nil, err
			}
		}

		item := podcast.Item{
			Title:       i.Title,
			Link:        i.Link,
			Description: i.Description,
			PubDate:     &publishTime,
		}

		if i.Duration != "" {
			duration, err := parseDuration(i.Duration)
			if err != nil {
				// Error is non-fatal, just log
				slog.Error("error parsing duration", "err", err)
			}
			item.AddDuration(int64(duration.Seconds()))
		}

		if i.ThumbnailSrc != "" {
			item.AddImage(i.ThumbnailSrc)
		}

		if _, err := p.AddItem(item); err != nil {
			return nil, fmt.Errorf("error adding item: %w", err)
		}
	}

	slog.Info("feed", "url", req.URL, "item count", len(p.Items))

	return &p, nil
}
