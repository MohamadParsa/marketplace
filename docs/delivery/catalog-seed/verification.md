# شاهد بسته کاتالوگ و ابزار ورود

تاریخ: ۲۰۲۶-۱۱-۲۲. سورس/داده بررسی‌شده: [`8921ca0`](https://github.com/MohamadParsa/marketplace/commit/8921ca0497aa016d9373ae240c12025046bd3a6d). [خروجی ماشین‌خوان](checks.json) · [SHA و تعدادها](manifest.json).

## داده و رسانه

- JSON واحد: ۵۰٬۵۷۳٬۸۱۴ بایت؛ SHA-256: `c8154676f86f134c0ba98c7f53a7614e38f0c6cad2828b83e61b82eccdfbee98`.
- ۳۵۳۴ محصول، ۵۰۲۹ تنوع، ۵۹۹۲ عرضه، ۲۴۸ دسته، ۱۲۷۷ برند، ۴۴ فروشنده، ۲۴۲ تعریف خصوصیت.
- سهمیه۲۰ برای تمام گروه‌هایی که حداقل۲۰ محصول واجد شرایط داشتند تکمیل شد؛ `selection_shortage=0` برای برند و دسته. گروه‌های کم‌موجودی و فاقد محصول در coverage مشخص‌اند.
- ۱۰۴۹۵ لینک تصویر بررسی شد؛ ۴۳۷ لینک فرعی/آیکن قدیمی ۴۰۴ داشتند و از نگاشت نمایشی حذف شدند. هیچ تصویر اصلی محصول حذف نشد.
- ابزار Go برای **تمام ۱۰۰۵۸ تصویر باقی‌مانده** checksum، اندازه، decode واقعی و تبدیل PNG را بدون نوشتن دیتابیس اجرا کرد: `validation=passed_with_media` و `applied=false`.
- PNGهای حاصل: ۳٬۳۴۹٬۳۸۷٬۲۸۹ بایت. فایل JSON فقط مرجع/URL/SHA تصویر دارد؛ cache و باینری دیتابیس وارد گیت نشده‌اند.
- connection string، رمز، آدرس SQL Server و فایل خصوصی اتصال در فایل‌های staged وجود نداشتند. فایل موقت اتصال پس از استخراج حذف شد. اصل مبدأ هیچ دستور تغییردهنده دریافت نکرد.

## آزمون کد و ذخیره‌سازی

`go test ./...` با PostgreSQL مجزا و cache تصویر واقعی: **۱۸۴ تست/زیرتست موفق، صفر fail و صفر skip**. `go vet ./...` موفق بود. دو frontend تغییر نکرده‌اند؛ آزمون مرورگر تازه‌ای برای این تحویل ادعا نمی‌شود.

آزمون‌های معنادار:

- `TestCatalogSeedRejectsInvalidRelationsAndSourceURLs`: تومان/namespace، مرجع خراب، چرخه دسته، دسته پنهان، انتخاب محور ناسازگار، تنوع تکراری، عرضه متقاطع، پول منفی، تخفیف کاذب، موجودی رزروشده، نبود عرضه/تصویر و URL غیرمجاز.
- `TestCatalogSeedConcurrentRetriesPreserveAccountsAndMoney`: چهار درخواست هم‌زمان، یک ورود و یک audit؛ عدم تغییر کاربران، نقش‌ها، مالی و ledger؛ پاک‌سازی script و حفظ قالب‌بندی مجاز.
- `TestCatalogSeedCollisionRollsBackEverything`: شناسه موجود و ویرایش محلی حفظ می‌شود و هیچ ورود جزئی باقی نمی‌ماند.
- `TestCatalogSeedMissingMediaWritesNothing`: بدون رسانه تأییدشده، کاتالوگ تغییر نمی‌کند.
- `TestCatalogSeedPostgresRestartAndRetry`: ورود فایل‌محور در تراکنش، retry پس از restart، نبود باینری در workspace، ارائه PNG از جدول جدا پس از کنترل دسترسی و حفظ تصویر در نوشتن نامرتبط.
- `TestMediaCacheEnforcesChecksumsAndFullDecode`: hash نامنطبق، PNG ناقص و SVG رد می‌شوند.
- تست‌های موجود خانه، PDP، PLP، آیکن و مالی هم در همان suite موفق بودند؛ API URL تصویر تغییر نکرده است.

در CI، آزمون snapshot واقعی با نبود cache صریحاً skip می‌شود تا CI به CDN و cache خصوصی وابسته نباشد؛ تست‌های PostgreSQL با fixture کوچک همچنان اجرا می‌شوند. race محلی Windows اجرا نشده؛ وضعیت CI جدا ثبت می‌شود.

## ورود نمونه از همین JSON در PostgreSQL موقت

`TestCatalogSnapshotPostgresRoundTrip` با cache خصوصی و دیتابیس آزمون اجرا شد:

| بررسی | نتیجه |
| --- | --- |
| محصول واقعی | ۲۰ |
| تنوع | ۳۴ |
| عرضه | ۳۵ |
| رسانه واقعی | ۱۰۱ |
| HTTP خانه، PLP، بیست PDP و یک تصویر | ۲۳ پاسخ۲۰۰ |
| اجرای دوباره | `already_applied=true` |
| راه‌اندازی مجدد repository | داده محفوظ |
| کاربران، نقش‌ها و مالی | بدون تغییر |
| باینری inline در workspace | صفر |
| schema آزمون | در پایان حذف شد |

فرمان بازتولید از ریشه API، با اتصال PostgreSQL **فقط محیط آزمون** و cache دانلودشده:

```powershell
$env:CATALOG_SEED_FILE = 'D:/Cherriva/data/catalog/gooshishop-2026-11-22.json'
$env:CATALOG_MEDIA_CACHE = "$env:USERPROFILE/.codex/local-runtimes/cherriva/catalog-export/media-cache"
# MARKETPLACE_TEST_POSTGRES_URL را خصوصی به PostgreSQL آزمون تنظیم کنید.
go test ./cmd/catalog-import -run TestCatalogSnapshotPostgresRoundTrip -v
```

این شاهد، ورود کامل ۳۵۳۴ محصول در دیتابیس اصلی را ادعا نمی‌کند. کاتالوگ فعلی الیاس و دیتابیس سرور تغییر نکرده‌اند؛ اجرای ورود کامل فقط با دستور بعدی و [پرامپت آماده](prompts.md) انجام می‌شود. انتشار Liara هم انجام نشده است.

## نتیجه CI

[CI روی d301f72](https://github.com/MohamadParsa/marketplace/actions/runs/45036665925) با هر پنج job موفق شد: API شامل Linux race/PostgreSQL/vet، بک‌آفیس، سایت، پنل فروشنده و contracts. کامیت بعدی فقط ثبت همین نتیجه در مدارک است و tree سورس/JSON آن با head آزمایش‌شده یکسان کنترل می‌شود. همگام‌سازی main با پیام `[skip ci]` انجام می‌شود تا workflow استقرار خودکار Liara شروع نشود؛ هیچ تنظیم workflow غیرفعال نشده است.
