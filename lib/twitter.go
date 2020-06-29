package histweet

import (
	"errors"
	"log"
	// "regexp"
	// "time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Builds a new Twitter client with given args
func NewTwitterClient(args *Args) (*twitter.Client, error) {
	config := oauth1.NewConfig(args.ConsumerKey, args.ConsumerSecret)
	token := oauth1.NewToken(args.AccessToken, args.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify the user
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, errors.New("Invalid user credentials provided")
	}

	log.Printf("Verified user: %+v", user)

	return client, nil
}
