package main

import (
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/dghubble/go-twitter/twitter"

	"github.com/aksiksi/histweet/lib"
)

const (
	MIN_DAEMON_INTERVAL = 30
)

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

func runSingle(args *Args, client *twitter.Client) error {
	var tweets []histweet.Tweet
	var err error

	if args.Archive == "" {
		// Fetch tweets based on provided rules
		// For now, we assume that user wants to use the timeline API
		tweets, err = histweet.FetchTimelineTweets(&args.Rule, client)
		if err != nil {
			return err
		}
	} else {
		tweets, err = histweet.FetchArchiveTweets(args.Rule.Tweet, args.Archive)
		if err != nil {
			return err
		}
	}

	numTweets := len(tweets)

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

	err = histweet.DeleteTweets(tweets, client)
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
		return fmt.Errorf("The minimum daemon interal is %d", MIN_DAEMON_INTERVAL)
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

func run(args *Args) error {
	fmt.Println("\nRules")
	fmt.Println("=====")

	// TODO: Print summary of provided rules
	if args.Rule.Tweet != nil {
		fmt.Printf("  * Rule: %s", args.Rule.Input)
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
	count := c.Int("count")
	archive := c.String("archive")
	noPrompt := c.Bool("no-prompt")
	daemon := c.Bool("daemon")
	interval := c.Int("interval")

	var inputRule string

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
	var ruleCount *histweet.RuleCount
	var ruleTweet *histweet.ParsedRule

	if c.Command.HasName("count") {
		// Count-based rule
		ruleCount = &histweet.RuleCount{
			N: count,
		}

		isRuleProvided = true
	} else if c.Command.HasName("rule") {
		if c.Args().Len() == 0 {
			return cli.Exit("Please specify a rule string!", 1)
		}

		inputRule = c.Args().Get(0)

		parser := histweet.NewParser(inputRule)

		// Parse the provided tweet-based rule
		res, err := parser.Parse()
		if err != nil {
			return err
		}

		ruleTweet = res

		isRuleProvided = true
	}

	// If no rules were provided, let's bail out here
	if !isRuleProvided {
		return cli.Exit("No rules provided... aborting", 1)
	}

	// Build the combined rule
	rule := histweet.Rule{
		Tweet: ruleTweet,
		Count: ruleCount,
		Input: inputRule,
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
	err := run(args)
	if err != nil {
		return err
	}

	return nil
}
