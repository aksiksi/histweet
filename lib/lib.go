package histweet

import (
	"log"
	"regexp"
	"time"
	// "github.com/dghubble/go-twitter/twitter"
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
}

// CLI args struct
type Args struct {
	// Whether or not to run in daemon mode
	Daemon bool

	// Twitter API key
	ApiKey string

	// Rules for tweet deletion
	Rules *Rules
}

func runSingle(args *Args) error {
	log.Println("Called runSingle")
	return nil
}

func runDaemon(args *Args) error {
	log.Println("Called runDaemon")
	return nil
}

func Run(args *Args) error {
	if args.Daemon {
		return runDaemon(args)
	} else {
		return runSingle(args)
	}
}
