FROM golang:1.13-alpine as build

RUN apk --no-cache add git make

RUN mkdir -p /go/src/github.com/urmaul/telegram-to-teams-bot
WORKDIR /go/src/github.com/urmaul/telegram-to-teams-bot

COPY Makefile .
COPY main.go .

RUN make build

RUN mkdir /app
RUN go build -o /app/telegram-to-teams-bot

FROM alpine:3.10

WORKDIR /app

RUN apk add --no-cache ca-certificates

RUN adduser -D runner

USER runner
WORKDIR /home/runner

COPY --from=build /app/telegram-to-teams-bot .

ENTRYPOINT [ "./telegram-to-teams-bot" ]
