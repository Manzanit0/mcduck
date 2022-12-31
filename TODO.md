# Some user stories

## PENDING REVIEW

This would be a page to review receipts, since the receipt parsing is imperfect.

- I want to be able to upload a receipt through Telegram and have all the
  expenses marked as "pending review" in the UI.
- I want to be able to have a dedicated view of "pending review" expenses.
- I want to be able to filter the "pending review" view by receipt (maybe
  one-at-a-time kinda of thing?).
- In the pending review view, I want to be able to see side by side the receipt
  and the expenses.
- In the pending review view, I want to be able to edit the expenses on the fly.
- In the pending review view, I want to be able to mark the receipt along with
  all the expenses as reviewed and g2g.

## RECEIPT VIEW

A new crud-ish page to view the new resource: receipts.

- I want to be able to have a dedicated view with all the receipts I have
  uploaded.
- I want to know (1) when I have uploaded the receipt, (2) the amount of the
  receipt and (3) the vendor.
- I want to be able to see all the expenses for a given receipt -> expenses view
  filter by receipt ID.
- I want to be able to re-review a receipt along with its expenses triggering
  the PENDING REVIEW page again.

## EXPENSES VIEW

This is the good 'ol: `https://<host>/expenses` page.

- I want to have a counter in the expenses view which displays the amount of
  expenses in the view given the filter applied.
- I want to be able to filter by "expenses without receipt"

## Some sensible stuff

This is just some stuff that's not critical, but nearly.

### accounts

- [x] `created_at` and `updated_at` fields for `users` table.
- [ ] send emails upon signup and login
- [ ] allow for resetting password
- [ ] allow for deleting account (settings page)

### expenses

- [x] Upon CSV upload, if logged in, save expenses
- [x] If logged in, display existing data instead of sample.
- [x] allow for adding, modifying and deleting expenses (expenses page)
- [ ] allow for downloading expenses in csv
- [ ] add checksum validation on uploads to prevent duplicate uploads
- [ ] allow for adding data in CSV in any order of columns (use header)

### teams

Basically to share data with somebody else, i.e. your partner.

- [ ] allow another user to join your expenses team
- [ ] allow for kicking a user from your team
- [ ] different team members can have r or rw permissions

### insights

Improve the analytics, which is the key value of the app.

- [ ] Configure month to dive into for charts
- [ ] MoM chart for sub categories too
- [ ] Add currency symbol to charts
- [ ] Add values to chart without having to hover to see

### mobile

Current UI is not mobile friendly.

- [ ] nice visualisation of charts through phone view
