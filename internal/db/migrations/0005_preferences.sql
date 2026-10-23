-- Each user's look: light, dark or follow the system; which theme to use in each
-- mode; and which logo. A user with no row has the defaults. The theme ids are
-- not checked here, because themes added later live outside the database.
CREATE TABLE preferences (
    user_id     INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    mode        TEXT NOT NULL CHECK (mode IN ('light', 'dark', 'system')),
    light_theme TEXT NOT NULL,
    dark_theme  TEXT NOT NULL,
    logo        TEXT NOT NULL CHECK (logo IN ('app', 'arches', 'spectrum'))
) STRICT;
