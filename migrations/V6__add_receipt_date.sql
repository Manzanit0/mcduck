BEGIN;

ALTER TABLE receipts
ADD COLUMN receipt_date DATE;

-- Backfill existing data with a sensible default.
UPDATE receipts set receipt_date=created_at;

ALTER TABLE receipts
ALTER COLUMN receipt_date SET NOT NULL;

COMMIT;
