# Cherriva — سورس Go و React

[برنامه۳۰درخواست جدید؛ وضعیت اجرا و شواهد انتخاب‌های مالک](docs/delivery/next-phase/README.md)

سورس اجرایی در `apps/api`، `apps/storefront-web`، `apps/backoffice-web` و `apps/seller-portal-web` و کد مشترک در `packages` است. مسیر کاری این سیستم **`D:\Cherriva`** است. پنل فروشنده از `/seller-login` در همان رابط بک‌آفیس باز می‌شود. مستندات و بسته‌های طراحی، جایگزین سورس نیستند.

برای بررسی محلی از پوشه پروژه اجرا کنید:

```powershell
Set-Location D:\Cherriva
.\Start-Cherriva.ps1
```

بک‌آفیس: http://localhost:8098 · سایت: http://localhost:5197. اجرای سریع پس از build: `.\Start-Cherriva.ps1 -NoBuild`. [راهنمای کامل اجرای لوکال](docs/local-development.md) · [حساب محلی تیم](docs/delivery/local-review-access.md) · [اصلاحات و تصاویر خرید](docs/delivery/initial-launch/purchase-refinements-verification.md) · [تطبیق کامیت سورس و مستندات](docs/delivery/source-document-map.md).

راه‌اندازی محلی با PostgreSQL جداگانه و `APP_ENV=test` است؛ انتشار لیارا توسط پارسا جداگانه انجام می‌شود.


## داده کاتالوگ آماده ورود

[فهرست آمادهٔ۳۱استان و۱۴۸۱شهر ایران و راهنمای ورود](data/iran-geography/README.md) در دیتابیس بررسی محلی وارد شده است؛ [گزارش حفظ داده‌های قبلی و آزمون](docs/delivery/iran-geography-import-2026-11-27.md).

[بسته JSON، پوشش برند/دسته، راهنمای ورود و پرامپت لوکال/سرور](docs/delivery/catalog-seed/README.md). بسته کامل در دیتابیس بررسی محلی وارد شده و [گزارش ۲۰۲۶-۱۱-۲۳، تصاویر و لینک نمونه‌ها](docs/delivery/catalog-seed/local-import-2026-11-23.md) موجود است؛ سرور هنوز تغییری نکرده است.

## اولویت جاری انتشار اولیه — ۲۰۲۶-۱۱-۲۱

[برنامه چهارمرحله‌ای انتشار و Issueها](docs/delivery/initial-launch/README.md): خانه → PLP → PDP → خرید/کیف پول، هر کدام با بک‌آفیس و API و سایت. کارت‌به‌کارت و محلی حضوری با تأیید دستی/بصری؛ ورودی وجه ابتدا شارژ و مصرف جدا. طرح A و نشان C حرف اول Cherriva مرجع‌اند. [ADR-007](docs/decisions/adr-007-initial-launch-sequence.md) شرط قدیمی انتظار کل پنل‌ها را فقط برای همین انتشار جایگزین می‌کند. خانه #27، PLP #28 و PDP #29 با شاهد فنی تکمیل شدند؛ خرید/کیف پول #30 نیز با [شاهد محلی](docs/delivery/initial-launch/issue-30-verification.md) تحویل شده است؛ گیت عمومی HTTPS/بانک/SMS/backup با پارسا و در #12 باز است. PRها و استقرار جدا هستند. [دسترسی تیم به نسخه محلی](docs/delivery/local-review-access.md).


## پاسخ تصمیم‌های نسخه۱.۱ — ۲۰۲۶-۱۱-۲۰

