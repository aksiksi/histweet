package histweet

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

// Keep the N latest tweets
// If `Latest` is set to `true`, delete the N latest tweets
type RuleCount struct {
	N      int
	Latest bool
}

// Rules that are processed based on tweet content or metadata
type RuleTweet struct {
	Before      time.Time
	After       time.Time
	Contains    *regexp.Regexp
	MaxLikes    int
	MaxReplies  int
	MaxRetweets int
}

// Rule for what kind of tweets to delete
// One or more rules can be defined to narrow down the conditions for tweet
// deletion. However, options within each rule are mutually exclusive.
type Rule struct {
	// Delete tweets based on an account-level count
	Count *RuleCount

	// Delete all tweets that match some tweet-based rules
	Tweet *RuleTweet

	// If true, delete tweets that do _not_ match the specified rules
	Invert bool
}

// Returns `true` if the given tweet matches this rule
func (rule *Rule) IsMatch(tweet *twitter.Tweet) (bool, error) {
	match := true

	createdAt, err := tweet.CreatedAtTime()
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not determine creation time of tweet: %d", tweet.ID))
	}

	// Check if we have a match in tweet-based rules
	if rule.Tweet != nil {
		tweetRule := rule.Tweet

		if !tweetRule.Before.IsZero() {
			match = match && createdAt.Before(tweetRule.Before)
		}

		if !tweetRule.After.IsZero() {
			match = match && createdAt.After(tweetRule.After)
		}

		// Check if we have a contains match
		if tweetRule.Contains != nil {
			res := tweetRule.Contains.FindStringIndex(tweet.Text)
			match = match && (res != nil)
		}

		if tweetRule.MaxLikes > 0 {
			match = match && (tweet.FavoriteCount < tweetRule.MaxLikes)
		}

		if tweetRule.MaxRetweets > 0 {
			match = match && (tweet.RetweetCount < tweetRule.MaxRetweets)
		}

		if tweetRule.MaxReplies > 0 {
			match = match && (tweet.ReplyCount < tweetRule.MaxReplies)
		}
	}

	if rule.Invert {
		match = !match
	}

	return match, nil
}
