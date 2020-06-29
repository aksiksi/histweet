package histweet

import (
	"regexp"
	"time"
)

// A time-based rule
// One or both values can be set
type RuleTime struct {
	Before *time.Time
	After  *time.Time
}

type RuleContains struct {
	Pattern *regexp.Regexp
}

// Rules for what kind of tweets to delete
// One or more rules can be defined to narrow down the conditions for tweet
// deletion. However, options within each rule are mutually exclusive.
type Rules struct {
	// Delete tweets based on publication time
	Time *RuleTime

	// Delete all tweets that contain some text
	Contains *RuleContains

	// If true, delete tweets that do not match the specified rules
	Invert bool
}

// CLI args struct
type Args struct {
	// Whether or not to run in daemon mode
	Daemon bool

	// Twitter API key
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string

	// Rules for tweet deletion
	Rules *Rules
}
