package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func buildCliApp() *cli.App {
	// Define CLI flags
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "consumer-key",
			Usage:    "Twitter API consumer `KEY`",
			EnvVars:  []string{"HISTWEET_CONSUMER_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "consumer-secret",
			Usage:    "Twitter API consumer secret `KEY`",
			EnvVars:  []string{"HISTWEET_CONSUMER_SECRET"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-token",
			Usage:    "Twitter API access `TOKEN`",
			EnvVars:  []string{"HISTWEET_ACCESS_TOKEN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-secret",
			Usage:    "Twitter API access secret `TOKEN`",
			EnvVars:  []string{"HISTWEET_ACCESS_SECRET"},
			Required: true,
		},
		&cli.TimestampFlag{
			Name:        "before",
			Aliases:     []string{"b"},
			Usage:       "Delete all tweets before this `DATE` (ex: 2020-May-10)",
			Layout:      "2006-Jan-02",
			DefaultText: "ignored",
		},
		&cli.TimestampFlag{
			Name:        "after",
			Aliases:     []string{"a"},
			Usage:       "Delete all tweets after this `DATE` (ex: 2020-May-10)",
			Layout:      "2006-Jan-02",
			DefaultText: "ignored",
		},
		&cli.StringFlag{
			Name:  "age",
			Usage: "Delete all tweets older than this `AGE` (ex: 10d, 1m, 1y, 1d6m, 1d3m1y)",
		},
		&cli.StringFlag{
			Name:        "contains",
			Aliases:     []string{"c"},
			Usage:       "Delete all tweets that match this `REGEX`",
			DefaultText: "ignored",
		},
		&cli.IntFlag{
			Name:    "max-likes",
			Aliases: []string{"l"},
			Usage:   "Only tweets with fewer than `N` likes will be deleted",
		},
		&cli.IntFlag{
			Name:    "max-replies",
			Aliases: []string{"r"},
			Usage:   "Only tweets with fewer than `N` replies will be deleted",
		},
		&cli.IntFlag{
			Name:    "max-retweets",
			Aliases: []string{"t"},
			Usage:   "Only tweets with fewer than `N` retweets will be deleted",
		},
		&cli.IntFlag{
			Name:    "count",
			Aliases: []string{"n"},
			Usage:   "Only keep the `N` most recent tweets (all other rules are ignored!)",
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
			Usage:   "Interval at which to check for tweets, in `SECONDS`",
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
