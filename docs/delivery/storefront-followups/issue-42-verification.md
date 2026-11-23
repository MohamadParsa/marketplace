---
id: DEL-CATALOG-URL-001
title: تحویل و شواهد URL عمومی کاتالوگ چریوا
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-CATALOG-PUBLIC-URL-001, API-PRODUCT-DETAIL-001, API-PLP-SEARCH-001, FEAT-SF-CACHE-001]
approved_by: []
implementation_status: review
---

# تحویل #42

این قابلیت پس از ادغام و بستن #41 روی سورس `D:\Cherriva` اجرا شد. [نیازمندی و ACها](public-catalog-urls.md)، [نگاشت کامل ۳۵۳۴ محصول و ۲۴۸ دسته](catalog-url-map.json) و [نتایج ماشین‌خوان](issue-42-checks.json) مرجع تحویل‌اند. وضعیت تأیید سند از وضعیت اجرای قابلیت جداست؛ انتشار لیارا انجام نشده است.

## رفتار اجراشده

- محصولات واردشده کد `chr-{شماره قبلی}` دارند؛ محصول۳۹۷۸۲ اکنون `/product/chr-39782` است. کد در فرم محصول فقط‌خواندنی و در PDP قابل مشاهده است. شناسه داخلی محصول، variant، offer، SKU و مرجع سفارش تغییر نکرده‌اند.
- دسته موبایل `/category/mobile`، پاوربانک `/category/power-bank` و ساعت هوشمند `/category/smart-watch` است. نام انگلیسی اصلی ساعت (`gadget`) دست‌نخورده ماند؛ slug مستقل و قابل ویرایش است. فهرست کامل برای بازبینی محتوایی در نگاشت آمده است؛ دسته واقعی۷۵۵ به موبایل وصل نشده است.
- مقدار پیش‌فرض slug از انگلیسی، با شکستن camelCase و تبدیل فاصله/underscore/نشانه‌ها به dash ساخته می‌شود. مقدار نهایی باید ASCII lowercase، حداکثر۱۰۰نویسه و مطابق `[a-z0-9]+(?:-[a-z0-9]+)*` باشد. دسته واردشده بدون نام انگلیسی و slug، قالب نادرست و تعارض در dry-run متوقف می‌شوند؛ هیچ دسته تازه‌ای بی‌صدا پسوند عددی نمی‌گیرد. fixtureهای قدیمی غیرـgsp با ID معتبر، آدرس قبلی خود را در محیط آزمون نگه می‌دارند.
- تغییر نام انگلیسی به‌تنهایی slug ذخیره‌شده را عوض نمی‌کند. تغییر slug دسته منتشرشده در طرح A تأیید کاربر، version و مجوز فعلی می‌خواهد؛ فرمان و fingerprint آن در audit موجود ثبت می‌شود. alias قبلی برای همان دسته حفظ و قابل تصاحب توسط دسته دیگر نیست.
- لینک قدیمی HTML پس از بررسی انتشار/visibility، با یک **301 سروری** به مقصد جاری همان موجودیت می‌رود؛ تمام aliasهای قدیمی مستقیم به slug فعلی وصل‌اند. variant/offer معتبر و query مجاز فیلتر/مرتب‌سازی حفظ می‌شوند؛ پارامتر `category` تکرار مسیر از مقصد حذف می‌شود. انتخاب نامعتبر400 و داده پنهان/حذف‌شده404 است؛ حتی alias قدیمی هم وجود draft را با redirect افشا نمی‌کند.
- API عمومی، کد/slug یا ID قبلی را می‌پذیرد و در DTO همچنان ID داخلی را برای selection و خرید برمی‌گرداند. **POST redirect ندارد**. تغییر فرم/URL مجوز تازه‌ای ایجاد نمی‌کند؛ seller حق ویرایش دسته مشترک پیدا نکرده است.
- helper مشترک Go لینک خانه، ریل، تخفیف، PLP، جست‌وجو/پیشنهاد، دسته‌ها، breadcrumb، PDP و سبد را می‌سازد. لینک بنر در DTO عمومی canonical می‌شود؛ محتوای ذخیره‌شده و نسخه تاریخی بنر بازنویسی نمی‌شود. صفحه سفارش فعلی لینک محصول نداشت؛ هیچ لینک قدیمی تازه‌ای در آن تولید نشده است.
- HTML اولیه PDP/PLP و navigation در SPA، canonical جدید دارند. canonical انتخاب تنوع به صفحه پایه محصول اشاره می‌کند؛ noindex/nofollow محیط بررسی و تنظیم محصول پابرجاست. sitemap اجرایی قبلاً وجود نداشت و سرویس جداگانه‌ای در این کار اضافه نشد.
- قالب DTO و کلید Redis به **`cherriva:public:v3:`** ارتقا یافت. کلید محصول همچنان بر ID داخلی است؛ backfill در همان تراکنش نسخه‌های PostgreSQL را تازه می‌کند. پاک‌کردن Redis یا نیاز به روشن‌بودن آن شرط مهاجرت نیست. startup پس از restart خانه،۴PLP و۱۰۰PDP را بدون خطا گرم کرد.

