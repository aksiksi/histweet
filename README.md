# Histweet

## Build

`go build`

## Usage

```
NAME:
   histweet - Manage your tweets via an intuitive CLI

USAGE:
   histweet [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --daemon                 Run the CLI in daemon mode (default: false)
   --interval value         Interval at which to check for tweets, in seconds (default: 30)
   --consumer-key value     Twitter API consumer key [$HISTWEET_CONSUMER_KEY]
   --consumer-secret value  Twitter API consumer secret key [$HISTWEET_CONSUMER_SECRET]
   --access-token value     Twitter API access token [$HISTWEET_ACCESS_TOKEN]
   --access-secret value    Twitter API access secret [$HISTWEET_ACCESS_SECRET]
   --before value           Delete all tweets before this time (default: ignored)
   --after value            Delete all tweets after this time (default: ignored)
   --contains value         Delete all tweets that match a regex pattern (default: ignored)
   --invert                 Delete tweets that do _not_ match the specified rules (default: false)
   --no-prompt              Do not prompt user to confirm deletion - ignored in daemon mode (default: false)
   --help, -h               show help (default: false)
```
