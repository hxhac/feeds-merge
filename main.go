package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/actions-go/toolkit/core"
	"gopkg.in/yaml.v3"
)

type env struct {
	path       string
	author     string
	Categories []Categories `yaml:"categories"`
	timeout    int
}

type Feeds struct {
	URL    string `yaml:"url"`
	Name   string `yaml:"name"`
	Remark string `yaml:"remark"`
}

type Categories struct {
	Name  string  `yaml:"name"`
	Feeds []Feeds `yaml:"feeds"`
}

var (
	once sync.Once
	e    *env
)

func newEnv() *env {
	once.Do(func() {
		timeout := core.GetInputOrDefault("CLIENT_TIMEOUT_SECONDS", "30")
		ti, err := strconv.Atoi(timeout)
		if err != nil {
			core.Errorf("set env CLIENT_TIMEOUT_SECONDS error: %v", err)
			return
		}
		path := core.GetInputOrDefault("PATH", "feeds.yml")
		fx, err := os.ReadFile(path)
		if err != nil {
			core.Errorf("Read Config file error: %v", err)
			return
		}
		var cates []Categories
		err = yaml.Unmarshal(fx, &cates)
		if err != nil {
			core.Errorf("Unmarshal: %v", err)
			return
		}
		e = &env{
			path:       path,
			timeout:    ti,
			author:     core.GetInputOrDefault("AUTHOR_NAME", "github-actions"),
			Categories: cates,
		}
	})
	return e
}

func main() {
	// path := core.GetInputOrDefault("path", "feeds.yml")
	//
	// fmt.Printf(`::set-output name=myOutput::%s`, path)

	// LoadConfig()
	// bucket := viper.GetString("s3_bucket")
	// filename := viper.GetString("s3_filename")

	ev := newEnv()
	fmt.Print(ev)

	// combinedFeed := GetAtomFeed()
	// atom, _ := combinedFeed.ToAtom()
	// core.Errorf("Rendered RSS with %v items", len(combinedFeed.Items))

	// if no S3 bucket is defined, simply print the feed on standard output
	// if len(bucket) == 0 {
	// 	fmt.Print(atom)
	// 	return
	// }
	// Upload the feed to S3
	// sess, err := session.NewSession(&aws.Config{})
	// uploader := s3manager.NewUploader(sess)
	// _, err = uploader.Upload(&s3manager.UploadInput{
	// 	Bucket:      aws.String(bucket),
	// 	Key:         aws.String(filename),
	// 	Body:        strings.NewReader(atom),
	// 	ContentType: aws.String("text/xml"),
	// 	ACL:         aws.String("public-read"),
	// })
	// if err != nil {
	// 	log.Fatalf("Unable to upload %q to %q, %v", filename, bucket, err)
	// }
	// log.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}
