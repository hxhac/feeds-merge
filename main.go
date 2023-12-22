package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

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
		ti := EnvStrToInt("CLIENT_TIMEOUT", 30)
		feedLimit := EnvStrToInt("FEED_LIMIT", 300)
		path := ReadEnv("FEEDS_PATH", "feeds.yml")
		fx, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Read Config file [%s] error: %v", path, err)
			return
		}
		var cates []Categories
		err = yaml.Unmarshal(fx, &cates)
		if err != nil {
			fmt.Printf("Unmarshal: %v", err)
			return
		}
		e = &env{
			path:        path,
			timeout:     ti,
			feedLimit:   feedLimit,
			author:      ReadEnv("AUTHOR_NAME", "github-actions"),
			feedLink:    ReadEnv("FEED_LINK", ""),
			feedsFolder: ReadEnv("FEEDS_FOLDER", "feeds"),
			categories:  cates,
		}
	})
	return e
}

func main() {
	newEnv()
	if _, err := os.Stat(e.feedsFolder); err != nil {
		if os.IsNotExist(err) {
			// fmt.Printf("feeds directory is not exist")
			err = os.Mkdir(e.feedsFolder, os.ModePerm)
			if err != nil {
				fmt.Printf("Mkdir feeds error: %v", err)
				return
			}
			return
		}
		// fmt.Printf("Stat feeds directory error: %v", err)
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
				fmt.Printf("Rendere RSS error: %v", err)
				return
			}
			err = os.WriteFile("feeds/"+feedsTitle+".atom", []byte(atom), os.ModePerm)
			if err != nil {
				fmt.Printf("Write file error: %v", err)
				return
			}
			// core.SetOutput("FEEDS_FOLDER", "feeds")
			fmt.Printf(`::set-output name=FEEDS_FOLDER::%s`, e.feedsFolder)
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
		// fmt.Printf("set env [%s] error: %v", envKey, err)
		return def
	}
	return ti
}
