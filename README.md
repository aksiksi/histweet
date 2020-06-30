# Histweet

## Build

`cd cli && go build -o histweet`

## Usage

```
NAME:
   histweet - Manage your tweets via an intuitive CLI

USAGE:
   histweet [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --consumer-key value     Twitter API consumer key [$HISTWEET_CONSUMER_KEY]
   --consumer-secret value  Twitter API consumer secret key [$HISTWEET_CONSUMER_SECRET]
   --access-token value     Twitter API access token [$HISTWEET_ACCESS_TOKEN]
   --access-secret value    Twitter API access secret [$HISTWEET_ACCESS_SECRET]
   --before value           Delete all tweets before this time (ex: 2020-May-10) (default: ignored)
   --after value            Delete all tweets after this time (ex: 2020-May-10) (default: ignored)
   --age value              Delete all tweets older than this age (ex: 10d, 1m, 1y, 1d6m, 1d3m1y)
   --contains value         Delete all tweets that match a regex pattern (default: ignored)
   --max-likes value        Only tweets with fewer likes will be deleted (default: 0)
   --max-replies value      Only tweets with fewer replies will be deleted (default: 0)
   --max-retweets value     Only tweets with fewer retweets will be deleted (default: 0)
   --count value            Only keep the "count" most recent tweets (all other rules are ignored!) (default: 0)
   --invert                 Delete tweets that do _not_ match the specified rules (default: false)
   --no-prompt              Do not prompt user to confirm deletion - ignored in daemon mode (default: false)
   --daemon                 Run the CLI in daemon mode (default: false)
   --interval value         Interval at which to check for tweets, in seconds (default: 30)
   --help, -h               show help (default: false)
```
