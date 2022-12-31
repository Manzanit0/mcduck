# mcduck web service

## prerequisits

- direnv to load `.envrc` for local development. You can do this manually if preferred though.
- docker compose to bootstrap database

## getting started

To run everything locally simply:

```sh
$ direnv allow
$ docker compose up -d
$ go run .
```
