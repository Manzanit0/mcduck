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

## TODO

accounts:
- [X] `created_at` and `updated_at` fields for `users` table.
- [ ] send emails upon signup and login
- [ ] allow for resetting password
- [ ] allow for deleting account (settings page)

expenses:
- [X] Upon CSV upload, if logged in, save expenses
- [X] If logged in, display existing data instead of sample.
- [ ] allow for adding, modifying and deleting expenses (expenses page)
- [ ] allow for downloading expenses in csv
- [ ] add checksum validation on uploads to prevent duplicate uploads
- [ ] allow for adding data in CSV in any order of columns (use header)

teams:
- [ ] allow another user to join your expenses team
- [ ] allow for kicking a user from your team
- [ ] different team members can have r or rw permissions

insights:
- [ ] Configure month to dive into for charts
- [ ] MoM chart for sub categories too
- [ ] Add currency symbol to charts
- [ ] Add values to chart without having to hover to see

mobile:
- [ ] nice visualisation of charts through phone view

nice stuff:
- [ ] upload receipts as expenses: should be able to amend line items before final upload.