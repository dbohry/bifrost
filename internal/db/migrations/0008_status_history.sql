-- How many status checks were made for an app in each hour, and how many failed. It feeds the
-- 24-hour bar on the status card. hour is the Unix time divided by 3600, so every hour starts
-- on the hour in UTC. Old rows are deleted as they age out.
CREATE TABLE status_hours (
    app_id   INTEGER NOT NULL REFERENCES apps (id) ON DELETE CASCADE,
    hour     INTEGER NOT NULL,
    checks   INTEGER NOT NULL CHECK (checks >= 0),
    failures INTEGER NOT NULL CHECK (failures >= 0 AND failures <= checks),
    PRIMARY KEY (app_id, hour)
) STRICT, WITHOUT ROWID;
