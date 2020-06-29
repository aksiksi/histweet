package main

import (
	"log"
	"os"
	"regexp"

	"github.com/urfave/cli/v2"

	"github.com/aksiksi/histweet/lib"
)

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
	args := &histweet.Args{
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
	err := histweet.Run(args)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Define the histweet CLI
	app := &cli.App{
		Name:  "histweet",
		Usage: "Manage your tweets via an intuitive CLI",
		Flags: []cli.Flag{
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
		},
		Action: func(c *cli.Context) error {
			return handleCli(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
