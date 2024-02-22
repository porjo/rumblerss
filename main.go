package main

import (
	"fmt"

	"github.com/antchfx/htmlquery"
	"github.com/gorilla/feeds"
)

func main() {
	feed := &feeds.Feed{
		Title:       "jmoiron.net blog",
		Link:        &feeds.Link{Href: "http://jmoiron.net/blog"},
		Description: "discussion about tech, footie, photos",
		Author:      &feeds.Author{Name: "Jason Moiron", Email: "jmoiron@jmoiron.net"},
		Created:     now,
	}

	doc, err := htmlquery.LoadURL("https://www.bing.com/search?q=golang")
	if err != nil {
		panic(err)
	}
	// Find all news item.
	list, err := htmlquery.QueryAll(doc, "//ol/li")
	if err != nil {
		panic(err)
	}
	for i, n := range list {
		a := htmlquery.FindOne(n, "//a")
		if a != nil {
			fmt.Printf("%d %s(%s)\n", i, htmlquery.InnerText(a), htmlquery.SelectAttr(a, "href"))
		}
	}

}
