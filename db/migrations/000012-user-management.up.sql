ALTER TABLE review_users
 ADD COLUMN first_name text NOT NULL DEFAULT '',
 ADD COLUMN last_name text NOT NULL DEFAULT '',
 ADD COLUMN gender text NOT NULL DEFAULT '',
 ADD COLUMN version integer NOT NULL DEFAULT 0 CHECK (version >= 0),
 ADD COLUMN password_change_required boolean NOT NULL DEFAULT false;
ALTER TABLE review_users ADD CONSTRAINT review_users_gender CHECK (gender IN ('','female','male'));
