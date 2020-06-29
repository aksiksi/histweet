package histweet

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

// A time-based rule
// One or both values can be set
type RuleTime struct {
	Before time.Time
	After  time.Time
}

// Keep the N latest tweets
// If `Latest` is set to `false`, delete the N latest tweets
type RuleCount struct {
	N      int
	Latest bool
}

type RuleContains struct {
	Pattern *regexp.Regexp
}

// Rule for what kind of tweets to delete
// One or more rules can be defined to narrow down the conditions for tweet
// deletion. However, options within each rule are mutually exclusive.
type Rule struct {
	// Delete tweets based on publication time
	Time *RuleTime

	// Delete tweets based on a fixed count
	Count *RuleCount

	// Delete all tweets that contain some text
	Contains *RuleContains

	// If true, delete tweets that do not match the specified rules
	Invert bool
}

// Returns `true` if the given tweet matches this rule
func (rule *Rule) IsMatch(tweet *twitter.Tweet) (bool, error) {
	match := true

	createdAt, err := tweet.CreatedAtTime()
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not determine creation time of tweet: %d", tweet.ID))
	}

	// Check if we have a match in time-based rules
	if rule.Time != nil {
		if !rule.Time.Before.IsZero() {
			match = match && createdAt.Before(rule.Time.Before)
		}

		if !rule.Time.After.IsZero() {
			match = match && createdAt.After(rule.Time.After)
		}
	}

	// Check if we have a contains match
	if rule.Contains != nil {
		res := rule.Contains.Pattern.FindStringIndex(tweet.Text)
		match = match && (res != nil)
	}

	return match, nil
}
