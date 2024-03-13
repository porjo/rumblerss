package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/eduncan911/podcast"
	"github.com/labstack/echo/v4"
)

const (
	rumbleBaseURL = "https://rumble.com"
	dateLayout    = "2006-01-02T15:04:05-07:00"

	httpClientTimeout      = 10 * time.Second
	httpServerReadTimeout  = 5 * time.Second
	httpServerWriteTimeout = 300 * time.Second
)

type Request struct {
	Link string
}
type Item struct {
	Title        string
	Description  string
	Duration     string
	PublishTime  string
	ThumbnailSrc string
	Link         string
}

func FeedHandler(c echo.Context) error {

	var req Request

	req.Link = c.QueryParam("link")

	if req.Link == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "link is required")
	}
	cBits := strings.Split(req.Link, rumbleBaseURL)
	if len(cBits) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "link must start with "+rumbleBaseURL)
	}

	feed, err := GetFeed(c.Request().Context(), req)
	if err != nil {
		return err
	}

	err = feed.Encode(c.Response().Writer)
	if err != nil {
		return err
	}

	return nil
}

func GetFeed(ctx context.Context, r Request) (*podcast.Podcast, error) {

	ctx2, cancel2 := context.WithTimeout(ctx, httpClientTimeout)
	defer cancel2()

	req, err := http.NewRequestWithContext(ctx2, "GET", r.Link, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
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

		item := Item{}
		item.Duration = strings.TrimSpace(s.Find("div.videostream__badge").Text())

		item.Title = strings.TrimSpace(s.Find("h3.thumbnail__title").Text())
		if item.Title == "" {
			item.Title = "unknown title"
		}
		item.Description = strings.TrimSpace(s.Find("div.videostream__description").Text())
		if item.Description == "" {
			item.Description = "unknown description"
		}

		publishTimeEl := s.Find("div.videostream__data time")
		item.PublishTime, _ = publishTimeEl.Attr("datetime")

		link := s.Find("a.videostream__link")
		item.Link, _ = link.Attr("href")
		if item.Link == "" {
			item.Link = rumbleBaseURL
		}

		item.ThumbnailSrc, _ = s.Find("img.thumbnail__image").Attr("src")

		items = append(items, item)
	})

	now := time.Now()

	p := podcast.New(
		feedTitle,
		r.Link,
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
				log.Fatal(err)
			}
		}

		item := podcast.Item{
			Title:       i.Title,
			Link:        rumbleBaseURL + i.Link,
			Description: i.Description,
			PubDate:     &publishTime,
		}

		if i.Duration != "" {
			duration, err := parseDuration(i.Duration)
			if err != nil {
				return nil, err
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

	return &p, nil
}
