# slackln

Println(Slack).

## Usage

### channel history

``` console
$ SLACK_OAUTH_TOKEN=xxx slackln history -c my-channel --duration=3days
```

## Install

**homebrew tap:**

```console
$ brew install k1LoW/tap/slackln
```

**manually:**

Download binany from [releases page](https://github.com/k1LoW/slackln/releases)

**go get:**

```console
$ go get github.com/k1LoW/slackln
```

## TODO

- [ ] `slackln search`: Search messages ( using https://api.slack.com/methods/search.messages )
- [ ] `slackln server`: OAuth server
