DROP INDEX review_users_email_unique;
ALTER TABLE review_users DROP COLUMN password_hash;
ALTER TABLE review_users DROP COLUMN email;
