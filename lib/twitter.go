package histweet

import (
	"fmt"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const (
	maxTimelineTweets = 3200
)

// Tweet represents a single Twitter tweet
type Tweet struct {
	ID          int64
	CreatedAt   time.Time
	Text        string
	NumLikes    int
	NumRetweets int
	NumReplies  int
	IsRetweet   bool
	IsReply     bool
}

// IsMatch returns true if this tweet matches all set fields in the given rule.
func (tweet *Tweet) IsMatch(rule *RuleTweet) bool {
	if rule == nil {
		return false
	}

	createdAt := tweet.CreatedAt
	isMatch := true

	if !rule.Before.IsZero() {
		isMatch = isMatch && createdAt.Before(rule.Before)
	}

	if !rule.After.IsZero() {
		isMatch = isMatch && createdAt.After(rule.Before)
	}

	if rule.Contains != "" {
		isMatch = isMatch && strings.Contains(tweet.Text, rule.Contains)
	}

	if rule.Match != nil {
		isMatch = isMatch && (rule.Match.FindStringIndex(tweet.Text) != nil)
	}

	if rule.MaxLikes > 0 {
		isMatch = isMatch && (tweet.NumLikes < rule.MaxLikes)
	}

	if rule.MaxRetweets > 0 {
		isMatch = isMatch && (tweet.NumRetweets < rule.MaxRetweets)
	}

	return isMatch
}

// Convert an API tweet to internal tweet struct
func convertAPITweet(from *twitter.Tweet) Tweet {
	createdAt, _ := from.CreatedAtTime()

	tweet := Tweet{
		ID:          from.ID,
		CreatedAt:   createdAt,
		Text:        from.Text,
		NumLikes:    from.FavoriteCount,
		NumRetweets: from.RetweetCount,
		IsRetweet:   from.RetweetedStatus != nil,
		IsReply:     from.InReplyToStatusID != 0,
	}

	return tweet
}

// NewTwitterClient is a helper that builds a Twitter client using
// provided info.
func NewTwitterClient(
	consumerKey string,
	consumerSecret string,
	accessToken string,
	accessSecret string,
	verify bool) (*twitter.Client, error) {
	// Build the Twitter client
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Verify the user
	if verify {
		verifyParams := &twitter.AccountVerifyParams{
			SkipStatus:   twitter.Bool(true),
			IncludeEmail: twitter.Bool(true),
		}

		_, _, err := client.Accounts.VerifyCredentials(verifyParams)
		if err != nil {
			return nil, fmt.Errorf("Invalid user credentials provided")
		}
	}

	return client, nil
}

// FetchTimelineTweets collects all timeline tweets for a given user that match
// the provided `Rule`.
//
// This function sequentially calls the Twitter user timeline API without any
// throttling.
func FetchTimelineTweets(rule *Rule, client *twitter.Client) ([]Tweet, error) {
	// TODO: Handle throttling gracefully here
	validCount := 0
	totalCount := 0
	tweets := make([]Tweet, 0, maxTimelineTweets)
	var maxID int64

	timelineParams := &twitter.UserTimelineParams{}

	for {
		if totalCount == maxTimelineTweets {
			// We've hit the absolute max for this API, so stop here
			break
		}

		timelineParams.MaxID = maxID

		// Fetch a set of tweets (max. 200)
		returnedTweets, _, err := client.Timelines.UserTimeline(timelineParams)
		if err != nil {
			return nil, fmt.Errorf("Something went wrong while fetching timeline tweets: %s", err.Error())
		}

		if rule.Count != nil {
			n := rule.Count.N

			// A count-based rule was provided, so ignore all per-tweet checks
			if (totalCount + len(returnedTweets)) > n {
				// Figure out where to start deleting from in the returned
				// tweets slice
				startIdx := (n - totalCount)
				if startIdx < 0 {
					startIdx = 0
				}

				for _, tweet := range returnedTweets[startIdx:] {
					converted := convertAPITweet(&tweet)
					tweets = append(tweets, converted)
				}
			}
		} else {
			// Figure out if any of these tweets match the given rules
			for _, tweet := range returnedTweets {
				converted := convertAPITweet(&tweet)

				// Evaluate the tweet against the parsed rule.
				// This walks the entire parse tree and ensures that all rules
				// match.
				match := rule.Tweet.Eval(&converted)

				if match {
					tweets = append(tweets, converted)
					validCount++
				}
			}
		}

		if len(returnedTweets) < 200 {
			// We've reached the end, stop here
			break
		}

		// Store the last tweet's maxId to search for older tweets on the next
		// API call
		maxID = returnedTweets[len(returnedTweets)-1].ID

		totalCount += len(returnedTweets)
	}

	return tweets, nil
}

// DeleteTweets deletes the provided list of tweets
func DeleteTweets(tweets []Tweet, client *twitter.Client) error {
	// TODO: Handle throttling gracefully here
	for _, tweet := range tweets {
		_, _, err := client.Statuses.Destroy(tweet.ID, &twitter.StatusDestroyParams{})
		if err != nil {
			return err
		}
	}

	return nil
}
