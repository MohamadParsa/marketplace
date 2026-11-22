# نتیجه پیاده‌سازی و بررسی Issue #7

تاریخ:۲۰۲۶-۱۱-۲۱ | قابلیت: CV-R2-07 | وضعیت کد: review، پیاده‌سازی و بررسی محلی انجام‌شده | وضعیت مشخصات: in-review | ادغام/استقرار انجام نشده

## دامنه فعال‌شده و رفتار حاصل

این کار قبلاً deferred بود. دستور مستقیم مالک برای اجرای #7 با شرایط تحویل قبلی، فقط بنر دو اندازه را طبق [ADR-006](../../decisions/adr-006-responsive-banner-activation.md) فعال کرد. #8 ریل دستی و #9 sticky معوق‌اند؛ صفحه اصلی ثابت و اتصال نهایی سایت #11 خودکار تغییر نکردند.

صفحه `/banners` ساخت بنر، فرم typeدار متن جایگزین/لینک داخلی/فعالیت، دو جایگاه مستقل انتخاب/پیش‌نمایش/ثبت/حذف تصویر و انتشار نسخه‌دار دارد. انتخاب و لغو محلی فایل upload نمی‌کنند؛ ثبت تصویر و ذخیره فرم فقط draft را تغییر می‌دهند. Published کپی مستقل Config است و revision خودش را دارد.

فایل draft از مسیر عمومی قابل دریافت نیست. رسانه منتشرشده قبلی تا انتشار بعدی حفظ می‌شود؛ حذف/تعویض draft یا شکست انتشار عمومی قبلی را خراب نمی‌کند. HTML مستقل نسخه منتشرشده با picture در۳۹۰ و۱۴۴۰ asset مناسب را انتخاب می‌کند؛ این viewer برای بررسی است و صفحه اصلی مشتری نیست.

## شواهد پذیرش

| معیار | شاهد | نتیجه |
| --- | --- | --- |
| CV-R2-07-AC01 | [مرورگر](../../../tests/e2e/responsive-banners.cjs)، [تست HTTP](../../../apps/api/internal/bootstrap/banners_test.go) | دو فایل و preview مستقل، انتخاب desktop در۱۴۴۰ و mobile در۳۹۰ با currentSrc واقعی بررسی شد |
| CV-R2-07-AC02 | [تست سرویس](../../../apps/api/internal/application/backoffice/banners_test.go)، HTTP و مرورگر | عدم نشت draft، حفظ public قبل از publish، رد فایل/لینک نامعتبر و رسانه ناقص، انتشار اتمیک revision بعدی بررسی شد |

کنترل مکمل:401/403 برای مهمان/viewer/seller، تغییر نقش جاری، نسخه کهنه، یک برنده از هشت publish هم‌زمان، retry یکسان با یک audit، برگشت رسید اصلی upload در retry، غیرفعال‌شدن public فقط پس از انتشار، جمع‌آوری فایل بدون مرجع و حفظ فایل عمومی قبلی آزموده شدند. فایل خام در list یا workspace نشت نمی‌کند.

## نتیجه اجرای بررسی‌ها

- `go test ./...` و `go vet ./...` موفق؛۶۴ تست و زیرتست موفق در مجموعه API. [خلاصه ماشین‌خوان](issue-7-checks.json) از خروجی JSON اجرای واقعی تولید شده است. تست‌های قبلی آیکن نیز موفق‌اند؛ normalizer مشترک سقف قبلی آیکن را تغییر نداده است.
- `npm run lint` و `npm run build` در backoffice-web موفق. هشدارهای غیرمسدودکننده قبلی bundle/annotation وابستگی باقی‌اند.
- Playwright CLI:۲۰/۲۰ بررسی نهایی موفق، بدون pageerror. رد عمدی لینک ناامن با400 و بررسی private asset با404 وجود دارد؛ ادعای نبود تمام خطاهای console/network نمی‌شود.
- تصاویر دسکتاپ/موبایل به‌صورت بصری بازبینی شدند؛ فرم موبایل overflow افقی ندارد و محتوای بلند داخل dialog اسکرول می‌شود. پیام موفقیت داخل فرم بنر است و روی دکمه‌های اقدام نمی‌افتد.
- `go test -race` محلی اجرا نشد؛ CGO ویندوز غیرفعال است. CI لینوکس آن را جدا اجرا می‌کند و نتیجه نهایی در PR/کامنت ایشو ثبت می‌شود.
- دو سناریوی [BDD](../../bdd/features/experiences/responsive-banners.feature) ثبت شد؛ runner مستقل Gherkin اجرا نشده است.

## تکرار بررسی

پس از build وب، از `apps/api` با fixture تازه اجرا کنید؛ مرورگر موفق داده را تغییر می‌دهد:

~~~powershell
$env:APP_ENV='test'
$env:PORT='8097'
$env:WEB_ROOT='../backoffice-web/build'
go run ./cmd/api
~~~

سپس از ریشه مخزن:

~~~powershell
npx --yes --package @playwright/cli playwright-cli -s=cherriva-banners open http://127.0.0.1:8097/
npx --yes --package @playwright/cli playwright-cli -s=cherriva-banners run-code --filename tests/e2e/responsive-banners.cjs
~~~

دو fixture ساده و ساختگی با canvas ساخته شده‌اند: [دسکتاپ](../../../tests/e2e/fixtures/banner-desktop.png) و [موبایل](../../../tests/e2e/fixtures/banner-mobile.png). این‌ها محتوای تبلیغاتی نهایی یا طرح بنر برند نیستند.

## تصاویر مرور

- [فرم و دو preview در دسکتاپ](../../../output/playwright/issue-7/editor-desktop.png)
- [فرم بنر در موبایل](../../../output/playwright/issue-7/editor-mobile.png)
- [فایل انتخاب‌شده در نمایش عمومی۱۴۴۰](../../../output/playwright/issue-7/public-desktop.png)
- [فایل انتخاب‌شده در نمایش عمومی۳۹۰](../../../output/playwright/issue-7/public-mobile.png)

## محدودیت و تحویل

داده و فایل حافظه‌ای آزمون‌اند و با restart پاک می‌شوند. سیاست۵MiB/۴۰۹۶/۸مگاپیکسل و لینک داخلی، انتخاب فنی این برش است؛ سیاست تولید نیست. storage/CDN واقعی، Page/reorder، زمان‌بندی، ریل دستی، sticky و زیرسایت اجرا نشده‌اند. منتشرشدن snapshot داخل سرویس آزمون با deploy کد روی Liara یا انتشار فروشگاه واقعی یکی نیست.

[قرارداد](../../architecture/contracts/responsive-banners.md) · [شرح کار](tasks/cv-r2-07-responsive-banner.md) · [Issue#7](https://github.com/MohamadParsa/marketplace/issues/7). تغییرات روی شاخه #6 بنا شده‌اند؛ زنجیره PRهای #18/#19/#22 پیش از این PR باید تعیین تکلیف شود. طبق دستور مالک، پس از کنترل شواهد و CI کامنت نتیجه ثبت و Issue بسته می‌شود؛ وضعیت review کد از merge/deploy جداست.
