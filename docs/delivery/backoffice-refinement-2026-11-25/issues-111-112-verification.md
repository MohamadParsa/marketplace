---
id: DEL-BO-USERS-ROLES-001
title: گزارش اجرای محلی کاربران و نقش‌ها
status: in-review
owner: eliassarband
updated: 2026-11-25
related: [FEAT-BO-REFINEMENT-005, FEAT-BO-REFINEMENT-006, API-USER-MANAGEMENT-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# تحویل #111 و #112

مالک این دو Issue را برای اجرا انتخاب کرد؛ انتخاب مجوز پوش یا لیارا نیست. سورس در `D:\Cherriva` روی شاخه `codex/local-parsa-integration-20261124` است. [قرارداد و تصمیم‌ها](../../architecture/contracts/user-management.md)؛ وضعیت سند in-review می‌ماند.

## رفتار ساخته‌شده

- صفحه `/users`: ایجاد و ویرایش همان حساب واحد، نام/نام خانوادگی جدا، ایمیل اختیاری یکتا، جنسیت و active، رمز اولیه/بازنشانی موقت؛ موبایل کاربر موجود فقط خواندنی. نام‌های قدیمی بدون split و کد ملی بدون تغییر باقی‌اند.
- فیلتر ترکیبی نام/خانوادگی/موبایل/ایمیل/وضعیت/جنسیت/نقش/فروشگاه، مرتب‌سازی و paging سرور؛ پاسخ حداکثر۱۰۰ردیف و پیش‌فرض۲۰. پوسته جدید roster کامل کاربران را از workspace نمی‌گیرد؛ endpoint قدیمی برای سازگاری حفظ شده است.
- خروجی مشترک XLSX از تمام صفحات همان نتیجه یا انتخاب چندصفحه‌ای؛ فقط سلول‌های نمایشی مجاز، حداکثر۱۰هزار. مجوز export و دامنه فیلتر در سرور نیز بررسی می‌شوند. شارژ کیف پول روی ردیف همان فرم حفظ شده است.
- صفحه مستقل `/roles`: ایجاد/ویرایش/حذف امن نقشِ بدون استفاده، تفکیک global/store، جست‌وجوی مجوزها با توضیح فارسی. فرم تعریف نقش از کاربران حذف شد.
- نقش پایه، grant و عضویت چندفروشگاهی در dialog کاربر هستند؛ scopeها و منبع مجوز مؤثر جدا دیده می‌شوند. ارتقای admin از تأیید OTP قبلی استفاده می‌کند؛ آخرین ادمین، حذف نقش مصرف‌شده و نسخه قدیمی در سرور محافظت دارند.
- نسخه کاربر در ذخیره/انتساب و نسخه تعریف نقش در grant جدید کنترل می‌شوند. رمز تازه ادمین epoch نشست را تغییر می‌دهد؛ actor و scope در تراکنش دوباره بررسی می‌شوند.

## سیاست تازه تأییدشده مالک

حساب دارای رمز موقت با **OTP هم مانند رمز** قبل از هر دسترسی عادی رمز شخصی تعیین می‌کند؛ reset ادمین همین سیاست و ابطال نشست‌های قبلی را دارد. این تصمیم رمز را برای کاربران عادی OTPدار اجباری نمی‌کند. token محدود اجازه خواندن پروفایل/پنل/سبد را ندارد؛ فقط تکمیل رمز و خروج مجاز است. فرم مشترک سه محیط، رمز و تکرار آن را می‌گیرد و پس از تغییر موفق token جدید دریافت می‌شود.

رمز خام در DB/GET/log/URL/export ذخیره یا برگردانده نمی‌شود؛ retry check رمز با bcrypt خصوصی و fingerprint غیررمزی است. نمونه ثابت رمز برای حساب عملیاتی اضافه نشده. مشخصات نمونه آزمایش فقط در fixture حافظه‌ای است.

## معیار پذیرش و شواهد

| معیار | شاهد اجرا |
| --- | --- |
| #111 AC01 | یک ID برای مشتری/کارمند/عضو دو فروشگاه؛ ساخت/ویرایش و grant در UI؛ تست HTTP همان UserID |
| #111 AC02 | ایمیل/موبایل تکراری، قالب نامعتبر، تغییر موبایل، تزریق permissions، نبود مجوز و نسخه قدیمی؛ concurrency دو ذخیره با یک version فقط یک موفقیت |
| #111 AC03 | ورود با ایمیل/رمز و OTP در حالت موقت، API خصوصی۴۰۱، منع استفاده دوباره از رمز موقت، تکمیل اتمیک و ابطال token قدیم؛ reset هم نشست قبلی را باطل می‌کند؛ DTO/workspace فاقد credential/retry-private |
| #111 AC04 | migration افزایشی، PostgreSQL restart و حفظ نام قدیمی/کد ملی؛۱۲۵نتیجه آزمون فیلتر و۲۰ردیف صفحه؛ export همه صفحات و اشتراک منتخب |
| #112 AC01 | منوی مستقل نقش‌ها؛ نبود فرم تعریف نقش در کاربران؛ همه انتساب‌ها در فرم همان کاربر |
| #112 AC02 | نقش چندمجوزی و عضویت دو فروشگاه؛ store A حق store B نمی‌دهد؛ union در scope، سلب دسترسی فوری و مجوز فعلی سرویس کنترل می‌شوند |
| #112 AC03 | نقشِ دارای grant قابل حذف نیست؛ scope/عضویت نامعتبر و نسخه نقش/کاربر stale رد می‌شوند؛ غیرادمین نمی‌تواند نقش بسازد |

مرجع تست: [HTTP/PG/مجوز](../../../apps/api/internal/bootstrap/users_management_test.go)، [هم‌زمانی](../../../apps/api/internal/application/backoffice/users_management_test.go)، [رگرسیون migration](../../../apps/api/internal/bootstrap/review_postgres_test.go) و [مرورگر](../../../tests/e2e/users-roles.cjs). [سناریوی مستند BDD](../../bdd/features/experiences/user-management-and-roles.feature) runner مستقل اجراشده نیست.

## اجرای آزمون

از `apps/api`، `go test -json ./...` با PostgreSQL/Redis محلی و schemaهای ایزوله؛ fixture اصلی بررسی یا SQL Server مبدأ در این تست‌ها دست‌کاری نمی‌شود. تست snapshot بزرگ اختیاری و harnessهای دستی جدا گزارش می‌شوند. `go vet ./...` و build/TypeScript هر دو app نیز اجرا می‌شوند؛ هشدار قبلی chunk بک‌آفیس بیش از۵۰۰KiB باقی است، dependency تازه نصب نشده است.

مرورگر از harness حافظه‌ای موجود `TestSalesGridBrowserHarness` روی۸۰۹۷/۵۱۹۹ استفاده می‌کند تا زیرساخت fixture موازی ساخته نشود. پس از build هر دو app، `CHERRIVA_SALES_GRID_BROWSER=1 go test ./internal/bootstrap -run '^TestSalesGridBrowserHarness$' -v -count=1 -timeout 16m` و از ریشه `playwright-cli -s=cherriva-users-roles run-code --filename tests/e2e/users-roles.cjs`. همه role/user/passwordهای آزمایشی این مسیر در حافظه‌اند و رمزهای مرورگر هر بار تصادفی ساخته می‌شوند.

## مرز و محدودیت

خواندن داخلی identity snapshot هنوز همه هویت‌ها را بارگیری می‌کند؛ API جدید نتایج را در سرور فیلتر/صفحه‌بندی می‌کند، ولی query ایندکس‌دار SQL و بهینه‌سازی جامع workspace تحویل نشده است. scope جامع فروشنده در #71 و سایر فرم‌های #107تا#117 با این دو Issue done نمی‌شوند. تغییر موبایل، OTP ایمیلی، اتصال بیرونی و انتشار لیارا خارج این تحویل‌اند.

کنترل authmemory/مرورگر با پذیرش provider واقعی متفاوت است؛ race محلی ادعا نمی‌شود. نتیجه نهایی آزمون، حفظ داده پس از restart و رسید کامیت/کامنت در همین گزارش ثبت می‌شود. تغییرهای تازه طبق دستور مالک محلی می‌مانند.

## اجرای نسخه بررسی

برای اجرای این نسخه از ریشه `D:\Cherriva`، `Start-Cherriva.ps1` و پس از build قبلی `Start-Cherriva.ps1 -NoBuild` را اجرا کنید؛ migration۰۰۰۰۱۲ پیش از اجرای API اعمال می‌شود. صفحات بررسی: [کاربران](http://localhost:8098/users) و [نقش‌ها](http://localhost:8098/roles). حساب محلی تیم همان حساب [راهنمای دسترسی](../local-review-access.md) است؛ رمز/موبایل عملیاتی تازه در این تحویل تعریف نشده است.

## نتیجه نهایی کنترل‌ها

**۳۲۵تست/زیرتست Go موفق، صفر شکست؛۴skip** شامل snapshot بزرگ اختیاری و سه harness دستی، که harness این تحویل جدا اجرا شد. `go vet` موفق و lint/build/TypeScript هر دو app موفق است. **۳۰کنترل مرورگر موفق و صفر runtime pageerror**؛ خطای رمز/ورودی/نشست محدود عمداً آزموده شد.

فایل XLSX منتخب دقیقاً۲ردیف از دو صفحه متفاوت و تمام نتیجه۱۶۰ردیف دارد؛ ZIP/XML/CRC با parser مستقل تأیید و نبود formula و ستون حساس بررسی شد. [خلاصه ماشین‌خوان](issues-111-112-checks.json). تست مرورگر قدیمی #73/#81 نیز با کنترل‌های جدید فهرست کاربران همگام شد؛ تعداد تست جدید از اجرای آن دوباره شمرده نشده است.

پس از اجرای migration و restart واقعی نسخه محلی،۱۷بخش حساب/هش رمز/نقش/عضویت/سفارش/دفتر/منابع اعتبار و تنظیمات با قبل hash یکسان داشتند؛ فقط ستون‌های افزایشی از hash مقایسه قدیمی کاربران کنار گذاشته شدند. تست PostgreSQL جدا، ماندگاری ستون‌های تازه و حفظ کامل اشیای قبلی و موجودی غیرصفر آزمایشی را نیز بررسی کرد. اتصال و داده حساس در گزارش نیامده است.

بررسی نسخه اصلی محلی: query کاربران۲۰۰ و bounded، roster در پوسته حذف، آرایه عضویت همه ردیف‌ها معتبر و حالت رمز ادمین موجود همان قبل است. حساب اولیه تیم دوباره ساخته یا رمز آن تغییر داده نشد.

## تصاویر و فایل‌ها

[فیلتر کاربران](../../../output/playwright/issues-111-112/user-filters-desktop.png) · [عضویت و نقش در فرم کاربر](../../../output/playwright/issues-111-112/user-memberships-and-roles.png) · [مجوزهای نقش سراسری](../../../output/playwright/issues-111-112/global-role-permissions.png) · [نقش فروشگاه](../../../output/playwright/issues-111-112/store-role-permissions.png) · [گرید موبایل](../../../output/playwright/issues-111-112/users-mobile.png) · [نقش موبایل](../../../output/playwright/issues-111-112/role-dialog-mobile.png) · [رمز موقت بک‌آفیس](../../../output/playwright/issues-111-112/temporary-password-backoffice.png) · [رمز موقت سایت](../../../output/playwright/issues-111-112/temporary-password-storefront.png) · [فعال‌سازی ادمین با OTP](../../../output/playwright/issues-111-112/admin-reactivation.png).

[خروجی منتخب](../../../output/playwright/issues-111-112/users-selected.xlsx) · [تمام نتیجه](../../../output/playwright/issues-111-112/users-filtered-all.xlsx). داده‌ها ساختگی و تصاویر بازبینی بصری شده‌اند؛ بازکردن در برنامه Excel واقعی ادعا نمی‌شود. Source/گزارش محلی‌اند و پوش/لیارا انجام نشده است.

## ثبت تحویل

سورس و تست‌ها در کامیت محلی `37ec82c31408d69736a4f7826fa5ae07ec905077` ثبت شدند. [رسید GitHub](issues-111-112-delivery.json): [کامنت #111](https://github.com/MohamadParsa/marketplace/issues/111#issuecomment-7159606229) و [کامنت #112](https://github.com/MohamadParsa/marketplace/issues/112#issuecomment-7159608857)؛ هر دو با دلیل completed بسته شدند. فهرست والد #95 اصلاح و باز ماند؛ انتشار #105 جداست. طبق دستور مالک، **هیچ push/merge/deploy انجام نشد** و پیوند commit روی GitHub تا زمان پوش معتبر نیست. کامیت مدارک با trailer `Source-Commit` به این سورس متصل می‌شود.


## انتشار بعدی به دستور مالک — ۲۰۲۶-۱۱-۲۶

دستور تازه پوش پس از تکمیل، سیاست فقط‌لوکال قبلی را برای تحویل‌های آماده جایگزین کرد. سورس `37ec82c31408d69736a4f7826fa5ae07ec905077` و مدارک این دو Issue همراه شاخه [PR #106](https://github.com/MohamadParsa/marketplace/pull/106) پوش شدند؛ کامنت به‌روزرسانی در هر دو Issue بسته ثبت شد. بخش‌های «پوش‌نشده» بالا تاریخچه زمان تحویل اولیه‌اند؛ استقرار و ادغام همچنان جداست.
