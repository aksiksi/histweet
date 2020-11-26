package histweet

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	archiveTimeLayout = "Mon Jan 02 15:04:05 -0700 2006"
	archiveSkipHeader = "window.YTD.tweet.part0 = "
)

// Relevant fields for a tweet in archive JSON format
type archiveTweet struct {
	IDStr            string `json:"id"`
	CreatedAt        string `json:"created_at"`
	FullText         string `json:"full_text"`
	FavoriteCountStr string `json:"favorite_count"`
	RetweetCountStr  string `json:"retweet_count"`
}

type archiveEntry struct {
	Tweet archiveTweet `json:"tweet"`
}

// Convert an archive tweet to internal tweet struct
func convertArchiveTweet(from *archiveTweet) Tweet {
	tweetID, _ := strconv.ParseInt(from.IDStr, 10, 64)
	createdAt, _ := time.Parse(archiveTimeLayout, from.CreatedAt)
	favoriteCount, _ := strconv.Atoi(from.FavoriteCountStr)
	retweetCount, _ := strconv.Atoi(from.RetweetCountStr)

	tweet := Tweet{
		ID:          tweetID,
		CreatedAt:   createdAt,
		Text:        from.FullText,
		NumLikes:    favoriteCount,
		NumRetweets: retweetCount,
		IsRetweet:   from.FullText[:3] == "RT",
		IsReply:     from.FullText[:1] == "@",
	}

	return tweet
}

// FetchArchiveTweets parses all tweets in the provided Twitter archive and
// checks them against the provided Rule. The output is a list of Tweets that
// match the rule (i.e., to be deleted).
func FetchArchiveTweets(rule *ParsedRule, archive string) ([]Tweet, error) {
	var err error
	var f *os.File
	var info os.FileInfo

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
	numSkip := int64(len(archiveSkipHeader))
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
	var origTweets []archiveEntry
	err = json.Unmarshal(buf, &origTweets)
	if err != nil {
		return nil, err
	}

	log.Printf("Loaded %d tweets from provided archive", len(origTweets))

	// Allocate buffer for tweet IDs that need to be deleted
	tweets := make([]Tweet, 0, len(origTweets))

	for _, entry := range origTweets {
		// Convert tweet
		tweet := convertArchiveTweet(&entry.Tweet)

		// If the tweet matches the provided rule, append it to the tweet
		// list
		if rule.IsMatch(&tweet) {
			tweets = append(tweets, tweet)
		}
	}

	// Return the list of tweets to delete
	return tweets, nil
}
