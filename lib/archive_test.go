package histweet

import (
	"testing"
)

func TestTwitterArchive(t *testing.T) {
	tweetRule, _ := Parse(`likes >= 3 || text ~ "Potato"`)

	tweets, err := FetchArchiveTweets(tweetRule, "sample_archive.js")
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if len(tweets) != 2 {
		t.Error("Error: Expected 2 tweets to match the rule")
	}
}
