package histweet

import (
	"errors"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const (
	MAX_TIMELINE_TWEETS = 3200
)

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
			return nil, errors.New("Invalid user credentials provided")
		}
	}

	return client, nil
}

// Fetch all timeline tweets for a given user based on the provided `Rule`.
// This function sequentially calls the Twitter user timeline API without any
// throttling.
func FetchTimelineTweets(rule *Rule, client *twitter.Client) ([]int64, error) {
	// TODO: Handle throttling gracefully here
	validCount := 0
	totalCount := 0
	tweets := make([]twitter.Tweet, 0, MAX_TIMELINE_TWEETS)
	var maxId int64 = 0

	timelineParams := &twitter.UserTimelineParams{}

	for {
		if totalCount == MAX_TIMELINE_TWEETS {
			// We've hit the absolute max for this API, so stop here
			break
		}

		timelineParams.MaxID = maxId

		// Fetch a set of tweets (max. 200)
		returnedTweets, _, err := client.Timelines.UserTimeline(timelineParams)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Something went wrong while fetching timeline tweets: %s", err.Error()))
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

				tweetsToCopy := returnedTweets[startIdx:]

				tweets = append(tweets, tweetsToCopy...)
			}
		} else {
			// Figure out if any of these tweets match the given rules
			for i := 0; i < len(returnedTweets); i++ {
				tweet := returnedTweets[i]
				match, _ := rule.IsMatch(&tweet)

				if match {
					tweets = append(tweets, tweet)
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
		maxId = returnedTweets[len(returnedTweets)-1].ID

		totalCount += len(returnedTweets)
	}

	// Build an array of tweet IDs to return for deletion
	tweetIds := make([]int64, len(tweets))
	for i, tweet := range tweets {
		tweetIds[i] = tweet.ID
	}

	return tweetIds, nil
}

func DeleteTweets(tweets []int64, client *twitter.Client) error {
	// TODO: Handle throttling gracefully here
	for _, tweetId := range tweets {
		_, _, err := client.Statuses.Destroy(tweetId, &twitter.StatusDestroyParams{})
		if err != nil {
			return err
		}
	}

	return nil
}
