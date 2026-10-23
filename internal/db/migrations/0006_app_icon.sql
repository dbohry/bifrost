-- Which catalogue icon an app uses (the catalogue id, for example "jellyfin").
-- Empty means no icon: the tile shows the app's first letter instead.
ALTER TABLE apps ADD COLUMN icon TEXT NOT NULL DEFAULT '';
