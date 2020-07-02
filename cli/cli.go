package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/dghubble/go-twitter/twitter"

	"github.com/aksiksi/histweet/lib"
)

const (
	MIN_DAEMON_INTERVAL = 30
)

// CLI args struct
type Args struct {
	// Whether or not to run in daemon mode
	Daemon   bool
	Interval int
	NoPrompt bool
	Archive  string

	// Twitter API key
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string

	// Rule for tweet deletion
	Rule histweet.Rule
}

// Given an age string, converts it into a time-based rule (`RuleTime`)
func ConvertAgeToTime(age string) (time.Time, error) {
	var days int
	var months int
	var years int

	agePat := regexp.MustCompile(`(\d+y)?(\d+m)?(\d+d)?`)

	matches := agePat.FindStringSubmatch(age)
	if matches == nil {
		return time.Time{}, errors.New(fmt.Sprintf("Invalid age string provided: %s", age))
	}

	for _, match := range matches[1:] {
		if match == "" {
			continue
		}

		val, err := strconv.ParseInt(match[:len(match)-1], 10, 32)
		if err != nil {
			return time.Time{}, err
		}

		// The last character of this match must be one of: d, m, or y
		switch match[len(match)-1] {
		case 'd':
			days = int(val)
		case 'm':
			months = int(val)
		case 'y':
			years = int(val)
		default:
			return time.Time{}, errors.New("Invalid age string provided: must only contain \"d\", \"m\", or \"y\"")
		}
	}

	// This is how you go back in time
	now := time.Now().UTC()
	target := now.AddDate(-years, -months, -days)

	return target, nil
}

func runSingle(args *Args, client *twitter.Client) error {
	var tweetIds []int64
	var err error

	if args.Archive == "" {
		// Fetch tweets based on provided rules
		// For now, we assume that user wants to use the timeline API
		tweetIds, err = histweet.FetchTimelineTweets(&args.Rule, client)
		if err != nil {
			return err
		}
	} else {
		tweetIds, err = histweet.FetchArchiveTweets(&args.Rule, args.Archive)
		if err != nil {
			return err
		}
	}

	numTweets := len(tweetIds)

	if numTweets == 0 {
		fmt.Println("\nNo tweets to delete that match the given rule(s).")
		return nil
	}

	// Wait for user to confirm
	if !args.NoPrompt && !args.Daemon {
		fmt.Printf("\nDelete %d tweets that match the above? [y/n] ", numTweets)

		var input string
		fmt.Scanf("%s", &input)
		if input != "y" {
			fmt.Println("Aborting...")
			return nil
		}
	}

	err = histweet.DeleteTweets(tweetIds, client)
	if err != nil {
		return err
	}

	log.Printf("Deleted %d tweets!", numTweets)

	return nil
}

// Run the CLI in daemon mode
// The CLI will continously poll the user's timeline and delete any tweets
// that match the specified rules.
func runDaemon(args *Args, client *twitter.Client) error {
	interval := time.Duration(args.Interval)
	if interval < MIN_DAEMON_INTERVAL {
		return errors.New(fmt.Sprintf("The minimum daemon interal is %d", MIN_DAEMON_INTERVAL))
	}

	ticker := time.NewTicker(interval * time.Second)

	fmt.Printf("\nRunning in daemon mode (interval = %ds)...\n", interval)

	for {
		select {
		case <-ticker.C:
			err := runSingle(args, client)
			if err != nil {
				log.Fatalf("Failed: %s", err.Error())
				return err
			}
		}
	}

	return nil
}

func Run(args *Args) error {
	fmt.Println("\nRules")
	fmt.Println("=====")

	if args.Rule.Tweet != nil {
		before := args.Rule.Tweet.Before
		after := args.Rule.Tweet.After

		if !before.IsZero() && !after.IsZero() {
			fmt.Printf("  * Rule: delete all tweets between %s and %s\n",
				after, before)
		} else if !before.IsZero() {
			fmt.Printf("  * Rule: delete all tweets before %s\n", before)
		} else if !after.IsZero() {
			fmt.Printf("  * Rule: delete all tweets after %s\n", after)
		}

		if args.Rule.Tweet.Contains != "" {
			fmt.Printf("  * Rule: delete all tweets that contain \"%s\"\n", args.Rule.Tweet.Contains)
		}

		if args.Rule.Tweet.Match != nil {
			fmt.Printf("  * Rule: delete all tweets that match regex \"%s\"\n", args.Rule.Tweet.Match)
		}
	}

	if args.Rule.Count != nil {
		fmt.Printf("  * Rule: keep only the latest %d tweets", args.Rule.Count.N)
	}

	client, err := histweet.NewTwitterClient(args.ConsumerKey,
		args.ConsumerSecret,
		args.AccessToken,
		args.AccessSecret,
		true)
	if err != nil {
		return err
	}

	if args.Daemon {
		return runDaemon(args, client)
	} else {
		return runSingle(args, client)
	}
}

