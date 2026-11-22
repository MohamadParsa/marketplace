# اجرای محلی چریوا — Windows / D:\Cherriva

به‌روزرسانی: ۲۰۲۶-۱۱-۲۲. سورس کامل از [PR #38](https://github.com/MohamadParsa/marketplace/pull/38) در شاخه `main` موجود است؛ [کامیت ادغام سورس](https://github.com/MohamadParsa/marketplace/commit/3163d1702f2396d8d1818d12c7bf92bcab9a35bf). این راهنما اجرای بررسی محلی با PostgreSQL پایدار را توضیح می‌دهد.

## اجرای سریع روی سیستم فعلی الیاس

در PowerShell:

```powershell
Set-Location D:\Cherriva
.\Start-Cherriva.cmd
```

راه‌انداز وابستگی‌های دو frontend را با `npm ci` نصب می‌کند، هر دو را build می‌کند، Go API را می‌سازد و سایت و بک‌آفیس را اجرا می‌کند. PostgreSQL خصوصی نسخه بررسی همین سیستم نیز در صورت توقف راه‌اندازی می‌شود. اولین اجرا به اینترنت برای دریافت dependencyها نیاز دارد.

اگر frontendها از قبل با سورس فعلی build شده‌اند:

```powershell
Set-Location D:\Cherriva
powershell -NoProfile -ExecutionPolicy Bypass -File .\Start-Cherriva.ps1 -NoBuild
```

پس از تغییر React یا دریافت نسخه جدید، اجرای کامل را انجام دهید؛ `-NoBuild` فقط build frontend را رد می‌کند و می‌تواند رابط قدیمی نشان دهد.

| محیط | آدرس |
| --- | --- |
| سایت | http://localhost:5197 |
| بک‌آفیس | http://localhost:8098 |
| ورود فروشنده در پنل مشترک | http://localhost:8098/seller-login |

[حساب ادمین و رمز مشترک تیم](delivery/local-review-access.md) در دیتابیس بررسی همین سیستم موجود است. برای مشاهده تغییرات، مرورگر را refresh کنید و در صورت نیاز دوباره وارد شوید؛ نشست ورود پس از restart قابل اتکا نیست.

## راه‌اندازی روی سیستم دیگر، از جمله سیستم پارسا

پیش‌نیازها: Git، Go سازگار با `apps/api/go.mod` (حداقل 1.23)، Node.js/npm و PostgreSQL. این تحویل محلی با Node 24 و PostgreSQL 17 بررسی شده است. راه‌انداز Windows است و روی سیستم دیگر PostgreSQL را دانلود یا نصب نمی‌کند.

اگر پوشه پروژه هنوز وجود ندارد:

```powershell
git clone https://github.com/MohamadParsa/marketplace.git D:\Cherriva
Set-Location D:\Cherriva
```

اگر clone موجود است، ابتدا تغییرات محلی را بررسی کنید و سپس از شاخه `main` نسخه تازه بگیرید:

```powershell
Set-Location D:\Cherriva
git status
git switch main
git pull --ff-only
```

در PostgreSQL یک دیتابیس اختصاصی بررسی با کاربر دارای مجوز ایجاد جدول بسازید. در همان PowerShell، آدرس اتصال دیتابیس خود را جایگزین مقدار نمونه کنید؛ اطلاعات اتصال را commit نکنید:

```powershell
$env:REVIEW_DATABASE_URL = 'postgres://REVIEW_USER:REVIEW_PASSWORD@127.0.0.1:5432/cherriva_review?sslmode=disable'
Remove-Item Env:DATABASE_URL -ErrorAction SilentlyContinue
.\Start-Cherriva.cmd
```

متغیر اتصال باید در همان نشست تنظیم شود؛ راه‌انداز فایل `.env` را بارگذاری نمی‌کند. راه‌انداز ریشه پروژه migrationهای نسخه بررسی را پیش از شروع API اعمال می‌کند و داده در PostgreSQL باقی می‌ماند. `DATABASE_URL` اتصال محیط عمومی است و نباید همراه حالت آزمون تنظیم شود.

حساب مشترک الیاس روی دیتابیس تازه خودکار ساخته نمی‌شود. مدیر ساختگی دیتابیس بررسی تازه `09000000001` است: با شماره موبایل وارد شوید و کد OTP آزمون نمایش‌داده‌شده در فرم را وارد کنید. این کد فقط برای اجرای loopback و `APP_ENV=test` است. راه‌انداز، ارسال پیامک واقعی را غیرفعال می‌کند؛ داده بررسی جای داده عملیاتی نیست.

## داده نمونه تنوع و بررسی خرید

پس از روشن‌شدن بک‌آفیس، از ریشه پروژه می‌توان سه محصول ساختگی دارای محور رنگ/گارانتی و دو فروشنده را ساخت:

```powershell
node .\tests\e2e\seed-variant-showcase.cjs
```

اسکریپت فقط API محلی دارای OTP آزمون را می‌پذیرد و URL محصولات همین دیتابیس را چاپ می‌کند. یک ترکیب ناموجود نیز برای بررسی PDP دارد. اطلاعات استان/شهر و ارسال نمونه از داده بررسی‌اند؛ فهرست کامل جغرافیایی کشور نیستند. مدیریت استان/شهر و روش‌های ارسال از تنظیمات خرید در بک‌آفیس انجام می‌شود.

برای مرور اصلاحات: PDP ← افزودن/تغییر تعداد ← سبد ← مرحله آدرس و ارسال ← مرحله پرداخت؛ سپس جزئیات سفارش و واریز را در بک‌آفیس بررسی کنید. [گزارش، سناریوها و تصاویر](delivery/initial-launch/purchase-refinements-verification.md) مرجع جزئیات است. ارسال فعلی بازه حداقل/حداکثر روز دارد؛ اسلات ظرفیت‌دار در #16 باقی مانده است.

## خطاهای رایج و اطلاعات اجرا

- `go` یا `npm` پیدا نمی‌شود: نصب و PATH را بررسی و PowerShell جدید باز کنید.
- اتصال دیتابیس شکست می‌خورد: PostgreSQL، نام دیتابیس و `REVIEW_DATABASE_URL` را بررسی کنید. روی سیستم دیگر runtime خصوصی الیاس وجود ندارد.
- پورت 8098 یا 5197 اشغال است: راه‌انداز فقط API متعلق به runtime خودش را متوقف می‌کند؛ برنامه دیگر را خودتان بررسی کنید.
- رابط قدیمی است یا فایل build پیدا نمی‌شود: راه‌انداز را بدون `-NoBuild` اجرا کنید.
- روش ارسال برای آدرس نیست: فعال‌بودن استان/شهر و پوشش روش ارسال همان فروشنده را در پنل بررسی کنید.

لاگ‌ها و فایل اجرایی در `%USERPROFILE%\.codex\local-runtimes\cherriva` هستند؛ از جمله `backoffice.err.log` و `storefront.err.log`. فرآیندهای API در پس‌زمینه اجرا می‌شوند و بستن پنجره راه‌انداز آن‌ها را متوقف نمی‌کند. اجرای دوباره راه‌انداز، APIهای متعلق به خودش را جایگزین می‌کند.

این تحویل روی Liara منتشر نشده است؛ انتشار عمومی با پارسا و گیت باز #12 است. [راهنمای استقرار](delivery/liara-deployment.md) · [تطبیق کامیت سورس و مستندات](delivery/source-document-map.md).
