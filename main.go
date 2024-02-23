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
)

const (
	rumbleBaseURL    = "https://rumble.com"
	rumbleChannelURL = "/c/SebGorka/videos"
	dateLayout       = "2006-01-02T15:04:05-07:00"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pubDate := time.Now()
	updatedDate := time.Now()

	p := podcast.New(
		"eduncan911 Podcasts",
		"http://eduncan911.com/",
		"An example Podcast",
		&pubDate, &updatedDate,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", rumbleBaseURL+rumbleChannelURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("section.channel-listing__container div.videostream.thumbnail__grid--item").Each(func(i int, s *goquery.Selection) {

		durationStr := strings.TrimSpace(s.Find("div.videostream__badge").Text())

		duration, err := parseDuration(durationStr)
		if err != nil {
			log.Fatal(err)
		}

		title := strings.TrimSpace(s.Find("h3.thumbnail__title").Text())
		if title == "" {
			title = "unknown title"
		}
		description := strings.TrimSpace(s.Find("div.videostream__description").Text())
		if description == "" {
			description = "unknown description"
		}

		publishTime := time.Time{}
		publishTimeEl := s.Find("div.videostream__data time")
		publishTimeStr, found := publishTimeEl.Attr("datetime")
		if found {
			publishTime, err = time.Parse(dateLayout, publishTimeStr)
			if err != nil {
				log.Fatal(err)
			}
		}

		link := s.Find("a.videostream__link")
		href, _ := link.Attr("href")
		if href == "" {
			href = rumbleBaseURL
		}

		// create an Item
		item := podcast.Item{
			Title:       title,
			Link:        rumbleBaseURL + href,
			Description: description,
			PubDate:     &publishTime,
		}
		item.AddDuration(int64(duration.Seconds()))

		thumbnailSrc, found := s.Find("img.thumbnail__image").Attr("src")
		if found {
			item.AddImage(thumbnailSrc)
		}

		if _, err := p.AddItem(item); err != nil {
			log.Fatalf("error adding item: %q\n", err)
		}

	})

	fmt.Printf("%s\n", p.String())
}
