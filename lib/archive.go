package histweet

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	ARCHIVE_TIME_LAYOUT = "Mon Jan 02 15:04:05 -0700 2006"
	ARCHIVE_SKIP_HEADER = "window.YTD.tweet.part0 = "
)

// Relevant fields for a tweet in archive JSON format
type archiveTweet struct {
	IdStr            string `json:"id"`
	CreatedAt        string `json:"created_at"`
	FullText         string `json:"full_text"`
	FavoriteCountStr string `json:"favorite_count"`
	RetweetCountStr  string `json:"retweet_count"`
}

type archiveEntry struct {
	Tweet archiveTweet `json:"tweet"`
}

// Convert an archive tweet to internal tweet struct
func convertArchiveTweet(from *archiveTweet) *Tweet {
	tweetId, _ := strconv.ParseInt(from.IdStr, 10, 64)
	createdAt, _ := time.Parse(ARCHIVE_TIME_LAYOUT, from.CreatedAt)
	favoriteCount, _ := strconv.Atoi(from.FavoriteCountStr)
	retweetCount, _ := strconv.Atoi(from.RetweetCountStr)

	tweet := &Tweet{
		Id:          tweetId,
		CreatedAt:   createdAt,
		Text:        from.FullText,
		NumLikes:    favoriteCount,
		NumRetweets: retweetCount,
		IsRetweet:   from.FullText[:3] == "RT",
		IsReply:     from.FullText[:1] == "@",
	}

	return tweet
}

// Fetch tweets from provided Twitter archive
func FetchArchiveTweets(rule *Rule, archive string) ([]int64, error) {
	var err error
	var f *os.File
	var info os.FileInfo
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

	// Skip some characters to get to valid JSON
	numSkip := int64(len(ARCHIVE_SKIP_HEADER))
	_, err = f.Seek(numSkip, 0)
	if err != nil {
		return nil, err
	}

	// Allocate buffer of exact size to hold JSON
	buf := make([]byte, fileSize-numSkip)
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
		// Convert tweet
		old := &entry.Tweet
		tweet := convertArchiveTweet(old)

		res, err = rule.IsMatch(tweet)
		if err != nil {
			return nil, err
		}

		if res {
			tweetIds = append(tweetIds, tweet.Id)
		}
	}

	return tweetIds, nil
}
