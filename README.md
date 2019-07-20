# Redmine bot for Slack

Redmine issue name expander for Slack. 

Bot parses the issue numbers (`#54321`), the links to issue (`http://redmine.host.com/issues/54321`) and displays the issue name with attributes: 
* project
* tracker
* author
* assigned to
* category
* version

![Example](/static/screenshot.png?raw=true)

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
