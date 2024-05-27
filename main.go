package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/samber/lo"

	"github.com/actions-go/toolkit/core"

	"gopkg.in/yaml.v3"
)

type env struct {
	path       string
	author     string
	feedLink   string
	categories []Categories `yaml:"categories"`
	timeout    int
	feedLimit  int
}

type Categories struct {
	Type  string `yaml:"type"`
	Feeds []Feed
}

type Feed struct {
	Feed string `yaml:"feed"`
	Des  string `yaml:"des"`
	URL  string `yaml:"url"`
}

var (
	once sync.Once
	wg   sync.WaitGroup
	e    *env
)

func newEnv() *env {
	once.Do(func() {
		ti := EnvStrToInt("INPUT_CLIENT_TIMEOUT", 30)
		feedLimit := EnvStrToInt("INPUT_FEED_LIMIT", 300)
		path := ReadEnv("INPUT_FEEDS_PATH", ".github/workspace/feeds.yml")
		fx, err := os.ReadFile(path)
		if err != nil {
			core.Infof("Read Config file [%s] error: %v", path, err)
			return
		}
		var cates []Categories
		err = yaml.Unmarshal(fx, &cates)
		if err != nil {
			core.Infof("Unmarshal: %v", err)
			return
		}
		e = &env{
			path:       path,
			timeout:    ti,
			feedLimit:  feedLimit,
			author:     ReadEnv("INPUT_AUTHOR_NAME", "github-actions"),
			feedLink:   ReadEnv("INPUT_FEED_LINK", ""),
			categories: cates,
		}
	})
	return e
}

func main() {
	newEnv()

	for _, cate := range e.categories {
		wg.Add(1)
		go func(cate Categories) {
			defer wg.Done()
			feedsTitle := cate.Type
			feeds := cate.Feeds

			// check feed is valid

			urls := lo.Map(feeds, func(item Feed, index int) string {
				return item.Feed
			})

			allFeeds := e.fetchUrls(urls)
			combinedFeed := e.mergeAllFeeds(feedsTitle, allFeeds)
			atom, err := combinedFeed.ToAtom()
			if err != nil {
				core.Infof("Rendere RSS error: %v", err)
				return
			}
			err = os.WriteFile(feedsTitle+".atom", []byte(atom), os.ModePerm)
			if err != nil {
				core.Infof("Write file error: %v", err)
				return
			}
		}(cate)
	}
	wg.Wait()
	os.Exit(0)
}

func ReadEnv(envKey, def string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return def
}

func EnvStrToInt(envKey string, def int) int {
	val := os.Getenv(envKey)
	ti, err := strconv.Atoi(val)
	if err != nil {
		// core.Infof("set env [%s] error: %v", envKey, err)
		return def
	}
	return ti
}
