package histweet

import (
	"testing"
)

func TestTwitterArchive(t *testing.T) {
	tweetRule, _ := Parse(`likes >= 3 || text ~ "Potato"`)

	var inputs = []struct {
		archive         string
		expectedMatches int
	}{
		// Valid
		{"sample_archive.js", 2},

		// Invalid
		{"junk123.js", -1},
		{"sample_archive_no_size.js", -1},
		{"sample_archive_invalid.js", -1},
	}

	for _, input := range inputs {
		t.Run(input.archive, func(t *testing.T) {
			tweets, err := FetchArchiveTweets(tweetRule, input.archive)
			if err != nil {
				if input.expectedMatches == -1 {
					t.Logf("Invalid archive detected -- %s", err)
					return
				}

				t.Errorf("Failed: %s", err)
			}

			if len(tweets) != input.expectedMatches {
				t.Errorf("Error: Expected %d tweets to match the rule", input.expectedMatches)
			}
		})
	}

}
