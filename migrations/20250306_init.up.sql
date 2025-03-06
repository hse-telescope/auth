CREATE TABLE IF NOT EXISTS people (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

INSERT INTO people (username, password) VALUES ('user0', 'password0');