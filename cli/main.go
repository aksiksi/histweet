package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func buildCliApp() *cli.App {
	// Define CLI flags
	countFlags := []cli.Flag{
		&cli.IntFlag{
			Name:    "count",
			Aliases: []string{"n"},
			Usage:   "Only keep the `N` most recent tweets (all other rules are ignored!)",
		},
		&cli.StringFlag{
			Name:     "consumer-key",
			Usage:    "Twitter API consumer `key`",
			EnvVars:  []string{"HISTWEET_CONSUMER_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "consumer-secret",
			Usage:    "Twitter API consumer secret `key`",
			EnvVars:  []string{"HISTWEET_CONSUMER_SECRET"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-token",
			Usage:    "Twitter API access `token`",
			EnvVars:  []string{"HISTWEET_ACCESS_TOKEN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-secret",
			Usage:    "Twitter API access secret `token`",
			EnvVars:  []string{"HISTWEET_ACCESS_SECRET"},
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "no-prompt",
			Value: false,
			Usage: "Do not prompt user to confirm deletion - ignored in daemon mode",
		},
		&cli.BoolFlag{
			Name:    "daemon",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "Run the CLI in daemon mode",
		},
		&cli.IntFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Value:   MIN_DAEMON_INTERVAL,
			Usage:   "Interval at which to check for tweets, in `seconds`",
		},
	}

	tweetFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "age",
			Aliases: []string{"a"},
			Usage:   "Delete all tweets older than given `age` (ex: 10d, 1m, 1y, 6m1d, 1y3m1d)",
		},
		&cli.StringFlag{
			Name:        "contains",
			Aliases:     []string{"c"},
			Usage:       "Delete all tweets that contain the given `string`",
			DefaultText: "ignored",
		},
		&cli.StringFlag{
			Name:        "match",
			Aliases:     []string{"m"},
			Usage:       "Delete all tweets that match given `regex`",
			DefaultText: "ignored",
		},
		&cli.TimestampFlag{
			Name:        "before",
			Aliases:     []string{"b"},
			Usage:       "Delete all tweets before given `date` (ex: 2020-May-10)",
			Layout:      "2006-Jan-02",
			DefaultText: "ignored",
		},
		&cli.TimestampFlag{
			Name:        "after",
			Aliases:     []string{"f"},
			Usage:       "Delete all tweets after given `date` (ex: 2020-May-10)",
			Layout:      "2006-Jan-02",
			DefaultText: "ignored",
		},
		&cli.IntFlag{
			Name:    "max-likes",
			Aliases: []string{"l"},
			Usage:   "Delete all tweets with fewer than `N` likes",
		},
		&cli.IntFlag{
			Name:    "max-replies",
			Aliases: []string{"r"},
			Usage:   "Delete all tweets with fewer than `N` replies",
		},
		&cli.IntFlag{
			Name:    "max-retweets",
			Aliases: []string{"t"},
			Usage:   "Delete all tweets with fewer than `N` retweets",
		},
		&cli.StringFlag{
			Name:        "archive",
			Usage:       "Path to tweet archive `file` (tweet.js)",
			DefaultText: "Timeline API lookup",
		},
		&cli.BoolFlag{
			Name:  "invert",
			Value: false,
			Usage: "Delete tweets that do _not_ match the specified rules",
		},
		&cli.BoolFlag{
			Name:  "any",
			Value: false,
			Usage: "Delete tweets that match _any_ of the rules",
		},
		&cli.StringFlag{
			Name:     "consumer-key",
			Usage:    "Twitter API consumer `key`",
			EnvVars:  []string{"HISTWEET_CONSUMER_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "consumer-secret",
			Usage:    "Twitter API consumer secret `key`",
			EnvVars:  []string{"HISTWEET_CONSUMER_SECRET"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-token",
			Usage:    "Twitter API access `token`",
			EnvVars:  []string{"HISTWEET_ACCESS_TOKEN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-secret",
			Usage:    "Twitter API access secret `token`",
			EnvVars:  []string{"HISTWEET_ACCESS_SECRET"},
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "no-prompt",
			Value: false,
			Usage: "Do not prompt user to confirm deletion - ignored in daemon mode",
		},
		&cli.BoolFlag{
			Name:    "daemon",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "Run the CLI in daemon mode",
		},
		&cli.IntFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Value:   MIN_DAEMON_INTERVAL,
			Usage:   "Interval at which to check for tweets, in `seconds`",
		},
	}

	// Define the histweet CLI
	app := &cli.App{
		Name:     "histweet",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Assil Ksiksi",
			},
		},
		Usage: "Manage your tweets via an intuitive CLI",
		Commands: []*cli.Command{
			{
				Name:    "count",
				Flags:   countFlags,
				Usage:   "Simply delete all but the N latest tweets",
				Aliases: []string{"c"},
				Action:  handleCli,
			},
			{
				Name:    "rules",
				Flags:   tweetFlags,
				Usage:   "Delete all tweets that match one or more rules",
				Aliases: []string{"r"},
				Action:  handleCli,
			},
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
