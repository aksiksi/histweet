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
	before := c.Timestamp("before")
	after := c.Timestamp("after")
	contains := c.String("contains")
	invert := c.Bool("invert")

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
		pattern, err := regexp.Compile(c.String("contains"))
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
	rules := &histweet.Rules{
		Time:     ruleTime,
		Contains: ruleContains,
		Invert:   invert,
	}

	// Build the args struct to run the command
	args := &histweet.Args{
		Daemon:         daemon,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    accessToken,
		AccessSecret:   accessSecret,
		Rules:          rules,
	}

	// Run the command!
	err := histweet.Run(args)
	if err != nil {
		return cli.Exit("Failed to run the CLI", 1)
	}

	return nil
}

func main() {
	// Define the histweet CLI
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "daemon",
				Value: false,
				Usage: "Run the CLI in daemon mode",
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
				Usage:    "Twitter API acccess token",
				EnvVars:  []string{"HISTWEET_ACCESS_TOKEN"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "access-secret",
				Usage:    "Twitter API acccess secret",
				EnvVars:  []string{"HISTWEET_ACCESS_SECRET"},
				Required: true,
			},
			&cli.TimestampFlag{
				Name:        "before",
				Usage:       "Delete all tweets before this time.",
				Layout:      "2006-01-02T15:04:05",
				DefaultText: "ignored",
			},
			&cli.TimestampFlag{
				Name:        "after",
				Usage:       "Delete all tweets after this time.",
				Layout:      "2006-01-02T15:04:05",
				DefaultText: "ignored",
			},
			&cli.StringFlag{
				Name:        "contains",
				Usage:       "Delete all tweets that match a regex pattern.",
				DefaultText: "ignored",
			},
			&cli.BoolFlag{
				Name:  "invert",
				Value: false,
				Usage: "Delete tweets that do _not_ match the specified rules",
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
