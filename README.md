# mcduck

Simple tool to run analytics on personal finance records.

## Getting started

Run everything in docker:

```sh
# Spins up service dependencies, such as postgres within a network called "service_mcduck".
docker compose -f docker-compose.yml up -d

# Builds a docker image of the service
earthly +service-docker

# Runs the service
docker run -p 8080:8080 --env-file cmd/service/docker.env --network service_mcduck --rm mcduck:latest

# Builds a docker image of the bot
earthly +service-docker

# Runs the bot
docker run -p 8080:8080 --env-file cmd/bot/docker.env --network service_mcduck --rm mcduck-bot:latest
```

## The bot

McDuck app also has a telegram bot to parse receipts.
