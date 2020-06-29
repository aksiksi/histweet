package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/urfave/cli/v2"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/aksiksi/histweet/lib"
)

// CLI args struct
type Args struct {
	// Whether or not to run in daemon mode
	Daemon   bool
	Interval int
	NoPrompt bool

	// Twitter API key
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string

	// Rule for tweet deletion
	Rule *histweet.Rule
}

// Handles the CLI arguments and calls into the histweet lib to run the command
func handleCli(c *cli.Context) error {
	daemon := c.Bool("daemon")
	interval := c.Int("interval")
	before := c.Timestamp("before")
	after := c.Timestamp("after")
	contains := c.String("contains")
	invert := c.Bool("invert")
	noPrompt := c.Bool("no-prompt")

	// Twitter API info
	consumerKey := c.String("consumer-key")
	consumerSecret := c.String("consumer-secret")
	accessToken := c.String("access-token")
	accessSecret := c.String("access-secret")

	// Pointer to each of the available rule types
	var ruleTime *histweet.RuleTime
	var ruleContains *histweet.RuleContains

	// If we have a contains rule, build it
	if contains != "" {
		pattern, err := regexp.Compile(contains)
		if err != nil {
			return cli.Exit("Invalid regex pattern passed into \"contains\"", 1)
		}

		ruleContains = &histweet.RuleContains{
			Pattern: pattern,
		}
	}

	// If we have a time-based rule, build it
	if before != nil || after != nil {
		ruleTime = &histweet.RuleTime{
			Before: before,
			After:  after,
		}
	}

	// If no rules were provided, let's bail out here
	if ruleTime == nil && ruleContains == nil {
		return cli.Exit("No rules provided... aborting", 1)
	}

	// TODO: Can we have a list of rules?
	// CLI would only support a single rule, but lib can be flexible
	rules := &histweet.Rule{
		Time:     ruleTime,
		Contains: ruleContains,
		Invert:   invert,
	}

	// Build the args struct to run the command
	args := &Args{
		Daemon:         daemon,
		Interval:       interval,
		NoPrompt:       noPrompt,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    accessToken,
		AccessSecret:   accessSecret,
		Rule:           rules,
	}

	// Run the command!
	err := Run(args)
	if err != nil {
		return err
	}

	return nil
}

func runSingle(args *Args, client *twitter.Client) error {
	// Fetch tweets based on provided rules
	// For now, we assume that user wants to use the timeline API
	tweets, err := histweet.FetchTimelineTweets(args.Rule, client)
	if err != nil {
		return err
	}

	numTweets := len(tweets)

	if numTweets > 0 {
		fmt.Printf("Delete %d tweets that match the above? [y/n] ", len(tweets))
	} else {
		fmt.Println("No tweets to delete that match the given rule(s).")
		return nil
	}

	// Wait for user to confirm
	if !args.NoPrompt && !args.Daemon {
		var input string
		fmt.Scanf("%s", &input)
		if input == "n" {
			fmt.Println("Aborting...")
			return nil
		}
	}

	err = histweet.DeleteTweets(tweets, client)
	if err != nil {
		return err
	}

	return nil
}

func runDaemon(args *Args, client *twitter.Client) error {
	return nil
}

func Run(args *Args) error {
	if args.Rule.Time != nil {
		before := args.Rule.Time.Before
		after := args.Rule.Time.After

		if before != nil && after != nil {
			fmt.Printf("Rule: delete all tweets between %s and %s\n",
				after, before)
		} else if before != nil {
			fmt.Printf("Rule: delete all tweets before %s\n", before)
		} else if before != nil {
			fmt.Printf("Rule: delete all tweets after %s\n", after)
		}
	}

	if args.Rule.Contains != nil {
		fmt.Printf("Rule: delete all tweets that contain \"%s\"\n",
			args.Rule.Contains.Pattern)
	}

	// Build the Twitter client
	config := oauth1.NewConfig(args.ConsumerKey, args.ConsumerSecret)
	token := oauth1.NewToken(args.AccessToken, args.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Verify the user
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return errors.New("Invalid user credentials provided")
	}

	if args.Daemon {
		return runDaemon(args, client)
	} else {
		return runSingle(args, client)
	}
}

func main() {
	// Define CLI flags
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "daemon",
			Value: false,
			Usage: "Run the CLI in daemon mode",
		},
		&cli.IntFlag{
			Name:  "interval",
			Value: 30,
			Usage: "Interval at which to check for tweets, in seconds",
		},
		&cli.StringFlag{
			Name:     "consumer-key",
			Usage:    "Twitter API consumer key",
			EnvVars:  []string{"HISTWEET_CONSUMER_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "consumer-secret",
			Usage:    "Twitter API consumer secret key",
			EnvVars:  []string{"HISTWEET_CONSUMER_SECRET"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-token",
			Usage:    "Twitter API access token",
			EnvVars:  []string{"HISTWEET_ACCESS_TOKEN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-secret",
			Usage:    "Twitter API access secret",
			EnvVars:  []string{"HISTWEET_ACCESS_SECRET"},
			Required: true,
		},
		&cli.TimestampFlag{
			Name:        "before",
			Usage:       "Delete all tweets before this time",
			Layout:      "2006-01-02T15:04:05",
			DefaultText: "ignored",
		},
		&cli.TimestampFlag{
			Name:        "after",
			Usage:       "Delete all tweets after this time",
			Layout:      "2006-01-02T15:04:05",
			DefaultText: "ignored",
		},
		&cli.StringFlag{
			Name:        "contains",
			Usage:       "Delete all tweets that match a regex pattern",
			DefaultText: "ignored",
		},
		&cli.BoolFlag{
			Name:  "invert",
			Value: false,
			Usage: "Delete tweets that do _not_ match the specified rules",
		},
		&cli.BoolFlag{
			Name:  "no-prompt",
			Value: false,
			Usage: "Do not prompt user to confirm deletion - ignored in daemon mode",
		},
	}

	// Define the histweet CLI
	app := &cli.App{
		Name:  "histweet",
		Usage: "Manage your tweets via an intuitive CLI",
		Flags: flags,
		Action: func(c *cli.Context) error {
			return handleCli(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
