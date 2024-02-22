package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

const (
	rumbleURL  = "https://rumble.com/c/SebGorka/videos"
	dateLayout = "2006-01-02T15:04:05-07:00"
)

func main() {
	feed := &feeds.Feed{
		Title:       "jmoiron.net blog",
		Link:        &feeds.Link{Href: "http://jmoiron.net/blog"},
		Description: "discussion about tech, footie, photos",
		Author:      &feeds.Author{Name: "Jason Moiron", Email: "jmoiron@jmoiron.net"},
		Created:     time.Now(),
	}

	feed.Items = []*feeds.Item{}

	res, err := http.Get(rumbleURL)
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
		href, found := link.Attr("href")

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       "Limiting Concurrency in Go",
			Link:        &feeds.Link{Href: href},
			Description: "A discussion on controlled parallelism in golang",
			Author:      &feeds.Author{Name: "Jason Moiron", Email: "jmoiron@jmoiron.net"},
			Created:     publishTime,
		})

		fmt.Printf("%v href: https://rumble.com%s, %s, %s\n", found, href, duration, publishTime.Local())
	})
}
