CREATE TABLE email (
    id INTEGER PRIMARY KEY AUTOINCREMENT
        NOT NULL,
    from TEXT NOT NULL,
    to TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_type TEXT NOT NULL,
    body TEXT NOT NULL,
    count_try_send INTEGER,
    time_first_send INTEGER NOT NULL,
    time_send INTEGER,
    err TEXT
);