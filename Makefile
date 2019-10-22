go-get:
	go get \
		github.com/go-telegram-bot-api/telegram-bot-api \
		github.com/alexcesaro/log \
		github.com/mkideal/cli

build: go-get
	go build

docker-build:
	docker build -t telegram-to-teams-bot .
