CREATE TABLE users(
    email VARCHAR(255) NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,

    PRIMARY KEY(email)
);
