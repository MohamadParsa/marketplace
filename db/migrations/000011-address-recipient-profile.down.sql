DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM review_users WHERE national_id <> '') THEN
  RAISE EXCEPTION 'Profile national IDs exist; use a forward migration';
 END IF;
END $$;
ALTER TABLE review_users DROP COLUMN national_id;
