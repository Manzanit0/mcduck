VERSION 0.6
FROM golang:1.19.1-alpine3.15
WORKDIR /mcduck
RUN apk add build-base bash postgresql-client

project-files:
    COPY go.mod go.mod
    COPY go.sum go.sum
    COPY cmd cmd
    COPY pkg pkg
    COPY internal internal

service-build:
    FROM +project-files
    RUN go build -o build/mcduck cmd/service/main.go
    SAVE ARTIFACT build/mcduck /mcduck AS LOCAL build/mcduck

service-docker:
    COPY +service-build/mcduck .
    ENTRYPOINT ["/mcduck/mcduck"]
    SAVE IMAGE mcduck:latest

bot-build:
    FROM +project-files
    RUN go build -o build/mcduck-bot cmd/bot/main.go
    SAVE ARTIFACT build/mcduck-bot /mcduck-bot AS LOCAL build/mcduck-bot

bot-docker:
    COPY +bot-build/mcduck-bot .
    ENTRYPOINT ["/mcduck-bot/mcduck-bot"]
    SAVE IMAGE mcduck-bot:latest

unit-test:
    FROM +project-files
    RUN go test ./...

integration-test:
    FROM +project-files
    COPY docker-compose.yml ./
    WITH DOCKER --compose docker-compose.yml
        RUN while ! pg_isready --host=localhost --port=5431 --dbname=postgres --username=root; do sleep 1; done ;\
            go test ./...
    END

all:
    BUILD +service-build
    BUILD +bot-build
    BUILD +unit-test
    BUILD +integration-test
