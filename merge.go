package main

import (
	"net/http"
	"time"

	"github.com/actions-go/toolkit/core"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

func (e env) fetchUrl(url string, ch chan<- *gofeed.Feed) {
	// core.Infof("Fetching URL: %v\n", url)
	fp := gofeed.NewParser()
	fp.Client = &http.Client{
		Timeout: time.Duration(e.timeout) * time.Second,
	}
	feed, err := fp.ParseURL(url)
	if err == nil {
		ch <- feed
	} else {
		core.Infof("Error on URL [%s]: (%v)", url, err)
		ch <- nil
	}
}

func (e env) fetchUrls(urls []string) []*gofeed.Feed {
	allFeeds := make([]*gofeed.Feed, 0)
	ch := make(chan *gofeed.Feed)
	for _, url := range urls {
		go e.fetchUrl(url, ch)
	}
	for range urls {
		feed := <-ch
		if feed != nil {
			allFeeds = append(allFeeds, feed)
		}
	}
	return allFeeds
}

// TODO: there must be a shorter syntax for this
type byPublished []*gofeed.Feed

func (s byPublished) Len() int {
	return len(s)
}

func (s byPublished) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPublished) Less(i, j int) bool {
	date1 := s[i].Items[0].PublishedParsed
	if date1 == nil {
		date1 = s[i].Items[0].UpdatedParsed
	}
	date2 := s[j].Items[0].PublishedParsed
	if date2 == nil {
		date2 = s[j].Items[0].UpdatedParsed
	}
	return date1.Before(*date2)
}

func (e env) getAuthor(feed *gofeed.Feed) string {
	if feed.Author != nil {
		return feed.Author.Name
	}
	if feed.Items[0].Author != nil {
		return feed.Items[0].Author.Name
	}
	core.Infof("Using Default Author for [%s]", feed.Link)
	return e.author
}

func (e env) mergeAllFeeds(feedTitle string, allFeeds []*gofeed.Feed) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       feedTitle,
		Link:        &feeds.Link{Href: e.feedLink},
		Description: "Merged feeds from " + feedTitle,
		Author: &feeds.Author{
			Name: e.author,
		},
		Created: time.Now(),
	}
	// sort.Sort(sort.Reverse(byPublished(allFeeds)))
	limitPerFeed := e.feedLimit
	seen := make(map[string]bool)
	for _, sourceFeed := range allFeeds {
		for i, item := range sourceFeed.Items {
			if i > limitPerFeed {
				break
			}
			if seen[item.Link] {
				continue
			}
			// created := item.PublishedParsed
			// if created == nil {
			// 	created = item.UpdatedParsed
			// }

			created := GetToday()
			if item.UpdatedParsed != nil {
				created = *item.UpdatedParsed
			}
			if item.PublishedParsed != nil {
				created = *item.PublishedParsed
			}
			if item.UpdatedParsed == nil && item.PublishedParsed == nil {
				created = time.Now()
			}

			feed.Items = append(feed.Items, &feeds.Item{
				Title:       item.Title,
				Link:        &feeds.Link{Href: item.Link},
				Description: item.Description,
				Author:      &feeds.Author{Name: e.getAuthor(sourceFeed)},
				Created:     created,
				Content:     item.Content,
			})
			seen[item.Link] = true
		}
	}
	return feed
}

func GetToday() time.Time {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 00:00:00", time.Local)
	return t
}
