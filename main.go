package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/actions-go/toolkit/core"
	"gopkg.in/yaml.v3"
)

type env struct {
	path        string
	author      string
	feedLink    string
	feedsFolder string
	categories  []Categories `yaml:"categories"`
	timeout     int
	feedLimit   int
}

type Categories struct {
	Name  string   `yaml:"name"`
	Feeds []string `yaml:"feeds"`
}

var (
	once sync.Once
	wg   sync.WaitGroup
	e    *env
)

func newEnv() *env {
	once.Do(func() {
		ti := EnvStrToInt("CLIENT_TIMEOUT")
		feedLimit := EnvStrToInt("FEED_LIMIT")
		path := core.GetInputOrDefault("FEEDS_PATH", "feeds.yml")
		fx, err := os.ReadFile(path)
		if err != nil {
			core.Errorf("Read Config file [%s] error: %v", path, err)
			return
		}
		var cates []Categories
		err = yaml.Unmarshal(fx, &cates)
		if err != nil {
			core.Errorf("Unmarshal: %v", err)
			return
		}
		e = &env{
			path:        path,
			timeout:     ti,
			feedLimit:   feedLimit,
			author:      core.GetInputOrDefault("AUTHOR_NAME", ""),
			feedLink:    core.GetInputOrDefault("FEED_LINK", ""),
			feedsFolder: core.GetInputOrDefault("FEEDS_FOLDER", "feeds"),
			categories:  cates,
		}
	})
	return e
}

func main() {
	newEnv()
	if _, err := os.Stat(e.feedsFolder); err != nil {
		if os.IsNotExist(err) {
			// core.Errorf("feeds directory is not exist")
			err = os.Mkdir(e.feedsFolder, os.ModePerm)
			if err != nil {
				core.Errorf("Mkdir feeds error: %v", err)
				return
			}
			return
		}
		// core.Errorf("Stat feeds directory error: %v", err)
		return
	}

	for _, cate := range e.categories {
		wg.Add(1)
		go func(cate Categories) {
			defer wg.Done()
			feedsTitle := cate.Name
			urls := cate.Feeds

			allFeeds := e.fetchUrls(urls)
			combinedFeed := e.mergeAllFeeds(feedsTitle, allFeeds)
			atom, err := combinedFeed.ToAtom()
			if err != nil {
				core.Errorf("Rendere RSS error: %v", err)
				return
			}
			err = os.WriteFile("feeds/"+feedsTitle+".atom", []byte(atom), os.ModePerm)
			if err != nil {
				core.Errorf("Write file error: %v", err)
				return
			}
			core.SetOutput("FEEDS_FOLDER", "feeds")
		}(cate)
	}
	wg.Wait()
	os.Exit(0)
}

func EnvStrToInt(envKey string) int {
	val := core.GetInputOrDefault(envKey, "30")
	ti, err := strconv.Atoi(val)
	if err != nil {
		core.Errorf("set env [%s] error: %v", envKey, err)
	}
	return ti
}
