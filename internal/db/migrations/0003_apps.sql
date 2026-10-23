-- Apps on a user's dashboard. Each app belongs to one user and is never shared.
-- Categories are plain text, created by typing a new name when adding an app.
-- position orders a user's apps; drag-to-reorder will rewrite it.
CREATE TABLE apps (
    id         INTEGER PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    url        TEXT    NOT NULL,
    category   TEXT    NOT NULL,
    favourite  INTEGER NOT NULL DEFAULT 0 CHECK (favourite IN (0, 1)),
    new_tab    INTEGER NOT NULL DEFAULT 1 CHECK (new_tab IN (0, 1)),
    position   INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE INDEX apps_user_position ON apps (user_id, position, id);