## پایداری و محدوده مهاجرت

migration **000007** دو جدول `catalog_product_codes` و `catalog_category_urls` و sequence پایدار PostgreSQL اضافه می‌کند. شماره/public code و token دسته constraint یکتا دارند؛ نگاشت حذف‌شده پاک نمی‌شود. قفل همان ردیف aggregate، تخصیص و نگاشت را میان writerها سری می‌کند؛ sequence سقف رزروشده را نگه می‌دارد. ابتدا تمام شماره‌های واردشده رزرو و سپس محصولات بومی بالاتر از بزرگ‌ترین شماره رزروشده تخصیص می‌گیرند؛ شماره گمشده دوباره استفاده نمی‌شود. در داده فعلی آخرین شماره۳۹۸۴۵ است و نخستین محصول بومی بعدی۳۹۸۴۶ خواهد بود؛ این رقم ثابت در کد نیست. عدد حداقل شروع بومی۱۰۰۰ است؛ صفر پیشرو و سقف چهاررقمی وجود ندارد. واردات بعدیِ دارای شماره متعارض عمداً conflict می‌دهد.

startup و تمام writerهای repository از sync مشترک استفاده می‌کنند. backfill اولیه فقط آرایه‌های محصول/دسته را با دو فیلد آدرس به‌روز می‌کند و projection عمومی را می‌سازد؛ مسیر save هویت/دفتر مالی replay نمی‌شود. تغییر کد عمومی هم fingerprint تأیید محتوای فروشنده را عوض نمی‌کند؛ تغییر محتوای واقعی همچنان آن را عوض می‌کند.

در apply محلی، هر۹مقایسه پیش/پس موفق بود: سایر payload، محصولات بدون `public_code`، دسته‌ها بدون `slug`، کاربران/رمزها، نقش‌ها، grants، عضویت‌ها، تعداد رسانه و حجم باینری‌ها. سفارش/رزرو، سبد قبلی، آدرس، کیف پول/دفتر، بانک، روش‌های پرداخت/ارسال، قیمت/موجودی و import marker در همین payload محفوظ‌اند. قواعد پرداخت هنوز به ID داخلی دسته وابسته‌اند؛ ممنوعیت محلی طلا برای `gsp-cat-828/831/878` و زیرشاخه‌ها و سقف۳۰/۱۵میلیون تومان تغییر نکرده‌اند.

SHA فایل snapshot همچنان `c8154676f86f134c0ba98c7f53a7614e38f0c6cad2828b83e61b82eccdfbee98` است. مبدأ SQL Server خوانده نشد، JSON/manifest تاریخی تغییر نکرد، محتوای منتشرشده فروشگاه‌ها دست‌کاری نشد و بنر/کمپین مبدأ فعال نشد.

## اجرا برای لوکال یا سرور

اتصال همان محیط را به شکل خصوصی در `REVIEW_DATABASE_URL` یا `DATABASE_URL` قرار دهید؛ هر دو هم‌زمان ممنوع‌اند. مقادیر اتصال و backup خصوصی وارد گیت نشوند. برای سرور، backup کامل استاندارد دیتابیس و پنجره توقف writerها نیز لازم است؛ backup فرمان زیر فقط payload workspace است، نه نسخه کامل جدول‌های حساب/رسانه.

از `apps/api`، ابتدا با binary جدید و بدون اجرای سرویس قدیمی:

```sh
go run ./cmd/catalog-urls -mode review > public-url-dry-run.json
go run ./cmd/migrate -dir ../../db/migrations
go run ./cmd/catalog-urls -mode review -apply -backup /private/new-workspace-backup.json > public-url-apply.json
```

dry-run پیش از migration۷ هم با جدول‌های خالی فرضی کار می‌کند. در صورت conflict یا نام غایب، قبل از apply داده محتوایی را از مسیر مجاز اصلاح کنید؛ ابزار کدی را حدسی جایگزین نمی‌کند. حالت `live` فقط همراه متغیر خصوصی همان محیط مجاز است. apply نخست backup با فایل تازه و بدون overwrite می‌سازد؛ تکرار همان فرمان پس از اعمال، **`already_applied=true`** می‌دهد و حتی backup قبلی را بازنویسی نمی‌کند. تکرار واقعی و برابری۹کنترل در لوکال بررسی شد.

