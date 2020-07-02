package histweet

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

const (
	ARCHIVE_TIME_LAYOUT = "Mon Jan 02 15:04:05 -0700 2006"
)

// Relevant fields for a tweet in archive JSON format
type archiveTweet struct {
	IdStr            string `json:"id"`
	Id               int64
	CreatedAt        string `json:"created_at"`
	FullText         string `json:"full_text"`
	FavoriteCountStr string `json:"favorite_count"`
	FavoriteCount    int
	RetweetCountStr  string `json:"retweet_count"`
	RetweetCount     int
}

type archiveEntry struct {
	Tweet archiveTweet `json:"tweet"`
}

// Fetch tweets from provided Twitter archive
func FetchArchiveTweets(rule *Rule, archive string) ([]int64, error) {
	var err error
	var f *os.File
	var info os.FileInfo
	var tweetId int64
	var favoriteCount, retweetCount int
	var res bool

	f, err = os.Open(archive)
	if err != nil {
		return nil, err
	}

	// Find file size
	info, err = f.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := info.Size()

	// Skip first 25 characters to get to valid JSON
	// Note: archive file starts with this line:
	// 	  window.YTD.tweet.part0 = [ {
	_, err = f.Seek(25, 0)
	if err != nil {
		return nil, err
	}

	// Allocate buffer of exact size to hold JSON
	buf := make([]byte, fileSize-25)
	_, err = f.Read(buf)
	if err != nil {
		return nil, err
	}

	// Parse tweet archive as JSON
	var tweets []archiveEntry
	err = json.Unmarshal(buf, &tweets)
	if err != nil {
		return nil, err
	}

	log.Printf("Loaded %d tweets from provided archive", len(tweets))

	// Allocate buffer for tweet IDs that need to be deleted
	tweetIds := make([]int64, 0, len(tweets))

	for _, entry := range tweets {
		tweet := &entry.Tweet

		tweetId, _ = strconv.ParseInt(tweet.IdStr, 10, 64)
		tweet.Id = tweetId

		favoriteCount, _ = strconv.Atoi(tweet.FavoriteCountStr)
		tweet.FavoriteCount = favoriteCount

		retweetCount, _ = strconv.Atoi(tweet.RetweetCountStr)
		tweet.RetweetCount = retweetCount

		res, err = rule.IsArchiveMatch(tweet)
		if err != nil {
			return nil, err
		}

		if res {
			tweetIds = append(tweetIds, tweetId)
		}
	}

	return tweetIds, nil
}
