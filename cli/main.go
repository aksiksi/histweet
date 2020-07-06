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
			Name:        "archive",
			Usage:       "Path to tweet archive `file` (tweet.js)",
			DefaultText: "Timeline API lookup",
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
				Name:    "rule",
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
