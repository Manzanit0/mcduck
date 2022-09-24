# mcduck

Simple tool to run analytics on personal finance records.

## Getting started

Run everything in docker:

```sh
# Spins up service dependencies, such as postgres within a network called "service_mcduck".
docker compose -f cmd/service/docker-compose.yml up -d

# Builds a docker image of the service
earthly +docker

# Runs the service
docker run -p 8080:8080 --env-file cmd/service/docker.env --network service_mcduck --rm mcduck:latest
```