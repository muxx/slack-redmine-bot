# Redmine bot for Slack

Redmine issue name expender for Slack.

## Setup

```
mkdir -p go/src go/bin go/pkg
cd go
export GOPATH=`pwd`

go get github.com/muxx/slack-redmine-bot
```

## Build

```
go install github.com/muxx/slack-redmine-bot
```

The binary will be installed at `./bin/slack-redmine-bot` folder.

## Config

Copy and fill config file `config.yml.example`
```
cp ./src/github.com/muxx/slack-redmine-bot/config.yml.example ./bin/config.yml
nano ./bin/config.yml
```

You can put `config.yml` in folder with the binary file or in the folder `/etc/slack-redmine-bot/`.