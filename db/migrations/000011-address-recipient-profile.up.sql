-- داده‌های قبلی محفوظ‌اند؛ تکمیل کد ملی تنها هنگام انتخاب گیرنده خود مشتری الزامی است.
ALTER TABLE review_users ADD COLUMN national_id text NOT NULL DEFAULT '';
ALTER TABLE review_users ADD CONSTRAINT review_users_national_id_format CHECK (national_id = '' OR national_id ~ '^[0-9]{10}$');
