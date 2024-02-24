package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/eduncan911/podcast"
)

const (
	rumbleBaseURL = "https://rumble.com"
	dateLayout    = "2006-01-02T15:04:05-07:00"

	httpClientTimeout      = 10 * time.Second
	httpServerReadTimeout  = 5 * time.Second
	httpServerWriteTimeout = 300 * time.Second

	shutdownTimeout = 10 * time.Second
)

type Request struct {
	Title       string
	Description string
	Link        string
	PublishTime time.Time
	UpdatedTime time.Time
}
type Item struct {
	Title        string
	Description  string
	Duration     string
	PublishTime  string
	ThumbnailSrc string
	Link         string
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	port := flag.Int("port", 8080, "listen on this port")
	flag.Parse()

	http.Handle("/", FeedHandler(ctx))

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  httpServerReadTimeout,
		WriteTimeout: httpServerWriteTimeout,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Printf("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func FeedHandler(ctx context.Context) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var req Request

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Link == "" {
			http.Error(w, fmt.Sprintf("link is required"), http.StatusBadRequest)
			return
		}
		cBits := strings.Split(req.Link, rumbleBaseURL)
		if len(cBits) != 2 {
			http.Error(w, fmt.Sprintf("link must start with "+rumbleBaseURL), http.StatusBadRequest)
			return
		}
		if req.Title == "" {
			req.Title = "unknown title"
		}
		if req.Description == "" {
			req.Description = "unknown description"
		}

		if req.PublishTime.IsZero() {
			req.PublishTime = time.Now()
		}
		if req.UpdatedTime.IsZero() {
			req.UpdatedTime = time.Now()
		}

		feed, err := GetFeed(ctx, req)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = feed.Encode(w)
		if err != nil {
			log.Print(err)
		}
	}
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

	p := podcast.New(
		r.Title,
		r.Link,
		r.Description,
		&r.PublishTime, &r.UpdatedTime,
	)

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
