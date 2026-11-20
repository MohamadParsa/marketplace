# شاهد CV-R2-03 — پروفایل، رمز و بازیابی OTP

تاریخ: ۲۰۲۶-۱۱-۲۱ · محیط: `local/test` با repository حافظه‌ای و داده ساختگی · وضعیت سند: `in-review`

## قرارداد و تغییر قابل مشاهده

- `GET/PATCH /{surface}/account/profile` پروفایل همان نشست را می‌خواند/ویرایش می‌کند.
- `POST /{surface}/auth/password-login` موبایل یا ایمیل را با یک credential resolve می‌کند.
- `POST /{surface}/auth/password-recovery/request-otp` و `.../reset` بازیابی موبایل را با challenge محدود و single-use انجام می‌دهند.
- ایمیل در User اختیاری و یکتا است؛ مقصد ایمیلی بازیابی تا سیاست تأیید انتخاب‌شده فعال نمی‌شود.

## نتیجه AC

| AC | شاهد | نتیجه |
|---|---|---|
| AC01 | `TestProfilePasswordAndMobileRecovery` و `UpdateProfile` | پاس؛ ایمیل تکراری conflict و همان User حفظ می‌شود |
| AC02 | تست password login و پروفایل | پاس؛ hash فقط سمت سرور و DTO فاقد رمز است |
| AC03 | `TestProfilePasswordAndMobileRecovery` و `consumeOTP` | پاس برای موبایل؛ OTP اشتباه/منقضی و مصرف دوباره رد می‌شود |

اعتبارسنجی: `go test ./...` در `apps/api` پاس شد. تست مرورگر/provider واقعی برای این برش اجرا نشده و دروازه اتصال واقعی مستقل است.