راه‌اندازی روزمره در PowerShell:

```powershell
cd D:\Cherriva
.\Start-Cherriva.ps1
```

این فرمان build، migration، backfill تکرارپذیر و warmup را اجرا می‌کند. پیش از startup نسخه جدید در محیطی که داده واقعی دارد، dry-run و backup بالا انجام شوند. در rollback عادی نگاشت‌ها و sequence را حذف نکنید؛ roll-forward اصلاح محتوا یا برگشت slug به alias متعلق به همان دسته، مخرب نیست. فایل down فقط برای rollback آگاهانه schema است و حذف نگاشت‌های صادرشده در محیط دارای لینک عمومی توصیه نمی‌شود. سرویس قدیمی را هم‌زمان writer نگذارید.

## آزمون و شواهد

| پذیرش | شاهد |
| --- | --- |
| AC01/02/04: نمونه chr و slug، یک301، انتخاب/query، canonical، visibility | تست HTTP واقعی `TestCatalogURLHTTPRedirectSEOAndHiddenAliases` و مرورگر |
| AC03/06: یکتایی، sequence، alias، عدم تصاحب، rollback و restart | `catalog_urls_test.go` در application و bootstrap؛ دو writer هم‌زمان؛ dry-run کامل و replay فرمان |
| AC05: حفظ داده و ارتباطات | ۹مقایسه مستقل metadata-only و snapshot SHA؛ تست موجود `TestCatalogSeedPostgresRestartAndRetry` و تست fingerprint محتوای منتظر |
| AC07: فرم و موبایل/دسکتاپ، گیت تحویل | اسکریپت مرورگر، تصاویر زیر، نتیجه CI/سورس در انتهای این سند |

- Go: **۲۰۹ تست/زیرتست موفق، بدون شکست**؛ PostgreSQL و Redis واقعی در schema/namespaceهای آزمون. یک تست اختیاری roundtrip رسانه۳GB اجرا نشد. `go vet ./...` موفق بود؛ race در Windows محلی اجرا نشده و نتیجه آن به CI لینوکس وابسته است.
- lint و build بک‌آفیس و storefront موفق؛ هشدار قبلی اندازه bundle پنل همچنان غیرمسدودکننده است.
- [۳۹کنترل مرورگر](../../../tests/e2e/catalog-public-urls.cjs)، عرض۱۴۴۰/۳۹۰، موفق و بدون runtime pageerror: API/redirect/canonical، خانه و بنر، breadcrumb، تنوع و تغییر انتخاب، سبد، slug/اعتبارسنجی فرم، نام انگلیسی محفوظ و کد فقط‌خواندنی. آزمون با context مهمان مستقل بود؛ سبد خود را خالی کرد، سفارش/پرداخت یا ویرایش محتوایی ثبت نکرد.
- startup/restart واقعی و `already_applied=true` بررسی شد؛ مسیرهای نمونه پس از restart معتبرند. آزمون importer replay روی schema موقت انجام شد؛ دانلود/ورود دوباره کل۳GB در دیتابیس اصلی محلی لازم نشد.

[PLP دسکتاپ](../../../output/playwright/issue-42-urls/plp-watch-desktop.png) · [PLP موبایل](../../../output/playwright/issue-42-urls/plp-watch-mobile.png) · [PDP ساعت](../../../output/playwright/issue-42-urls/pdp-watch-desktop.png) · [PDP موبایل](../../../output/playwright/issue-42-urls/pdp-earbuds-mobile.png) · [سبد](../../../output/playwright/issue-42-urls/cart-mobile.png) · [فرم slug](../../../output/playwright/issue-42-urls/backoffice-category-slug.png) · [کد محصول در پنل](../../../output/playwright/issue-42-urls/backoffice-product-code.png).

نمونه‌های اجرای محلی: [موبایل](http://localhost:5197/category/mobile)، [ساعت](http://localhost:5197/category/smart-watch)، [محصول۳۹۷۸۲](http://localhost:5197/product/chr-39782)، [محصول۱۸۱۸۹](http://localhost:5197/product/chr-18189)، [طلای نمونه](http://localhost:5197/product/chr-37354).

## تحویل Git

کامیت سورس: [1f4fdd0](https://github.com/MohamadParsa/marketplace/commit/1f4fdd01cafae849b0ff564d7c1c55d1c4bf4d2b). مدارک و نتیجه CI پس از ثبت PR در این بخش و [نقشه سورس/سند](../source-document-map.md) ثبت می‌شوند. Issue تا عبور گیت تحویل باز است. این تحویل فقط محلی است؛ مجوز/درگاه واقعی و انتشار لیارا را اعلام نمی‌کند.
