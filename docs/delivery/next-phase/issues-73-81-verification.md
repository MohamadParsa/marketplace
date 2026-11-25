---
id: DEL-NEXT-073-081
title: گزارش اجرای محلی روش فروش و گریدهای پنل
status: in-review
owner: eliassarband
updated: 2026-11-25
related: [FEAT-NEXT-012, FEAT-NEXT-020, API-SALES-METHODS-001, API-BACKOFFICE-GRIDS-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# نتیجه اجرای #73 و #81

مالک این دو تسک را انتخاب و نگاشت پرداخت/فروش، بازه و دقت نرخ و خروجی XLSX را صریح تأیید کرد. تغییرها در `D:\Cherriva` روی شاخه `codex/local-parsa-integration-20261124` محلی‌اند؛ دستور عدم پوش پابرجاست. وضعیت سند in-review است. [قرارداد روش فروش](../../architecture/contracts/sales-methods.md) · [قرارداد گرید](../../architecture/contracts/backoffice-grids.md) · [ماتریس ممیزی](backoffice-grid-audit.md).

## رفتار تحویل‌شده

- صفحه مستقل `/sales-methods`، نقش مشاهده/ویرایش، نگاشت چند پرداخت به یک فروش، درصد۰تا۱۰۰ و override دسته با نمایش نرخ ارث‌رسیده/منبع؛ صفر صریح و حذف override مستقل‌اند.
- resolver مشترک PDP/checkout؛ انتخاب پرداخت قیمت را تازه می‌کند و consent قبلی را پاک می‌کند. کوپن و اعتبار، روش انتخاب‌شده را بی‌صدا به نقدی برنمی‌گردانند. snapshot نرخ/مبلغ/منبع/نام، سفارش گذشته را از تغییر امروز جدا می‌کند.
- گرید مشترک با search، فیلترهای متنی ترکیبی، sort، count، پاک‌سازی، paging و انتخاب چندصفحه‌ای. فهرست سفارش/واریز/بازپرداخت هم از کارت‌های بی‌صفحه به گرید منتقل شد؛ dialog رسیدگی و دسترسی‌های قبلی محفوظ‌اند.
- XLSX واقعی تمام نتیجه یا منتخب، تا۱۰هزار ردیف، مجوز مستقل قابل انتساب، لغو، پیام خطا، متن امن و حذف اطلاعات حساس. خروجی JSON قدیمی گزارش حذف شد تا مسیر موازی بدون مجوز export نماند.

## شاهد معیارهای پذیرش

| معیار | شاهد |
| --- | --- |
| #73 AC01/02 | تست دامنه default، هم‌زمانی، والد/فرزند، صفر/حذف و سبد دو دسته؛ مرورگر فرم مستقل، انتقال نگاشت و نرخ دسته |
| #73 AC03 | آزمون دقیق۱۰۰۰+۴۰−۱۰۴=۹۳۶، سود مستقیم۲۰۰؛ مرورگر نقدی۱۰۰هزار،۲٪، کوپن۱۰٪ و اعتبار۱۰هزار با حمل۱۰هزار: مبلغ۱۰۱۸۰۰، کسری۹۱۸۰۰ |
| #73 AC04 | رد consent با نرخ تازه؛ تغییر نرخ پس از پذیرش بدون تغییر line snapshot؛ تأیید مالی با همان مبلغ؛ نمایش markup در dialog سفارش |
| #73 AC05 | نشست/مجوز، ورودی/نرخ/null،409، retry و concurrency، محدودیت طلا، جلوگیری از عبور با کیف پول کامل؛ PostgreSQL restart و تغییر dependency کش و پذیرش مسیر کد عمومی chr محصول |
| #81 AC01 | [تمام۲۷گرید و فهرست‌های دیگر](backoffice-grid-audit.md)، مالک شکاف‌های فرم/حجم مشخص است |
| #81 AC02 |۱۵۰کاربر ساختگی،۲۰ردیف در صفحه، فیلتر ترکیبی، انتخاب دو صفحه، پاک‌کردن و عرض۳۹۰؛ اکشن sticky و جست‌وجو/مرتب‌سازی مشترک |
| #81 AC03 | خروجی منتخب۲ردیف و تمام نتیجه۱۴۹ردیف؛ ZIP/XML/CRC معتبر؛ نام `=SUM(1,1)` سلول متن، بدون formula؛ حد۱۰هزار/۱۰۰۰۱، حجم/سلول و لغو |
| #81 AC04 | grant و revoke مجوز، نشست، سطح backoffice/seller، منع workspace ترکیبی فاقد مجوز یکی از فروشگاه‌ها؛ API export هیچ id/data دریافت/بازگشت نمی‌دهد و داده تازه‌ای برای دورزدن scope نمی‌خواند |

## اجرا و بازتولید

- `go test -json ./...` با PostgreSQL/Redis محلی در schemaهای جدا:۳۲۰تست/زیرتست موفق، صفر شکست.۴skip شامل snapshot بزرگ اختیاری و سه harness دستی؛ harness همین تحویل جدا اجرا شد. `go vet ./...` موفق؛ race محلی اجرا نشده است.
- `npm run build` هر دو app و کنترل TypeScript موفق‌اند. هشدار قبلی chunk بیش از۵۰۰KiB بک‌آفیس باقی است؛ dependency تازه اضافه نشد.
- [تست مولد](../../../tests/unit/grid-xlsx.cjs): از ریشه `node tests/unit/grid-xlsx.cjs` پس از نصب dependency فعلی بک‌آفیس؛۹کنترل موفق شامل۱۰هزار ردیف، حدها، XML، متن فرمول‌نما و لغو.
- **۲۲کنترل مرورگر موفق و صفر runtime pageerror**؛ خطای مجوز و لغو عمداً شبیه‌سازی شدند. [تست مرورگر](../../../tests/e2e/issues-73-81.cjs) با Playwright CLI و [harness حافظه‌ای](../../../apps/api/internal/bootstrap/sales_grid_browser_test.go). `CHERRIVA_SALES_GRID_BROWSER=1 go test ./internal/bootstrap -run '^TestSalesGridBrowserHarness$' -v -count=1 -timeout 16m`، سپس از ریشه `playwright-cli -s=cherriva-sales-grid run-code --filename tests/e2e/issues-73-81.cjs`؛ ابتدا باید session مرورگر باز باشد. پورت آزمون۸۰۹۷/۵۱۹۹ و داده‌های مالی آن صرفاً حافظه‌ای‌اند؛ به دیتابیس اصلی یا SQL Server مبدأ وصل نمی‌شود.
- آزمون‌های سرور: [دامنه](../../../apps/api/internal/application/backoffice/sales_methods_test.go) و [HTTP/مجوز/PG](../../../apps/api/internal/bootstrap/sales_grid_test.go). [خلاصه ماشین‌خوان](issues-73-81-checks.json).

## تصاویر و نمونه خروجی

[روش‌های فروش](../../../output/playwright/issues-73-81/sales-methods-desktop.png) · [نرخ دسته](../../../output/playwright/issues-73-81/category-inherited-rate.png) · [خرید](../../../output/playwright/issues-73-81/checkout-markup-desktop.png) · [جزئیات سفارش](../../../output/playwright/issues-73-81/order-markup-detail.png) · [گرید دسکتاپ](../../../output/playwright/issues-73-81/grid-filters-desktop.png) · [موبایل](../../../output/playwright/issues-73-81/grid-mobile.png).

[XLSX منتخب](../../../output/playwright/issues-73-81/selected.xlsx) · [تمام نتیجه](../../../output/playwright/issues-73-81/filtered.xlsx) · [نام فرمول‌نما](../../../output/playwright/issues-73-81/formula-as-text.xlsx). همه داده نمونه ساختگی‌اند؛ فایل‌ها با parser ZIP/XML مستقل بررسی شده‌اند، بازشدن در Excel واقعی ادعا نمی‌شود.

## محدودیت‌های واقعی

صفحه‌بندی/فیلتر مشترک فعلاً client-side روی داده scoped دریافت‌شده است؛ حذف workspace کامل/فیلترهای تخصصی پنج فرم در #108/#111/#113/#115/#116 و انتقال نقش‌ها در #112 باقی است. #81 ممیزی/زیرساخت مشترک را تکمیل می‌کند و آن تسک‌ها را بسته نمی‌کند. خروجی عددها در این نسخه متن قالب‌بندی‌شده است؛ برای تحلیل عددی در Excel باید تبدیل شود. scope جامع سلر #71، PlaceOrder جامع #70، فرمول طلا #72، مجوز فروشگاه #114 و BNPL واقعی نیز تحویل این دو نیستند.

نرخ۴٪/۲٪، کوپن و شارژ تستی فقط در harness اعمال شدند. برنامه اصلی با داده موجود و fallback نقدی۰٪ بالا می‌آید؛ بانک/دفتر/کاربران/رمزها/نقش‌ها بازنویسی نشده‌اند. ۱۷بخش حساب/نقش/عضویت/سفارش/دفتر/منابع اعتبار/تنظیمات شامل مقصد بانکی، قبل و بعد از restart با hash یکسان تطبیق داده شدند؛ اطلاعات اتصال در گزارش نیامده است. وضعیت بستن Issue در رسید تحویل ثبت می‌شود؛ پوش/main/لیارا انجام نشده‌اند.

## رسید سورس و نسخه اصلی بررسی

کامیت سورس محلی `8b0351893dedc248ef29fdf99fcfa5fb15d2f92a` شامل کد، تست و تصاویر است؛ پوش نشده. بررسی نهایی نمونه [chr-10270](http://localhost:5197/product/chr-10270) و API قیمت آن۲۰۰، روش نقدی فعلی۰٪ و API خصوصی بدون نشست۴۰۱ بود. اصلاح ردیف مستقل مارکاپ در dialog سفارش نیز پس از build آخر با مرورگر بررسی شد. [رسید کامنت‌ها و وضعیت گیت](issues-73-81-delivery.json).

هر دو Issue با کامنت شواهد و دلیل completed بسته شدند؛ والد #95 و انتشار #105 بازند. اطلاعات پوش‌نشدن سورس در شرح/کامنت هر دو صریحاً ثبت شده است. لینک کامنت‌ها در رسید تحویل بالاست.
