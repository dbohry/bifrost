-- When a user last signed in or used a session (Unix seconds). NULL means never.
-- It is updated at sign-in and when a session is extended, so it is accurate to
-- a few minutes, which is enough for "Active now" and "Last seen yesterday".
ALTER TABLE users ADD COLUMN last_seen_at INTEGER;
