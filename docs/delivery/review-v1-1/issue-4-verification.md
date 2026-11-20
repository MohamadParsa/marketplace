# نتیجه پیاده‌سازی و بررسی Issue #4

تاریخ:۲۰۲۶-۱۱-۲۰ | قابلیت: CV-R2-04 | وضعیت: پیاده‌سازی و بررسی محلی انجام‌شده، در انتظار بازبینی/ادغام | وضعیت سند: in-review

## رفتار حاصل

در صفحه دسته‌ها و صفحه جدید برندهای موجود، مدیرکل/عملیات می‌تواند آیکن انتخاب، پیش‌نمایش، ثبت یا حذف کند. نبود و شکست دریافت به آیکن داخلی برمی‌گردد و نام رکورد حفظ می‌شود. انتخاب لغوشده عمومی نمی‌شود؛ فایل نامعتبر یا نسخه کهنه، تصویر قبلی را از بین نمی‌برد. مجوز سرور مستقل از دکمه رابط کنترل می‌شود.

سیاست فنی این نسخه آزمون: PNG/JPEG/WebP واقعی تا۱MiB و هر ضلع۱۰۲۴؛ decode کامل و بازتولید PNG. [قرارداد کامل](../../architecture/contracts/catalog-icons.md) مرجع داده، route، خطا و محدودیت است.

## شواهد پذیرش

| معیار | شواهد قابل تکرار | نتیجه |
| --- | --- | --- |
| CV-R2-04-AC01؛ نبود، حذف و دریافت خراب | [مرورگر](../../../tests/e2e/catalog-icons.cjs)، roundtrip/remove API | موفق روی دسته و برند؛ نام و fallback باقی می‌مانند |
| CV-R2-04-AC02؛ upload معتبر و حفظ قبلی در خطا | [تست HTTP](../../../apps/api/internal/bootstrap/catalog_icon_test.go)، preview/invalid browser | PNG/JPEG/WebP، فایل جعلی/ناقص/بزرگ و ابعاد نامجاز بررسی شدند؛ stale409 و retry امن |
| CV-R2-04-AC03؛ مجوز و draft | تست HTTP401/403، [تست سرویس](../../../apps/api/internal/application/backoffice/icons_test.go)، preview-cancel و viewer browser | کاربر فاقد مجوز رد؛ actor جاری دوباره کنترل؛ draft/رکورد آرشیو/تصویر بدون اتصال عمومی نمی‌شود |

آزمون هم‌زمانی سرویس: دو حالت کلیدهای متفاوت و retry یکسان؛ فقط یک تغییر/فایل/audit قطعی ثبت شود. رد فایل و تغییر نسخه rollback می‌شوند. داده باینری در workspace خصوصی هم ارسال نمی‌شود.

## فرمان‌ها و گزارش‌ها

- `go test ./...`: موفق؛ [خلاصه ماشین‌خوان](issue-4-checks.json).
- `go vet ./...`: موفق.
- `npm ci --no-audit --no-fund` و `npm run lint` و `npm run build` در backoffice-web: موفق. هشدارهای build درباره اندازه bundle و annotation کتابخانه مانع build نبودند؛ بهینه‌سازی bundle دامنه این کار نیست.
- Playwright CLI: ۱۴/۱۴ بررسی موفق، بدون pageerror؛ شکست شبکه تصویر عمداً برای fallback شبیه‌سازی شد.
- `go test -race` محلی اجرا نشده است: CGO این ویندوز غیرفعال و کامپایلر C در PATH در دسترس نیست. تست‌های هم‌زمانی معمولی موفق‌اند؛ نتیجه race ادعا نمی‌شود.

برای تکرار مرورگر، API را از `apps/api` با `APP_ENV=test` و `PORT=8094` و `WEB_ROOT=../backoffice-web/build` اجرا و از root ریپو فراخوانی کنید:

~~~powershell
npx --yes --package @playwright/cli playwright-cli -s=cherriva-icons open http://127.0.0.1:8094/
npx --yes --package @playwright/cli playwright-cli -s=cherriva-icons run-code --filename tests/e2e/catalog-icons.cjs
~~~

## تصاویر مرور

- [پیش‌نمایش دسته قبل از ثبت](../../../output/playwright/issue-4/category-preview.png)
- [ثبت آیکن در فهرست](../../../output/playwright/issue-4/category-saved.png)
- [رد فایل خراب و حفظ آیکن قبلی](../../../output/playwright/issue-4/invalid-file.png)
- [پیش‌نمایش برند در موبایل](../../../output/playwright/issue-4/brand-mobile-preview.png)
- [فهرست برند در موبایل](../../../output/playwright/issue-4/brands-mobile.png)

## محدودیت تحویل

فایل‌ها و داده‌ها حافظه‌ای و آزمون‌اند و با restart پاک می‌شوند. کاتالوگ کامل برند، storage واقعی، نقش سفارشی، CMS و deploy در این کار انجام نشده‌اند. سه برند fixture فقط امکان آزمون قابلیت را فراهم می‌کنند. پیاده‌سازی در PR تحویل می‌شود؛ merged/deployed و قبولی محیط واقعی از این گزارش نتیجه نمی‌شود. [شرح کار](tasks/cv-r2-04-category-brand-media.md) و [Issue#4](https://github.com/MohamadParsa/marketplace/issues/4) مرجع پیگیری ادغام هستند.
