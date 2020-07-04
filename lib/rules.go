package histweet

import (
	"regexp"
	"strings"
	"time"
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
	Match       *regexp.Regexp
	Contains    string
	MaxLikes    int
	MaxRetweets int
}

// Rule for what kind of tweets to delete
type Rule struct {
	// Delete tweets based on an account-level count
	Count *RuleCount

	// Delete all tweets that match some tweet-based rules
	Tweet *RuleTweet

	// If true, delete tweets that do _not_ match the specified rules
	Invert bool

	// If true, delete tweets that match _any_ of the specified rules
	Any bool
}

type Match struct {
	before      bool
	after       bool
	contains    bool
	match       bool
	maxLikes    bool
	maxRetweets bool
}

func (m *Match) Eval(rule *Rule) bool {
	var res bool

	if rule.Any {
		res = m.before || m.after || m.contains || m.match || m.maxLikes || m.maxRetweets
	} else {
		res = m.before && m.after && m.contains && m.match && m.maxLikes && m.maxRetweets
	}

	if rule.Invert {
		return !res
	} else {
		return res
	}
}

// Build a new Match struct
// If `any` is true, init all fields to false
func NewMatch(rule *Rule) *Match {
	if rule.Any {
		return &Match{
			before:      false,
			after:       false,
			contains:    false,
			match:       false,
			maxLikes:    false,
			maxRetweets: false,
		}
	} else {
		return &Match{
			before:      true,
			after:       true,
			contains:    true,
			match:       true,
			maxLikes:    true,
			maxRetweets: true,
		}
	}
}

// Returns `true` if the given tweet matches this rule
func (rule *Rule) IsMatch(tweet *Tweet) (bool, error) {
	m := NewMatch(rule)

	createdAt := tweet.CreatedAt
	text := tweet.Text

	// Check if we have a match in tweet-based rules
	if rule.Tweet != nil {
		tweetRule := rule.Tweet

		if !tweetRule.Before.IsZero() {
			m.before = createdAt.Before(tweetRule.Before)
		}

		if !tweetRule.After.IsZero() {
			m.after = createdAt.After(tweetRule.After)
		}

		if tweetRule.Contains != "" {
			m.contains = strings.Contains(text, tweetRule.Contains)
		}

		if tweetRule.Match != nil {
			m.match = (tweetRule.Match.FindStringIndex(text) != nil)
		}

		if tweetRule.MaxLikes > 0 {
			m.maxLikes = (tweet.NumLikes < tweetRule.MaxLikes)
		}

		if tweetRule.MaxRetweets > 0 {
			m.maxRetweets = (tweet.NumRetweets < tweetRule.MaxRetweets)
		}

		return m.Eval(rule), nil
	}

	return false, nil
}
