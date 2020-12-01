package histweet

import (
	"fmt"
	"net/http"
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

// Interfaces that wrap the required Twitter API services.
// Defining and using these interfaces allows us to easily mock
// the Twitter client in our tests.
type twitterAccountService interface {
	VerifyCredentials(params *twitter.AccountVerifyParams) (*twitter.User, *http.Response, error)
}

type twitterTimelineService interface {
	UserTimeline(params *twitter.UserTimelineParams) ([]twitter.Tweet, *http.Response, error)
}

type twitterStatusService interface {
	Destroy(id int64, params *twitter.StatusDestroyParams) (*twitter.Tweet, *http.Response, error)
}

// This interface wraps the Twitter client APIs that we use
type twitterClientAPI interface {
	accountService() twitterAccountService
	timelineService() twitterTimelineService
	statusService() twitterStatusService
}

// This is the actual implementation of twitterClientAPI
// Notice that it simply wraps the external Twitter API client
type twitterClient struct {
	client *twitter.Client
}

func (t *twitterClient) accountService() twitterAccountService {
	return t.client.Accounts
}

func (t *twitterClient) timelineService() twitterTimelineService {
	return t.client.Timelines
}

func (t *twitterClient) statusService() twitterStatusService {
	return t.client.Statuses
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

	if rule.Match != nil {
		isMatch = isMatch && (rule.Match.FindStringIndex(tweet.Text) != nil)
	}

	if rule.Likes > 0 {
		match := false

		switch rule.LikesComparator {
		case comparatorGt:
			match = tweet.NumLikes > rule.Likes
		case comparatorGte:
			match = tweet.NumLikes >= rule.Likes
		case comparatorLt:
			match = tweet.NumLikes < rule.Likes
		case comparatorLte:
			match = tweet.NumLikes <= rule.Likes
		case comparatorEq:
			match = tweet.NumLikes == rule.Likes
		case comparatorNeq:
			match = tweet.NumLikes != rule.Likes
		}

		isMatch = isMatch && match
	}

	if rule.Retweets > 0 {
		match := false

		switch rule.RetweetsComparator {
		case comparatorGt:
			match = tweet.NumRetweets > rule.Retweets
		case comparatorGte:
			match = tweet.NumRetweets >= rule.Retweets
		case comparatorLt:
			match = tweet.NumRetweets < rule.Retweets
		case comparatorLte:
			match = tweet.NumRetweets <= rule.Retweets
		case comparatorEq:
			match = tweet.NumRetweets == rule.Retweets
		case comparatorNeq:
			match = tweet.NumRetweets != rule.Retweets
		}

		isMatch = isMatch && match
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
func FetchTimelineTweets(rule *Rule, client twitterClientAPI) ([]Tweet, error) {
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
		returnedTweets, _, err := client.timelineService().UserTimeline(timelineParams)
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
func DeleteTweets(tweets []Tweet, client twitterClientAPI) error {
	// TODO: Handle throttling gracefully here
	for _, tweet := range tweets {
		_, _, err := client.statusService().Destroy(tweet.ID, &twitter.StatusDestroyParams{})
		if err != nil {
			return err
		}
	}

	return nil
}
