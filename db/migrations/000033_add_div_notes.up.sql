CREATE TABLE IF NOT EXISTS div_notes (
    div_id     INT       NOT NULL REFERENCES divs(id) ON DELETE CASCADE,
    user_id    INT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_md TEXT      NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (div_id, user_id)
);
