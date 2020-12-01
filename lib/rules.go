package histweet

import (
	"regexp"
	"time"
)

type ruleComparator int

const (
	comparatorGt ruleComparator = iota
	comparatorGte
	comparatorLt
	comparatorLte
	comparatorEq
	comparatorNeq
)

// RuleCount keeps the N latest tweets.
// If `Latest` is set to `true`, delete the N latest tweets
type RuleCount struct {
	N      int
	Latest bool
}

// RuleTweet checks each Tweet against a set of conditions
type RuleTweet struct {
	Before             time.Time
	After              time.Time
	Match              *regexp.Regexp
	Contains           string
	Likes              int
	LikesComparator    ruleComparator
	Retweets           int
	RetweetsComparator ruleComparator
}

// Rule for what kind of tweets to delete
type Rule struct {
	// Delete tweets based on an account-level count
	Count *RuleCount

	// Delete all tweets that match some tweet-based rules
	Tweet *ParsedRule

	// Raw input rule
	Input string
}
