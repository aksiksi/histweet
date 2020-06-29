package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/dghubble/go-twitter/twitter"

	"github.com/aksiksi/histweet/lib"
)

const MIN_DAEMON_INTERVAL = 30

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
	Rule histweet.Rule
}

// Given an age string, converts it into a time-based rule (`RuleTime`)
func ConvertAgeToRuleTime(age string) (*histweet.RuleTime, error) {
	var days int
	var months int
	var years int

	agePat := regexp.MustCompile(`(\d+d)?(\d+m)?(\d+y)?`)

	matches := agePat.FindStringSubmatch(age)
	if matches == nil {
		return nil, errors.New(fmt.Sprintf("Invalid age string provided: %s", age))
	}

	for _, match := range matches[1:] {
		if match == "" {
			continue
		}

		val, err := strconv.ParseInt(match[:len(match)-1], 10, 32)
		if err != nil {
			return nil, err
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
			return nil, errors.New("Invalid age string provided: must only contain \"d\", \"m\", or \"y\"")
		}
	}

	// This is how you go back in time
	now := time.Now().UTC()
	target := now.AddDate(-years, -months, -days)

	ruleTime := &histweet.RuleTime{
		Before: target,
	}

	return ruleTime, nil
}

// Handles the CLI arguments and calls into the histweet lib to run the command
func handleCli(c *cli.Context) error {
	daemon := c.Bool("daemon")
	interval := c.Int("interval")
	before := c.Timestamp("before")
	after := c.Timestamp("after")
	contains := c.String("contains")
	age := c.String("age")
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
	// NOTE(aksiksi): Times are interpreted as UTC unless TZ info specified in
	// input arg
	if before != nil || after != nil || age != "" {
		if age == "" {
			ruleTime = &histweet.RuleTime{}

			if before != nil {
				ruleTime.Before = (*before).UTC()
			}

			if after != nil {
				ruleTime.After = (*after).UTC()
			}
		} else {
			// Ignore before and after args if an age is provided
			log.Println("Age was specified; ignored the before/after arguments")

			resultRuleTime, err := ConvertAgeToRuleTime(age)
			if err != nil {
				return err
			}

			ruleTime = resultRuleTime
		}
	}

	// If no rules were provided, let's bail out here
	if ruleTime == nil && ruleContains == nil {
		return cli.Exit("No rules provided... aborting", 1)
	}

	// TODO: Can we have a list of rules?
	// CLI would only support a single rule, but lib can be flexible
	rule := histweet.Rule{
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
		Rule:           rule,
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
	tweets, err := histweet.FetchTimelineTweets(&args.Rule, client)
	if err != nil {
		return err
	}

	numTweets := len(tweets)

	if numTweets == 0 {
		fmt.Println("No tweets to delete that match the given rule(s).")
		return nil
	}

	// Wait for user to confirm
	if !args.NoPrompt && !args.Daemon {
		fmt.Printf("Delete %d tweets that match the above? [y/n] ", len(tweets))

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
		return errors.New(fmt.Sprintf("The minimum daemon interal is %d", MIN_DAEMON_INTERVAL))
	}

	ticker := time.NewTicker(interval * time.Second)

	fmt.Printf("Running in daemon mode (interval = %ds)...\n", interval)

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

	if args.Rule.Time != nil {
		before := args.Rule.Time.Before
		after := args.Rule.Time.After

		if !before.IsZero() && !after.IsZero() {
			fmt.Printf("  * Rule: delete all tweets between %s and %s\n",
				after, before)
		} else if !before.IsZero() {
			fmt.Printf("  * Rule: delete all tweets before %s\n", before)
		} else if !after.IsZero() {
			fmt.Printf("  * Rule: delete all tweets after %s\n", after)
		}
	}

	if args.Rule.Contains != nil {
		fmt.Printf("  * Rule: delete all tweets that contain \"%s\"\n",
			args.Rule.Contains.Pattern)
	}

	fmt.Println()

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

func buildCliApp() *cli.App {
	// Define CLI flags
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "daemon",
			Value: false,
			Usage: "Run the CLI in daemon mode",
		},
		&cli.IntFlag{
			Name:  "interval",
			Value: MIN_DAEMON_INTERVAL,
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
			Usage:       "Delete all tweets before this time (UTC by default)",
			Layout:      "2006-01-02T15:04:05",
			DefaultText: "ignored",
		},
		&cli.TimestampFlag{
			Name:        "after",
			Usage:       "Delete all tweets after this time (UTC by default)",
			Layout:      "2006-01-02T15:04:05",
			DefaultText: "ignored",
		},
		&cli.StringFlag{
			Name:  "age",
			Usage: "Delete all tweets older than this age (ex: 10d, 1m, 1y, 1d6m, 1d3m1y)",
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

	return app
}

func main() {
	app := buildCliApp()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
