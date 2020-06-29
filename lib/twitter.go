package histweet

import (
	"errors"
	"fmt"
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Builds a new Twitter client with given args
func newTwitterClient(args *Args) (*twitter.Client, error) {
	config := oauth1.NewConfig(args.ConsumerKey, args.ConsumerSecret)
	token := oauth1.NewToken(args.AccessToken, args.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify the user
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, errors.New("Invalid user credentials provided")
	}

	log.Println("Verified user!")

	return client, nil
}

// Fetch all timeline tweets for a given user
// This function will stop once any of the time-based rules are met
// This function sequentially calls the Twitter user timeline API without any
// throttling
func fetchTimelineTweets(rule *Rule, client *twitter.Client) ([]twitter.Tweet, error) {
	validCount := 0
	totalCount := 0
	var tweets []twitter.Tweet
	var maxId int64 = 0
	var timelineParams *twitter.UserTimelineParams

	for {
		if totalCount == 3200 {
			// We've hit the absolute max for this API, stop here
			break
		}

		if maxId == 0 {
			timelineParams = &twitter.UserTimelineParams{}
		} else {
			timelineParams = &twitter.UserTimelineParams{
				MaxID: maxId,
			}
		}

		// Fetch a set of tweets (max. 200)
		returnedTweets, _, err := client.Timelines.UserTimeline(timelineParams)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Something went wrong while fetching timeline tweets: %s", err.Error()))
		}

		// Figure out if any of these tweets match the given rules
		for i := 0; i < len(returnedTweets); i++ {
			tweet := returnedTweets[i]
			match, _ := rule.IsMatch(&tweet)

			if match {
				tweets = append(tweets, tweet)
				validCount = validCount + 1
			}
		}

		if len(returnedTweets) < 200 {
			// We've reached the end, stop here
			break
		}

		// Store the last tweet's maxId to search for older tweets on the next
		// API call
		maxId = returnedTweets[len(returnedTweets)-1].ID

		totalCount = totalCount + len(returnedTweets)
	}

	return tweets, nil
}

func deleteTweets(tweets []twitter.Tweet, client *twitter.Client) error {
	for i := 0; i < len(tweets); i++ {
		_, _, err := client.Statuses.Destroy(tweets[i].ID, &twitter.StatusDestroyParams{})
		if err != nil {
			return err
		}
	}

	return nil
}
