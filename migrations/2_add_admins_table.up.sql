CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY,
    user_id INTEGER REFERENCES users (id),
    app_id INTEGER REFERENCES apps (id)
);

