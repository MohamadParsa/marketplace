ALTER TABLE review_users DROP CONSTRAINT review_users_gender;
ALTER TABLE review_users DROP COLUMN password_change_required, DROP COLUMN version, DROP COLUMN gender, DROP COLUMN last_name, DROP COLUMN first_name;
