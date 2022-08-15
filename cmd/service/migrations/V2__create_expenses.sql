CREATE TABLE expenses(
    id SERIAL PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,

    expense_date DATE NOT NULL,
    amount BIGINT NOT NULL,
    category VARCHAR(255),
    sub_category VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_email FOREIGN KEY (user_email) REFERENCES users(email) ON DELETE CASCADE
);