// Handles the CLI arguments and calls into the histweet lib to run the command
func handleCli(c *cli.Context) error {
	before := c.Timestamp("before")
	after := c.Timestamp("after")
	age := c.String("age")
	contains := c.String("contains")
	match := c.String("match")
	maxLikes := c.Int("max-likes")
	maxReplies := c.Int("max-replies")
	maxRetweets := c.Int("max-retweets")
	count := c.Int("count")
	archive := c.String("archive")
	invert := c.Bool("invert")
	any := c.Bool("any")
	noPrompt := c.Bool("no-prompt")
	daemon := c.Bool("daemon")
	interval := c.Int("interval")

	isRuleProvided := false

	if !c.IsSet("consumer-key") || !c.IsSet("consumer-secret") ||
		!c.IsSet("access-token") || !c.IsSet("access-secret") {
		return cli.Exit("All Twitter API keys are required", 1)
	}

	// Twitter API info
	consumerKey := c.String("consumer-key")
	consumerSecret := c.String("consumer-secret")
	accessToken := c.String("access-token")
	accessSecret := c.String("access-secret")

	// Pointer to each of the available rule types
	var ruleTweet *histweet.RuleTweet
	var ruleCount *histweet.RuleCount

	if c.IsSet("count") {
		// Count-based rule
		ruleCount = &histweet.RuleCount{
			N: count,
		}

		isRuleProvided = true
	} else if c.IsSet("before") || c.IsSet("after") || c.IsSet("age") || c.IsSet("match") || c.IsSet("contains") {
		// Tweet-based rule
		ruleTweet = &histweet.RuleTweet{}

		// NOTE(aksiksi): Times are interpreted as UTC unless TZ info specified in
		// input date
		if c.IsSet("age") {
			// Ignore before and after args if an age is provided
			log.Println("Age was specified; ignored the before/after arguments")

			resultTime, err := ConvertAgeToTime(age)
			if err != nil {
				return err
			}

			ruleTweet.Before = resultTime
		} else if c.IsSet("before") || c.IsSet("after") {
			if before != nil {
				ruleTweet.Before = (*before).UTC()
			}

			if after != nil {
				ruleTweet.After = (*after).UTC()
			}
		}

		// If we have a match rule, build it
		if c.IsSet("match") {
			pattern, err := regexp.Compile(match)
			if err != nil {
				return cli.Exit("Invalid regex pattern passed into \"match\"", 1)
			}

			ruleTweet.Match = pattern
		}

		if c.IsSet("contains") {
			ruleTweet.Contains = contains
		}

		// Check for other tweet-based rules
		if c.IsSet("max-likes") {
			ruleTweet.MaxLikes = maxLikes
		}

		if c.IsSet("max-replies") {
			ruleTweet.MaxReplies = maxReplies
		}

		if c.IsSet("max-retweets") {
			ruleTweet.MaxRetweets = maxRetweets
		}

		isRuleProvided = true
	}

	// If no rules were provided, let's bail out here
	if !isRuleProvided {
		return cli.Exit("No rules provided... aborting", 1)
	}

	// Build the combined rule
	rule := histweet.Rule{
		Tweet:  ruleTweet,
		Count:  ruleCount,
		Invert: invert,
		Any:    any,
	}

	// Build the args struct to run the command
	args := &Args{
		Daemon:         daemon,
		Interval:       interval,
		NoPrompt:       noPrompt,
		Archive:        archive,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    accessToken,
		AccessSecret:   accessSecret,
		Rule:           rule,
	}

	// Run the command!
	err := Run(args)
	if err != nil {
		return err
	}

	return nil
}
