-- Status checks. An app can ask Bifrost to look at an address now and then and show
-- whether it answers. check_url is the address to look at; empty means the app's own
-- address. check_strict counts only a 2xx answer as online (otherwise any answer
-- does), and check_insecure accepts a certificate that is self-signed.
ALTER TABLE apps ADD COLUMN check_enabled  INTEGER NOT NULL DEFAULT 0 CHECK (check_enabled  IN (0, 1));
ALTER TABLE apps ADD COLUMN check_url      TEXT    NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN check_strict   INTEGER NOT NULL DEFAULT 0 CHECK (check_strict   IN (0, 1));
ALTER TABLE apps ADD COLUMN check_insecure INTEGER NOT NULL DEFAULT 0 CHECK (check_insecure IN (0, 1));

-- How often, and how long to wait. A user with no row has 60 and 5 seconds.
CREATE TABLE status_settings (
    user_id          INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    interval_seconds INTEGER NOT NULL CHECK (interval_seconds BETWEEN 10 AND 86400),
    timeout_seconds  INTEGER NOT NULL CHECK (timeout_seconds BETWEEN 1 AND 60)
) STRICT;
