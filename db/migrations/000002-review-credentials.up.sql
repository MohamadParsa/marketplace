-- Credentials remain private and persist with the same unified user.
ALTER TABLE review_users ADD COLUMN email text NOT NULL DEFAULT '';
ALTER TABLE review_users ADD COLUMN password_hash text NOT NULL DEFAULT '';
CREATE UNIQUE INDEX review_users_email_unique ON review_users (lower(email)) WHERE email <> '';
