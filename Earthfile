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

build:
    FROM +project-files
    RUN go build -o build/mcduck cmd/service/main.go
    SAVE ARTIFACT build/mcduck /mcduck AS LOCAL build/mcduck

docker:
    COPY +build/mcduck .
    ENTRYPOINT ["/mcduck/mcduck"]
    SAVE IMAGE mcduck:latest

unit-test:
    FROM +project-files
    RUN go test ./...

integration-test:
    FROM +project-files
    COPY cmd/service/docker-compose.yml ./ 
    WITH DOCKER --compose docker-compose.yml
        RUN while ! pg_isready --host=localhost --port=5431 --dbname=postgres --username=root; do sleep 1; done ;\
            go test ./...
    END

all:
    BUILD +build
    BUILD +unit-test
    BUILD +integration-test
