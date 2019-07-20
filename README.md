# Redmine bot for Slack

Redmine issue name expender for Slack.

## Run

### From sources

```
git clone https://github.com/muxx/slack-redmine-bot
cd slack-redmine-bot
cp config.yml.example config.yml
make build
make run
```

The binary will be installed at `./bin/slack-redmine-bot`.

### From docker

Create `config.yml` from example `config.yml.example`.

```
docker pull muxx/slack-redmine-bot
docker run --rm -it -v /path/to/config.yml:/config.yml muxx/slack-redmine-bot:latest
```

## Config and run

Copy and fill config file `config.yml.example`
```
cp config.yml.example config.yml
make run
```
