-- Which page background a user picked. Existing users keep the plain one.
ALTER TABLE preferences ADD COLUMN background TEXT NOT NULL DEFAULT 'plain'
    CHECK (background IN ('plain', 'glow', 'bridge', 'wash', 'grid'));
