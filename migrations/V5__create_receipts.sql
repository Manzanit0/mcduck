BEGIN;

CREATE TABLE receipts (
    id SERIAL PRIMARY KEY,

    receipt_image BYTEA,
    pending_review BOOLEAN DEFAULT TRUE,
    vendor VARCHAR(255),

    user_email VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_email
    FOREIGN KEY (user_email)
    REFERENCES users (email) ON DELETE CASCADE
);

CREATE TRIGGER receipts_set_timestamp
BEFORE UPDATE ON receipts
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

ALTER TABLE expenses
ADD COLUMN receipt_id INTEGER;

ALTER TABLE expenses
ADD COLUMN description VARCHAR(255);

COMMIT;
