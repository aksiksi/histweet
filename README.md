# Histweet

`histweet` is a simple CLI tool that automatically manages your Twitter timeline.

By default, due to limits in the Twitter API, `histweet` can only process that latest 3,200 tweets in your timeline. However, if you point `histweet` to an optional [Twitter archive](https://help.twitter.com/en/managing-your-account/how-to-download-your-twitter-archive), the tool can process tweets based on your entire history.

## Prerequisites

Before you can use `histweet`, you must first obtain a Twitter API key and access token. To do this, signup for a Twitter Developer account [here](https://developer.twitter.com/en/apply-for-access).

Once you have your keys, you can either pass them in as CLI flags or define the following environment variables:

```
$ export HISTWEET_ACCESS_TOKEN=[YOUR_KEY]
$ export HISTWEET_ACCESS_SECRET=[YOUR_KEY]
$ export HISTWEET_CONSUMER_KEY=[YOUR_KEY]
$ export HISTWEET_CONSUMER_SECRET=[YOUR_KEY]
```

## Quickstart

`histweet` comes with two basic modes: count mode and rules mode.

### Count Mode

In count mode, `histweet` just keeps the latest `N` tweets.  For example, we can keep the latest 300 tweets and delete everything else like so (`--daemon` keeps `histweet` running in the background).

```
histweet count -n 300 --daemon
```

### Rules Mode

Rules mode is the more powerful and... practical mode.  In this mode, you can specify one or more *rules*. `histweet` will delete **all** tweets that match **all** of the provided rules. You can pass in the `--any` flag to instead delete tweets that match **any** of the provided rules.

In this example, we delete all tweets that are:

1. Older than 3 months and 5 days, and;
2. Have fewer than 3 likes, and;
3. Contain the word "dt" (as a regex pattern)

```
histweet rules --age 3m5d --max-likes 3 --match '\sdt(\s|\.|$)'
```

To point `histweet` at your archive JSON, add the `--archive` flag like so:

```
histweet rules --age 3m5d --max-likes 3 --match '\sdt(\s|\.|$)' --archive /path/to/tweet.js
```

The tool will now run the rules against the contents of your archive, and then use the Twitter API to delete all matching tweets. **Note:** the archive does not contain number of replies, so that rule is simply *ignored*.

You can view full usage by passing in the `-h` flag.

## Build

`cd cli && go build -o histweet`