[پاسخ‌های الیاس برای تسک #13](docs/decisions/adr-005-review-v1-1-product-responses.md) ثبت شد؛ شامل فلو ورود تکمیلی، نقش و عضویت، خصوصیات، ساده‌سازی صفحه اصلی، حمل و ناموجود، انتخاب A پنل و تنظیمات عمومی. پیشنهاد تغییر موبایل و بخش‌های باقی‌مانده صریحاً مشخص‌اند؛ [شرح‌های بازنویسی‌شده مرحله۲](docs/delivery/review-v1-1/README.md) دامنه، وابستگی، AC و شاهد لازم را ثبت می‌کنند؛ هیچ تسک پیاده‌سازی با تدوین آن‌ها done نشده است.


## بسته طراحی برای بررسی مشترک — ۲۰۲۶-۱۱-۱۹

[دیزاین سایت و بک‌آفیس، تصاویر و اسناد تسک‌ها](docs/review-packages/README.md) آماده مشاهده است. ابتدا طرح‌ها را با هم مرور می‌کنیم؛ شفاف‌سازی و اولویت‌بندی نهایی تسک‌ها گام بعدی است. بسته مستقل و در وضعیت in-review منتشر شده و سورس محصول را تغییر نمی‌دهد.


مرجع محصول، مستندات و codebase سه محیط **سایت فروش PWA، پنل فروشگاه PWA و بک‌آفیس وب**. تصمیم‌های معماری اولیه در [ADR-002](docs/decisions/adr-002-codebase-architecture.md) و deployment در [ADR-003](docs/decisions/adr-003-liara-deployment.md) ثبت شده‌اند.

اولویت اولیه تاریخی، بک‌آفیس و پنل فروشندگان پیش از سایت مشتری بود؛ اولویت جاری همان چهار برش انتشار اولیه و ADR-007 در ابتدای این صفحه است. رودمپ کامل در [رودمپ پیاده‌سازی](docs/delivery/roadmap.md) و جزئیات برش‌ها در [برنامه تحویل](docs/delivery/README.md) است.

## از اینجا شروع کنید

1. [دامنه محصول](docs/product/scope.md): چه چیزی قرار است ساخته شود و چه چیزی هنوز تعهد محصول نیست.
2. [نقشه مستندات](docs/README.md): مسیر دسترسی به قواعد و قابلیت‌ها.
3. [پرسش‌های تصمیم‌گیری](docs/product/open-questions.md): فقط خلأهای واقعی، با اثر بر ساخت.
4. [ردیابی قابلیت‌ها](docs/governance/traceability.md): اتصال تجربه، قانون، قرارداد و پذیرش.
5. [راهنمای agent](AGENTS.md): نحوه استفاده بدون فرض‌سازی یا تکرار اطلاعات.

## ساختار و مرجع حقیقت

```text
AGENTS.md
 docs/
   product/               دامنه، اصطلاحات، مجوزها و پرسش‌های باز
   domains/               قواعد مشترک و چرخه‌های کسب‌وکار
   experiences/           رفتار سایت فروش، پنل فروشگاه و بک‌آفیس
   architecture/          مرزهای منطقی، روابط داده، کیفیت و قراردادها
   delivery/              برش‌های پیشنهادی و معیار آمادگی ساخت
   decisions/             دلیل تغییرات مهم تصمیم، بدون کپی قواعد
   governance/            شیوه نگارش و ردیابی
   templates/             قالب سند جدید
```

**قاعده در حوزه، تعامل در تجربه، دلیل انتخاب در تصمیم و ابهام در دفتر پرسش‌ها** نگهداری می‌شود. اسناد به هم لینک می‌دهند؛ نسخه‌های موازی قانون ساخته نمی‌شود.

ساختار اجرایی فعلی:

```text
apps/
  api/                    backend Go و modular monolith
  backoffice-web/         frontend React/Vite بک‌آفیس
  seller-portal-web/      محل frontend پنل فروشنده
  storefront-web/         محل frontend سایت مشتری
db/migrations/            migrationهای مرجع دیتابیس
packages/                 قراردادها و کد مشترک آینده
.github/workflows/        CI و deploy روی Liara
```

راهنمای اجرای این ساختار در [راهنمای استقرار Liara](docs/delivery/liara-deployment.md) آمده است.

قواعد صریح گفتگو ثبت شده‌اند؛ جزئیات طراحی پیشنهادی از آن‌ها جداست. وضعیت مشخصات `in-review` است و به معنی تکمیل پیاده‌سازی یا آمادگی همه مسیرها برای ساخت نیست. برای ساخت هر برش، [معیار آمادگی](docs/delivery/readiness.md) بررسی شود.

## بررسی صفحات بک‌آفیس

نسخه بررسی با داده ساختگی از API Go تغذیه می‌شود. ابتدا `npm ci` و `npm run build` را در `apps/backoffice-web` اجرا کنید. سپس از `apps/api`:

```sh
APP_ENV=test WEB_ROOT=../backoffice-web/build PORT=8087 go run ./cmd/api
```

رابط در `http://localhost:8087` باز می‌شود. حساب مدیر آزمون `09000000001` است؛ کد آزمون در مرحله دوم فرم نمایش داده می‌شود. ایمیل فقط با رمز و شماره موبایل با انتخاب رمز یا OTP وارد می‌شود؛ انتخاب نقش از فرم ورود حذف شده است. داده پیش‌فرض با restart بازنشانی می‌شود؛ برای نگهداری داده ساختگی، [PostgreSQL جداگانه آزمون](docs/architecture/contracts/backoffice-review.md#اجرای-نسخه-پایدار-آزمون) را تنظیم کنید.

[وضعیت مرحله ۱](docs/delivery/stage-1-progress.md) و [بسته استقرار Liara](deploy/liara/README.md) محدودیت‌ها و شواهد اجرا را مشخص می‌کنند.
