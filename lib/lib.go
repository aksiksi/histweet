package histweet

import (
	"fmt"
)

func runSingle(args *Args) error {
	// Build the Twitter client
	client, err := newTwitterClient(args)
	if err != nil {
		return err
	}

	// Fetch tweets based on provided rules
	// For now, we assume that user wants to use the timeline API
	tweets, err := fetchTimelineTweets(args.Rule, client)
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

	err = deleteTweets(tweets, client)
	if err != nil {
		return err
	}

	return nil
}

func runDaemon(args *Args) error {
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

	if args.Daemon {
		return runDaemon(args)
	} else {
		return runSingle(args)
	}
}
