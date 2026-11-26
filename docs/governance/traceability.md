# ردیابی قابلیت تا پذیرش

پیگیری۲۰۲۶-۱۱-۲۳: [FEAT-SF-CACHE-001](../delivery/storefront-followups/cache-policy.md) و [FEAT-CATALOG-PUBLIC-URL-001](../delivery/storefront-followups/public-catalog-urls.md) ← [صف و مسئولیت](../delivery/storefront-followups/README.md). سیاست کش با تأیید مالک محلی اجرا شد؛ [شاهد #41، کنترل‌های حفاظت و تصاویر](../delivery/storefront-followups/issue-41-verification.md). مهاجرت URL در #42 پس از پایان #41 است و هنوز اجرا نشده است.

برش فعال‌شده CV-R2-07: [ADR-006](../decisions/adr-006-responsive-banner-activation.md) ← [شرح](../delivery/review-v1-1/tasks/cv-r2-07-responsive-banner.md) ← [قرارداد](../architecture/contracts/responsive-banners.md) ← [دو سناریو](../bdd/features/experiences/responsive-banners.feature) ← [گزارش و تصاویر](../delivery/review-v1-1/issue-7-verification.md). فعال‌شدن این برش، سایر CMS را فعال نمی‌کند.

برش آزمون CV-R2-06: [شرح](../delivery/review-v1-1/tasks/cv-r2-06-product-workspace.md) ← [قرارداد](../architecture/contracts/product-workspace.md) ← [سه سناریو](../bdd/features/experiences/product-workspace.feature) ← [گزارش API/مرورگر و تصاویر](../delivery/review-v1-1/issue-6-verification.md). پایان برش آزمون با تکمیل محصول، ادغام یا پذیرش اتصال واقعی یکی نیست.

برش آزمون CV-R2-05: [شرح کار](../delivery/review-v1-1/tasks/cv-r2-05-shared-attributes.md) ← [قرارداد](../architecture/contracts/catalog-attributes.md) ← [سه سناریوی BDD](../bdd/features/experiences/shared-attributes.feature) ← [گزارش و شواهد اجرا](../delivery/review-v1-1/issue-5-verification.md). وضعیت بازبینی کد از وضعیت تأیید مشخصات جداست.

این جدول فقط نقشه لینک‌هاست؛ قانون و سناریو در محل مرجع نگهداری می‌شوند. بخش‌هایی از بک‌آفیس برای مرور با داده آزمون پیاده شده‌اند؛ وضعیت دقیق و شواهد در گزارش مرحله ۱ آمده است. شناسه‌های `AC` داخل سند قابلیت، شواهد طراحی پذیرش‌اند، نه گزارش آزمون موفق. صورت Gherkin همین سناریوها در [مستندات BDD](../bdd/README.md) نگهداری می‌شود.

| قابلیت / پذیرش همان سند | مرجع قواعد | عملیات قرارداد منطقی | کنترل کیفیت مرتبط | مرحله |
| --- | --- | --- | --- | --- |
| [FEAT-SF-001 — کشف](../experiences/storefront/discovery.md) | [catalog](../domains/catalog/README.md)، [pricing](../domains/pricing/README.md) | OP01–OP03 | Q01، Q06، Q13 | بعد |
| [FEAT-SF-002 — خرید](../experiences/storefront/checkout.md) | [checkout](../domains/checkout/README.md)، [identity](../domains/identity/README.md)، [shipping](../domains/shipping/README.md)، [payments](../domains/payments/README.md) | OP02–OP08؛ OP03 شامل تغییر ترکیب و فروشگاه قلم | Q03، Q04، Q06، Q07، Q12 | بعد |
| [FEAT-SF-003 — پیگیری/لغو](../experiences/storefront/orders.md) | [orders](../domains/orders/README.md)، [shipping](../domains/shipping/README.md)، [payments](../domains/payments/README.md) | OP09، OP21–OP22 و خواندن خرید مجاز | Q01، Q05، Q09 | بعد |
| [FEAT-SF-004 — کیف پول](../experiences/storefront/wallet.md) | [payments](../domains/payments/README.md) | OP07، OP14 و گردش مالی مجاز | Q01، Q03، Q07، Q10، Q14 | بعد |
| [FEAT-SP-001 — عرضه](../experiences/seller-portal/catalog.md) | [catalog](../domains/catalog/README.md) | OP10، OP17–OP18 | Q01، Q04، Q11 | اول |
| [FEAT-SP-002 — تأمین/ارسال](../experiences/seller-portal/fulfillment.md) | [orders](../domains/orders/README.md)، [shipping](../domains/shipping/README.md)، [notifications](../domains/notifications/README.md) | OP11–OP12، OP19 | Q01، Q02، Q05، Q08 | اول |
| [FEAT-BO-001 — مدیریت](../experiences/backoffice/management.md) | [catalog](../domains/catalog/README.md)، [pricing](../domains/pricing/README.md) | OP16–OP20 | Q07، Q09، Q11 | اول |
| [FEAT-BO-002 — رسیدگی](../experiences/backoffice/support.md) | [orders](../domains/orders/README.md)، [payments](../domains/payments/README.md)، [shipping](../domains/shipping/README.md)، [notifications](../domains/notifications/README.md) | OP13–OP14، OP21–OP22 | Q01، Q05، Q09 | اول |
| [FEAT-BO-003 — مالی](../experiences/backoffice/finance.md) | [payments](../domains/payments/README.md)، [settlements](../domains/settlements/README.md) | OP14–OP15 | Q03، Q07، Q09 | اول |

`OP*` به [API-MARKETPLACE-001](../architecture/contracts/interactions.md) و `Q*` در ستون آخر به [NFR-INDEX-001](../architecture/quality.md) اشاره دارد، نه پرسش‌های محصول. نام کامل کنترل مثلاً `NFR-INDEX-001-Q03` است.

## بررسی‌های میان‌حوزه‌ای پیش از انتشار

این موارد سناریوی ترکیبی‌اند و جای سناریوهای واحد حوزه را نمی‌گیرند:

- آخرین واحد کالا + دو جلسه پرداخت + callback دیررس: موجودی و اثر مالی هم‌زمان درست بمانند؛ نتیجه وابسته به Q-015 است.
- شارژ موفق + انقضای جلسه: شارژ گم نشود، سفارش بی‌دلیل موفق نشود و مسیر بعدی مشتری روشن بماند؛ Q-022.
- تغییر قیمت پایه + تمدید خودکار + پایان اعتبار میان جلسه: ساعت‌ها و تثبیت قانون مطابق حوزه‌های قیمت‌گذاری/checkout باشند؛ مبدأ و چندتب در Q-012 باقی مانده است.
- خرید چندفروشگاهی + عدم تأمین یک قلم + ارسال فروشگاه دیگر + لغو مشتری: قلم ارسال‌شده وارد لغو نشود و refund هر قلم قابل ردیابی بماند.
- تأمین جایگزین + تسویه + refund: ارسال مستقیم جایگزین، مبلغ ثابت مشتری، اختلاف هزینه پلتفرم و سهم/کمیسیون قرارداد تأمین‌کننده واقعی پاک یا دوبار ثبت نشوند؛ جزئیات Q-020/Q-024 باقی است.
- چند شارژ از کارت‌های مختلف + مصرف + دو درخواست خروج هم‌زمان: منشأ و ظرفیت‌ها نقض نشوند؛ Q-023.
- خروج از حساب PWA + ورود کاربر دیگر: اطلاعات آدرس/سفارش/کیف پول قبلی قابل بازیابی نباشد.

پس از آغاز ساخت، شواهد اجرا و لینک کار/آزمون کنار قابلیت مربوط ثبت شود؛ پیش از حل پرسش‌ها نتیجه قبولی فرض نشود.

## شواهد نسخه بررسی بک‌آفیس

نسخه صفحات BO-01 تا BO-12 و عملیات آزمون موجود در [گزارش مرحله ۱](../delivery/stage-1-progress.md) و [قرارداد API نسخه بررسی](../architecture/contracts/backoffice-review.md) ثبت شده‌اند. شواهد کد/آزمون و محدودیت هر بخش در همان گزارش آمده‌اند؛ هیچ‌یک جای گیت پذیرش کامل مالی/سفارش یا اجرای کل BDD را نمی‌گیرد.

## شواهد رابط نمونه سایت مشتری

تصاویر کارت، مودال و شمارنده نسخه ۱۶ نوامبر به‌عنوان سابقه نسخه بررسی در [گزارش بررسی سایت](../delivery/storefront-review.md#اجرای-طراحی-بصری--۲۰۲۶-۱۱-۱۶) پیوند خورده‌اند.

[گزارش بررسی Cherriva](../delivery/storefront-review.md) بخش‌های نمایشی FEAT-SF-001 و سبد FEAT-SF-002 را به [آزمون مرورگر](../../tests/e2e/storefront-smoke.py) پیوند می‌دهد. قیمت/موجودی fixture هستند؛ این آزمون جای پذیرش اتصال backend، اعتبار قیمت، احراز هویت یا پرداخت نیست. FEAT-SF-003/004 در این انتشار پیاده نشده‌اند.

ادامه درخواست در ۲۰۲۶-۱۱-۱۵: [آزمون جریان تا پرداخت](../../tests/e2e/storefront-checkout-smoke.py)، حفظ سبد پس از ورود نمونه (بخش رابط AC01)، ارسال دو فروشگاهی (بخش رابط AC02)، نمایش مسیر شارژ/کسری (بخش رابط AC03) و انقضا/بازگشت به سبد (بخش رابط AC04) را پوشش می‌دهد. تأیید واقعی شارژ، برداشت، رزرو، ادغام سبد حساب و ثبت سفارش هنوز شواهد پذیرش ندارند؛ وضعیت کل FEAT-SF-002 همچنان `in-review` است.


## بسته مستقل طراحی و تحلیل برای بازبینی — ۲۰۲۶-۱۱-۱۹

[بسته۱.۱](../review-packages/README.md) طرح‌های سایت۰.۴ و پنل۰.۲، تصاویر و شناسه‌های CV پایه/اصلاحی را بدون تغییر نگه می‌دارد. وضعیت in-review و ترتیب‌های پیشنهادی آن، اولویت نهایی تیم یا done نیستند. اختلاف با اسناد جاری در REVIEW-PKG-Q01 تا Q04 همان فهرست ثبت شده است؛ پس از بررسی مشترک به پرسش‌ها و مشخصات اصلی منتقل شود. [ردیابی داخلی بسته](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/traceability-fa.md) مرجع ارتباط صفحه، قاعده و تسک همین snapshot است؛ قواعد فعلی ریپو با انتشار آن خودکار بازنویسی نشده‌اند.

در ۲۰۲۶-۱۱-۲۰ طرح A «متعادل و شیک» برای ظاهر سایت مشتری انتخاب شد. نمونه‌های خانه، فهرست و جزئیات به `FEAT-SF-001` و نمونه‌های سبد، ورود، ارسال و پرداخت به `FEAT-SF-002` مرتبط‌اند. مسیر مستقیم از کارت خانه و فهرست به صفحه محصول در `Q-027` ثبت شد و سناریوهای وابسته با آن هماهنگ شدند. [طرح منتخب و تصاویر](../experiences/storefront/visual-design.md) · [کشف کالا](../experiences/storefront/discovery.md) · [سبد تا نتیجه](../experiences/storefront/checkout.md) · [Q-027](../product/open-questions.md#q-027)

[طرح پنل نسخه ۰.۲](../experiences/backoffice/panel-design.md) صفحات بسته را بر اساس موضوع به `FEAT-BO-001/002/003` و `FEAT-SP-001/002` پیوند می‌دهد. [انتخاب جهت پنل](../product/open-questions.md#q-028) و [اختلاف‌های هویت، نقش، خصوصیت و زمان آماده‌سازی](../product/open-questions.md#اختلافهای-بسته-بازبینی-پنل) در دفتر محصول قابل پیگیری‌اند.

## تحلیل پایه و معیارهای پذیرش

۹۱ تسک پایه و ۲۹ ردیف پوشش از طریق [نقشه تسک‌ها](../delivery/analysis-backlog.md) به حوزه‌ها و قابلیت‌های جاری پیوند دارند. [جدول گذار وضعیت](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/18-state-transitions.md) نمونه‌های ترکیبی رزرو، پرداخت، تأمین، اعزام، استرداد و تسویه را به شرط انجام و رفتار تکرار وصل می‌کند. [دفتر پرسش‌های تحلیل پایه](../product/open-questions.md#پرسشهای-تحلیل-پایه) وضعیت ۱۴ شناسه CHQ را در کنار تصمیم‌های ریپو ثبت می‌کند.


## اصلاحات جاری نسخه ۱.۱

| قاعده بسته | سند جاری |
| --- | --- |
| R2-01/02 | [هویت و ورود](../domains/identity/README.md) · [سبد و ورود](../experiences/storefront/checkout.md) |
| R2-03/04 | [آیکن و خصوصیت مشترک](../domains/catalog/README.md) |
| R2-05/06 | [فضای محصول و سکشن](../experiences/backoffice/management.md) |
| R2-07 | [وقفه فروشنده و ثبت وعده](../domains/shipping/README.md) |
| R2-08 | [تجربه سایت](../experiences/storefront/README.md) · [راهنمای شروع و تحویل یکپارچه](../review-packages/cherriva-2026-11-17-v1-1/START-HERE-fa.md) |

[۱۲ اصلاح و تسک‌های پایه](../delivery/analysis-backlog.md#اصلاحات-نسخه-۱۱) · [پرسش‌های تکمیلی](../product/open-questions.md#پرسشهای-اصلاحات-جاری)

## پاسخ #13 و مسیر انتقال به مرحله۲

مرجع: [ADR-005](../decisions/adr-005-review-v1-1-product-responses.md)؛ تاریخ۲۰۲۶-۱۱-۲۰. وضعیت ردیف‌های زیر ثبت تصمیم/دامنه است، نه اجرای نرم‌افزار.

| تصمیم | قواعد/تجربه جاری | کار متأثر و شاهد بعدی |
| --- | --- | --- |
| موبایل نخست، OTP بدون رمز و لینک رمز پس از موبایل، بازیابی | [هویت](../domains/identity/README.md) | #1..3؛ AC10 و قرارداد بازیابی؛ تغییر موبایل هنوز پیشنهاد |
| admin، مجموع مجوز scoped، عضویت چندبه‌چند | [نقش‌ها](../product/roles.md) | [CV-R2-01 / #1](../delivery/review-v1-1/tasks/cv-r2-01-single-user.md#نتیجه-برش-آزمون): هویت/مالکیت خرید آزمون، مرز فروشگاه، نقش سفارشی، سیاست ادمین و ماندگاری PostgreSQL شاهد دارند. #2/#12 همچنان متأثرند. |
| خصوصیت مشترک بدون ارث و سقف عمومی کارت | [کاتالوگ](../domains/catalog/README.md) | #5/#6/#11؛ شارژر حافظه والد را نگیرد، سقف کارت از تنظیم بیاید |
| CMS معوق، خانه ثابت؛ تنظیمات عمومی مستقل | [مدیریت](../experiences/backoffice/management.md) | #7..9 معوق؛ dependency آن‌ها در #11/#12 در مرحله۲ اصلاح شود |
| اسلات ظرفیت‌دار/حمل حداقل‌حداکثر، تقویم و snapshot | [ارسال](../domains/shipping/README.md) و [سفارش](../domains/orders/README.md) | #10/#11 و کار مستقل ظرفیت؛ قبل از شاهد هم‌زمانی و سیاست رزرو، پذیرش اسلات ادعا نشود |
| فقط سایت اصلی، زیرسایت/نشست بین‌میزبانی معوق | [دامنه](../product/scope.md) | #11 و SEO؛ ادامه ورود و خرید همان سایت پابرجاست |
| ناموجود و درخواست اطلاع همان تنوع | [کشف](../experiences/storefront/discovery.md#رفتار-ناموجود-و-درخواست-اطلاع)، [اطلاع‌رسانی](../domains/notifications/README.md) | #11/#12 و کار مستقل اعلان؛ سیاست مقصد/تکرار باز |
| A پنل و فروشنده | [طرح پنل](../experiences/backoffice/panel-design.md) | Q-028 حل‌شده؛ پذیرش بصری مطابق A |

[وضعیت سؤال‌ها](../product/open-questions.md#پرسشهای-اصلاحات-جاری) و [نقشه اثر بر تسک‌ها](../decisions/adr-005-review-v1-1-product-responses.md#اثر-بر-تسکها-و-مرحله-بعد) مبنای به‌روزرسانی #14 هستند. فایل‌های CV و گزارش‌های snapshot و شواهد نسخه آزمون قبلی بازنویسی نشده‌اند؛ معنای آن‌ها باید با تصمیم جدید و دامنه تاریخ‌دارشان خوانده شود.

## همگام‌سازی Issueها پس از پاسخ #13

[ماتریس پاسخ→Issue→سند](../delivery/review-v1-1/README.md#پاسخ-به-issue-و-سند-متأثر) مرجع بند اول #14 است. هر یک از ۱۵ شرح کار نیز پاسخ، مرجع، dependency، بخش مسدود و AC قابل ارزیابی دارد. source snapshot تغییر نکرده و شناسه‌های CV-R2-01..12 حفظ شده‌اند؛ CV-R2-13/14/15 شناسه کارهای جدید #15/#16/#17 هستند و با شماره Issue فرایندی #13/#14 یکی نیستند.

[گزارش کنترل همگام‌سازی](../delivery/review-v1-1/verification.md) بررسی لینک/گراف و تطبیق متن Issueها را از شاهد اجرای محصول جدا نگه می‌دارد.

## شواهد پیاده‌سازی CV-R2-04

AC01/02/03 آیکن دسته/برند به [قرارداد](../architecture/contracts/catalog-icons.md)، [گزارش بررسی](../delivery/review-v1-1/issue-4-verification.md) و [تست API و مرورگر](../../tests/e2e/catalog-icons.cjs) متصل‌اند. وضعیت review و منتظر ادغام؛ Go tests/vet و frontend lint/build و۱۴ بررسی مرورگر موفق‌اند. storage حافظه‌ای آزمون است و قبولی production اعلام نشده است.


## تکمیل محتوای کاتالوگ — ۲۰۲۶-۱۱-۲۱

درخواست مالک ← [تطبیق و نیازمندی‌ها](../domains/catalog/content-and-seo.md) ← [CV-R2-16 / #24](../delivery/review-v1-1/tasks/cv-r2-16-category-content-seo.md) و [CV-R2-17 / #25](../delivery/review-v1-1/tasks/cv-r2-17-product-content-seo.md) ← [سناریوهای مستند](../bdd/features/experiences/catalog-content-seo.feature) ← مصرف سایت #11 و پذیرش #12. تسک‌ها todo و اسناد in-review هستند؛ سناریوها اجرا نشده‌اند. این تغییر مستندات است و شاهد تکمیل محصول ندارد؛ جزئیات باز R2-Q07 تا R2-Q09 در دفتر پرسش‌هاست.


## مسیر انتشار اولیه و جریان وجه

درخواست مالک → [ADR-007](../decisions/adr-007-initial-launch-sequence.md) → [برنامه و چهار Issue #27 تا #30](../delivery/initial-launch/README.md) → [۲۴ سناریوی مستند](../bdd/features/experiences/initial-launch.feature) → گیت اولیه #12. هر Issue بک‌آفیس/API/سایت دارد؛ خانه با دستور مالک done و سایر مراحل todo هستند. [شاهد اجرای خانه](../delivery/initial-launch/issue-27-verification.md) و [ADR-008 ورود مشترک](../decisions/adr-008-shared-login-modes.md) به همین مسیر اضافه شده‌اند. LCH-Q01 با پاسخ مالک حل شد: ورودی وجه ابتدا شارژ، تأیید کارت‌به‌کارت دستی/بصری و سپس تأیید سفارش با مصرف جدا. LCH-Q02 تا04 فقط بخش وابسته را متوقف می‌کنند. خانه/PLP/PDP با طرح A و C حرف اول Cherriva پیش می‌روند.


## اجرای خانه — CV-LCH-01

[PR #32](https://github.com/MohamadParsa/marketplace/pull/32) → [قرارداد API-HOMEPAGE-LAUNCH-001](../architecture/contracts/homepage-launch.md) و [ADR-008](../decisions/adr-008-shared-login-modes.md) → [هفت معیار پذیرش](../delivery/initial-launch/tasks/cv-lch-01-homepage.md) → [شواهد Go/PostgreSQL/مرورگر و تصاویر](../delivery/initial-launch/issue-27-verification.md). اجرا done و #27 با دستور مالک بسته شد؛ در زمان گزارش خانه #28/#29/#30 اجرا نشده بودند. وضعیت جاری #28 در بخش پایان آن آمده است. ادغام feature با پیش‌نیاز هویت، انتشار لیارا یا پذیرش production نیست.


## رفع هم‌پوشانی صف با مراجع کاتالوگ — ۲۰۲۶-۱۱-۲۱

[تصمیم تکمیلی ADR-007](../decisions/adr-007-initial-launch-sequence.md) → [صف و مالکیت یکتای اجرا](../delivery/initial-launch/README.md#execution-ownership) → #24 / CV-R2-16-AC01..04 فقط از شواهد #28 و #25 / CV-R2-17-AC01..05 فقط از شواهد #29 استفاده می‌کنند؛ پایه موجود #27 بازنویسی نمی‌شود. شناسه‌های BDD ثابت‌اند و تست/PR تکراری برای مراجع ساخته نمی‌شود. معیار شرطی تعامل همچنان باز است؛ نوع reference مجوز حذف معیار نیست. در زمان آن اصلاح برنامه، فقط #27 done بود و کار اجرایی تازه انجام نشده بود؛ وضعیت پس از اجرای #28 در بخش بعد آمده است.

## شاهد پایان #28 و پوشش مشترک #24 — ۲۰۲۶-۱۱-۲۱

[گزارش PLP/جست‌وجو](../delivery/initial-launch/issue-28-verification.md) → CV-LCH-02-AC01..05 و CV-R2-16-AC01..04 → [قرارداد واقعی](../architecture/contracts/plp-search.md)، آزمون HTTP/PostgreSQL و مرورگر. #27/#28 done؛ #11 پیشرفت۲از۴ و #29 بعدی است. چهار معیار مرجع #24 از همان PR شاهد دارند؛ #25 و معیارهای شرطی تعامل بازند. Breadcrumb در PLP اجرا و در #29 به ترتیب خانه، دسته‌ها، برند و نام محصول الزامی شد. merge/deploy یا پذیرش تولید از این وضعیت نتیجه نمی‌شود.

## شاهد پایان #29 — ۲۰۲۶-۱۱-۲۲

[گزارش PDP](../delivery/initial-launch/issue-29-verification.md) → CV-LCH-03-AC01..05 و زیرمجموعه مشخص CV-R2-17 → [قرارداد](../architecture/contracts/product-detail.md)، آزمون HTTP/PostgreSQL و مرورگر. #11 پیشرفت۳از۴؛ #30 طبق دستور مالک ادامه دارد. #25 فقط AC01/02/05 کامل و AC03/04 بخشی شاهد دارند؛ مرجع باز می‌ماند. اطلاعات حساب محلی به درخواست صریح مالک در راهنمای دسترسی تیم ثبت شده است؛ نمونه عمومی یا حساب production فرض نشود.

## برش خرید و کیف پول انتشار اولیه — CV-LCH-04

[شرح #30](../delivery/initial-launch/tasks/cv-lch-04-checkout-wallet-card-transfer.md)، [قرارداد API-CHECKOUT-WALLET-LAUNCH-001](../architecture/contracts/checkout-wallet-launch.md)، [BDD با همان ACها](../bdd/features/experiences/initial-launch.feature)، [شاهد اجرا](../delivery/initial-launch/issue-30-verification.md) و [گیت عمومی](../delivery/initial-launch/production-handoff.md) به هم متصل‌اند. LCH-Q02/03/04 حل‌شده‌اند؛ وضعیت مستقل سفارش/پرداخت و تاریخچه در هر دو سطح ثبت شد. AC07 محیط عمومی همچنان unverified است؛ #12 و #11 بازند و اجرای محلی به معنی پذیرش کل محصول نیست.

## بازبینی #24/#25 پس از خرید اولیه

#24 با همان شواهد #28 بسته و بازبینی شد. [گردش مستقل فروشنده](../architecture/contracts/catalog-content-review.md) و [گزارش تکمیلی](../delivery/review-v1-1/issue-25-verification.md) بخش باقی‌مانده CV-R2-17-AC03 را پوشش می‌دهند؛ قیمت/موجودی/رزرو مستقل و نسخه عمومی قبلی محفوظ‌اند. مالک در۲۰۲۶-۱۱-۲۲ پیشنهادها و تفکیک دامنه را پذیرفت؛ AC04 با آزمون HTTP استقلال تنظیمات/پیش‌فرض و تست کنترل مشترک تکمیل است. #25 در دامنه جاری done؛ [پذیرش API کامل](../delivery/product-interactions-followup.md) در CV-WEB-15/16 معوق و CHQ-10 برای آن مرحله باز است.

## اصلاح تجربه خرید و دسته‌های موجود — ۲۰۲۶-۱۱-۲۲

درخواست مالک → [قرارداد اصلاح تجربه](../architecture/contracts/purchase-experience.md) → [گزارش و شواهد](../delivery/initial-launch/purchase-refinements-verification.md) → تست‌های شماره سفارش/مالکیت/هم‌زمانی/پایداری PostgreSQL، ویزارد و روش ارسال متناسب شهر، سه محصول نمونه تنوع و میان‌بر دسته خانه. همان Product/Variant/Offer و کیف پول مصرف شده‌اند؛ مدل مالی یا مسیر موازی ساخته نشده و گیت عمومی هنوز باز است.

## ورود کامل کاتالوگ در لوکال — ۲۰۲۶-۱۱-۲۳

[گزارش و شاهد ماشین‌خوان](../delivery/catalog-seed/local-import-2026-11-23.md) اتصال داده به خانه/PLP/PDP/سبد، مرز تومان فایل و ریال داخلی، گزینه‌های تنوع متعلق به محصول، منع محلی برای شاخه‌های تازه طلا و حفظ داده قبلی را ثبت می‌کند. اجرا فقط در دیتابیس بررسی محلی است و گیت عمومی #12 یا وضعیت تأیید مشخصات را تغییر نمی‌دهد.

## URL عمومی کاتالوگ — #42

[نیازمندی و AC01..07](../delivery/storefront-followups/public-catalog-urls.md) ← [نگاشت و گزارش اجرا](../delivery/storefront-followups/issue-42-verification.md) ← تست‌های `catalog_urls_test.go` و [مرورگر](../../tests/e2e/catalog-public-urls.cjs). وضعیت سند in-review و اجرای قابلیت done است؛ #41 پیش از شروع این کار بسته شد و #42 پس از عبور CI و ادغام PR #44 با کامنت پایان بسته شد.

## صف تجربه storefront — CV-SF-UX-01 تا10

[۱۰شرح مستقل و مرز مسئولیت](../delivery/storefront-followups/ux-followups.md) ← [فهرست ماشین‌خوان](../delivery/storefront-followups/ux-task-index.json) ← [۱۰سناریوی مستند](../bdd/features/experiences/storefront-ux-followups.feature). هر تسک۴AC دارد. وضعیت اجرا در زمان ثبت todo و منتظر انتخاب مالک بود؛ Issueهای #45تا#54 هنگام ثبت باز و اساین به eliassarband بودند؛ وضعیت جاری در فهرست ماشین‌خوان است. سناریوها اجرا یا passed اعلام نشده‌اند. جزئیات سنجه به SFUX-Q01 وصل‌اند؛ قواعد R20 کاتالوگ و تحویل‌های #41/#42 تکرار یا باز نمی‌شوند.

## اجرای UX-01 — #45

SFUX-Q01 با پاسخ مالک حل شد؛ [شرح](../delivery/storefront-followups/tasks/cv-sf-ux-01-best-selling-sort.md) ← `catalog_sales_test.go` در application/bootstrap و [گزارش/تصاویر](../delivery/storefront-followups/issue-45-verification.md). هر تغییر وضعیت پیاده‌سازی جدا از تأیید سند ثبت می‌شود.

## اجرای UX-02/03 — #46 و #47

[منوی درختی و شواهد #46](../delivery/storefront-followups/issue-46-verification.md) به تست API ناوبری و `category-menu.cjs` وصل است؛ [آیکن‌ها و شواهد #47](../delivery/storefront-followups/issue-47-verification.md) به `header-actions.cjs` و تصاویر قبل/بعد وصل است. #45/#46/#47 بسته‌اند و نتیجه CI در گزارش همان تسک ثبت است.

## اجرای UX-04 — #48

[دامنه قیمت و فیلتر جمع‌شونده](../delivery/storefront-followups/issue-48-verification.md) به `listing_price_range_test.go` و `plp-filters.cjs` وصل است. Switchها در #49 هستند و منطق انتخاب چندگانه این کار تغییر نکرده است.

## اجرای UX-05 — #49

[Switch مشترک و فهرست پوشش](../delivery/storefront-followups/issue-49-verification.md) به `storefront-switches.cjs` (۲۰کنترل) و رگرسیون `plp-filters.cjs` (۲۳کنترل) وصل است؛ تأیید خرید فقط fixture مرورگر و بدون write مالی تست شد.

تحویل انتخاب #45تا#49: هر پنج Issue با کامنت پایان و شواهد بسته، سورس/مدارک روی main و هر۵ job CI در PR متناظر موفق است. شاخص JSON صف وضعیت done این پنج و todo کارهای #50تا#54 را جدا نگه می‌دارد؛ تأیید سند همچنان از اجرای قابلیت مستقل است.

## اجرای UX-06تا08 — #50تا#52

[چیپ مرتب‌سازی](../delivery/storefront-followups/issue-50-verification.md) ← `sort-chips.cjs` و۸کنترل؛ [بارگذاری مشترک](../delivery/storefront-followups/issue-51-verification.md) ← `stable-progress.cjs` و `public-progress.cjs` با۳۱کنترل؛ [محتوای PDP](../delivery/storefront-followups/issue-52-verification.md) ← `pdp-read-more.cjs` با۱۹کنترل. تست‌های مالی این تغییر فقط fixture مرورگرند؛ شواهد اجرای UI از تأیید پرداخت واقعی جداست. سورس هر کار در نقشه کامیت ثبت شد؛ تأیید سند مستقل می‌ماند.

تحویل #50تا#52 در PR #60 با هر۵ job موفق ادغام شد؛۵۸کنترل مرورگر موفق، هر سه با کامنت بسته و سورس/مدارک روی main و D:\Cherriva همگام‌اند. #53/#54 خارج انتخاب جاری‌اند.

## اجرای UX-09/10 — #53/#54

[قالب متن ارسال #53](../delivery/storefront-followups/issue-53-verification.md) ← `dispatch-wording.cjs`،۱۱کنترل؛ [چیپ‌های تنوع #54](../delivery/storefront-followups/issue-54-verification.md) ← `dense-variant-chips.cjs` با۲۳ و `real-gold-variants.cjs` با۵کنترل خواندنی واقعی. رگرسیون۲۳کنترلی مسیر خرید نیز موفق است؛ هیچ تغییر مدل مالی/قیمت طلا/تقویم یا write سفارش در این تحویل نیست.

#53/#54 در PR #61 پس از هر۵ job موفق CI ادغام و با کامنت بسته شدند. هر۱۰تسک UX #45تا#54 done و مدارک/سورس روی main و D:\Cherriva همگام‌اند؛ تأیید سند همچنان از done مستقل است.

## برنامه درخواست‌های بعدی — فقط ثبت در ۲۰۲۶-۱۱-۲۳

[برنامه۳۰شرح](../delivery/next-phase/README.md) ← [فهرست ماشین‌خوان](../delivery/next-phase/task-index.json) ← [۳۰سناریوی مستند، اجرا نشده](../bdd/features/experiences/next-phase-planning.feature). هر شرح۴AC، دامنه و وابستگی دارد؛ ADR-009 و NEXT-Q01تا19 مرز تصمیم را ثبت می‌کنند. #8/#9 مالکان اجرایی ریل دستی/sticky حفظ می‌شوند، و نظر/Q&A ادامه CV-WEB-15/16 هستند. هیچ پیاده‌سازی یا تست passed با ایجاد این بسته ادعا نمی‌شود.

## انتخاب‌های #80/#83/#84/#85/#86

[فهرست جاری](../delivery/next-phase/README.md) به پنج گزارش و تست‌های checkbox/chip، مزایا/معایب، hover/scroll و priority متصل است. معیار دسته و داده شخصی با API/PG و مرورگر جدا بررسی شد؛ گزارش ثبت برنامه تاریخچه است، نه وضعیت اجرای کنونی. پنج کار در این انتخاب پیشین اجرا شدند.

تحویل این پنج انتخاب در [PR #92](https://github.com/MohamadParsa/marketplace/pull/92) پس از هر۵ job موفق CI ادغام شد؛ پنج Issue با کامنت شواهد closed و implementation_status آن‌ها done است. سایر۲۵کار این برنامه todo/open باقی‌اند؛ اسناد همچنان in-review و انتشار لیارا جداست.


## تحلیل محلی #66

[شرح](../delivery/next-phase/tasks/cv-next-05-checkout-amounts-usage-contract.md) ← [قرارداد مبلغ/usage](../architecture/contracts/checkout-amounts-usage.md) ← [گزارش و ACها](../delivery/next-phase/issue-66-analysis.md) ← [مثال‌ها و بررسی محاسبات](../delivery/next-phase/issue-66-checks.json) و [سناریوی پذیرش مستند](../bdd/features/experiences/checkout-amounts-usage.feature). پاسخ‌های مالک از مدل پیشنهادی جدا هستند؛ تحلیل با دستور ادامه مالک done محلی است؛ سند in-review و Issue گیت open مانده‌اند و کد مالی/دیتابیس/گیت تغییر نکرده‌اند. شش انتخاب دیگر در این نوبت شروع نشدند.

## اجرای محلی #63

[شرح](../delivery/next-phase/tasks/cv-next-02-customer-wishlist.md) ← [قرارداد](../architecture/contracts/customer-wishlist.md) ← [سناریوهای مشخص](../bdd/features/experiences/customer-wishlist.feature) ← [گزارش و تصاویر](../delivery/next-phase/issue-63-verification.md) ← [نتیجه۲۶کنترل مرورگر و۲۲۰تست/زیرتست Go](../delivery/next-phase/issue-63-checks.json). علاقه‌مندی خصوصی کاربر، آیکن کنار سبد و قلب PDP done محلی‌اند؛ هیچ push/کامنت/بستن GitHub انجام نشده و #64 شروع نشده است.

## اجرای محلی #64 و اصلاحات صریح UI

[شرح](../delivery/next-phase/tasks/cv-next-03-customer-recently-viewed.md) ← [قرارداد](../architecture/contracts/customer-recently-viewed.md) ← [BDD مستند](../bdd/features/experiences/customer-recently-viewed.feature) ← [گزارش/تصاویر](../delivery/next-phase/issue-64-verification.md) ← [نتیجه۶۵کنترل مرورگر و۲۲۷تست/زیرتست Go](../delivery/next-phase/issue-64-checks.json). تغییر قلب PDP/هدر و breadcrumb هم دستور صریح مالک همین نوبت‌اند؛ گزارش #63 پیوند به شواهد جدید دارد. قابلیت done محلی، سند in-review و Issue گیت open است؛ هیچ commit/push/کامنت/بستن انجام نشد.

## انتشار دسته مشتری در۲۰۲۶-۱۱-۲۴

تغییرات در [PR #93](https://github.com/MohamadParsa/marketplace/pull/93) پس از موفقیت هر۵ job CI، شامل Go race/vet، روی main با کامیت `a253d5edbe0adb467e1e7fb2ab5903608fb64bdc` ادغام شدند. این Issue با کامنت شواهد بسته شد. استقرار لیارا انجام نشد؛ نسخه D:\Cherriva نیز با main همگام است. تغییرهای بعد از این دسته تا دستور صریح تازه مالک لوکال می‌مانند. #63/#64 قابلیت اجرایی و #66 خروجی تحلیل هستند؛ هیچ کار انتخاب‌نشده‌ای done نشده است.

## اجرای محلی #67/#68/#69/#78 — ۲۰۲۶-۱۱-۲۴

FEAT-NEXT-006/007/008/017 با [قرارداد اجرایی](../architecture/contracts/checkout-adjustments.md)، [گزارش AC و تست‌ها](../delivery/next-phase/issues-67-68-69-78-verification.md) و [صف ماشین‌خوان](../delivery/next-phase/task-index.json) مرتبط‌اند. وضعیت پیاده‌سازی done و Issueهای #67/#68/#69/#78 به دستور تازه مالک completed/closed هستند؛ سورس در PR #94 پوش شده، ادغام/استقرار انجام نشده و اسناد in-review باقی‌اند. [کامنت‌های تحویل](../delivery/backlog-review-2026-11-24/completed-closures.json).

## پرسش و پاسخ محصول

[FEAT-NEXT-029 / CV-NEXT-29](../delivery/next-phase/tasks/cv-next-29-product-qa-service.md) ادامهٔ CV-WEB-16 است. [تصمیم‌های CHQ-10](../product/open-questions.md) به دسترسی ثبت و ادارهٔ انتشار، [سناریوهای پذیرش](../bdd/features/experiences/next-phase-planning.feature) به رفتار و [شواهد تسک ۹۰](../delivery/next-phase/issue-90-verification.md) به آزمون‌های API، داده پایدار و مرورگر پیوند دارند.

[FEAT-NEXT-028 — نظرات محصول](../delivery/next-phase/tasks/cv-next-28-product-reviews-service.md) ← [گیت مشترک کاتالوگ](../architecture/contracts/catalog-content-review.md#اجرای-کنترل-مشترک) ← [تصمیم CHQ-10](../product/open-questions.md) ← [پذیرش BDD](../bdd/features/experiences/next-phase-planning.feature) و [گردش مرورگر](../../tests/e2e/product-reviews.py).

## تطبیق کامل صف و درخواست‌های تازه۲۰۲۶-۱۱-۲۴

[PLAN-BACKLOG-RECONCILIATION-001](../delivery/backlog-review-2026-11-24/README.md) و [ماتریس کامل](../delivery/backlog-review-2026-11-24/audit-matrix.json)، مالک هر نیاز قدیمی/جدید و۱۱PR superseded را ثبت می‌کنند. FEAT-REV-001تا010 در #96تا#105 و شبکه اجتماعی در #87 پیگیری می‌شوند؛ این مرحله برنامه‌ریزی است، نه مجوز کدنویسی. دریافت/ادغام مجاز سه شاخه پارسا و مرز لوکال/انتشار در گزارش جدا ثبت است.

## تحویل منتخب #103/#104/#102

FEAT-REV-008/009/007 ← [گزارش AC، تصاویر و نتایج](../delivery/backlog-review-2026-11-24/issues-102-103-104-verification.md) ← [قرارداد توکن عمومی](../architecture/contracts/public-catalog-references.md). اجرای این سه قابلیت به دستور مالک محلی تکمیل شد؛ سایر تسک‌های تازه شروع نشده و انتشار این commit مجاز نشده است. وضعیت سند in-review باقی است.

## تکمیل #96 تا #99 — ۲۰۲۶-۱۱-۲۵

[چهار شرح کار](../delivery/backlog-review-2026-11-24/README.md) ← [قرارداد](../architecture/contracts/wallet-credit-policy.md) ← [۱۶ سناریوی BDD](../bdd/features/experiences/checkout-wallet-policy.feature) ← [گزارش و تصاویر](../delivery/backlog-review-2026-11-24/issues-96-99-verification.md). وضعیت اجرای این چهار برش done است؛ ادغام و انتشار از بستن Issue نتیجه نمی‌شود.


## اجرای محلی آدرس و نقشه #100/#101

FEAT-REV-005 و FEAT-REV-006 ← [قرارداد](../architecture/contracts/customer-address-map.md) ← [شواهد/تصاویر](../delivery/backlog-review-2026-11-24/issues-100-101-verification.md). پاسخ مالک در ۲۰۲۶-۱۱-۲۵ نام/کد ملی پروفایل برای self و نام/کد ملی/موبایل گیرنده دیگر را الزامی کرد. نشانی کامل بالاترین ردیف تمام‌عرض شد؛ اطلاعات تاریخی ثابت می‌مانند. نقشه طبق انتخاب مالک اتصال آماده دارد؛ کلید/دامنه/آزمون واقعی و انتشار Git/لیارا جدا هستند.


## اصلاحات پنل و کاتالوگ — ۲۰۲۶-۱۱-۲۵

درخواست مالک و چهار پاسخ تکمیلی ← [پوشش بندبه‌بند](../delivery/backoffice-refinement-2026-11-25/request-coverage.md) ← FEAT-BO-REFINEMENT-001 تا011 در #107 تا#117؛ [برنامه/وابستگی و سؤال‌ها](../delivery/backoffice-refinement-2026-11-25/README.md). #73 مالک روش فروش/مارکاپ و #81 مالک الگوی مشترک گرید/Excel باقی‌اند؛ شرح/مرز آن‌ها اصلاح شد و #15/#71 هماهنگی می‌کنند. اجرا todo و مجوز کد/پوش صادر نشده است.

## اجرای محلی #73 و #81 —۲۰۲۶-۱۱-۲۵

[مشخصات #73](../delivery/next-phase/tasks/cv-next-12-sales-method-markup.md) و [#81](../delivery/next-phase/tasks/cv-next-20-backoffice-grid-capabilities.md) ← [قرارداد روش فروش](../architecture/contracts/sales-methods.md) و [گرید](../architecture/contracts/backoffice-grids.md) ← [BDD](../bdd/features/experiences/sales-methods-and-grid-export.feature) ← [گزارش و شواهد](../delivery/next-phase/issues-73-81-verification.md). [ممیزی۲۷گرید و مالکان شکاف‌ها](../delivery/next-phase/backoffice-grid-audit.md). تحویل محلی با پوش/استقرار و تکمیل فرم‌های #108تا#117 یکسان نیست.

## اجرای محلی #111 و #112

[شرح کاربر](../delivery/backoffice-refinement-2026-11-25/tasks/unified-user-editing-and-filters.md) و [نقش](../delivery/backoffice-refinement-2026-11-25/tasks/dedicated-role-management-and-user-grants.md) ← [قرارداد](../architecture/contracts/user-management.md) ← [BDD](../bdd/features/experiences/user-management-and-roles.feature) ← [آزمون و تصاویر](../delivery/backoffice-refinement-2026-11-25/issues-111-112-verification.md). مجوز اجرا محدود همین دو و تغییرها محلی است؛ ممیزی عمومی scope در #71 باقی می‌ماند.


## مرز خصوصیات پنل — #110 و اصلاح آدرس

[تسک FEAT-BO-REFINEMENT-004](../delivery/backoffice-refinement-2026-11-25/tasks/cherriva-attribute-codes-and-labels.md) → [قرارداد خصوصیات](../architecture/contracts/catalog-attributes.md) → [گزارش API/مرورگر/حفظ داده و چیدمان آدرس](../delivery/backoffice-refinement-2026-11-25/issue-110-verification.md). گرید/دیالوگ/انتساب عنوان خوانا دارند و alias سازگار با ورودی قدیمی است؛ بازنویسی داده واردشده و تبدیل محور در این کار نیست.


## کاتالوگ مستقل و خصوصیات دسته — #107/#108

FEAT-BO-REFINEMENT-001/002 → [گزارش ناوبری و دسته](../delivery/backoffice-refinement-2026-11-25/issues-107-108-verification.md) → [قرارداد خصوصیات و سطح نشست](../architecture/contracts/catalog-attributes.md). تست API، cache/عدم ارث، صفحه‌بندی/فیلتر و مرورگر به همان گزارش متصل‌اند؛ تغییر محور مصرف‌شده #109 و تکمیل فیلدهای برند/محصول #115/#116 جداست.

FEAT-BO-REFINEMENT-003 → [گزارش #109](../delivery/backoffice-refinement-2026-11-25/issue-109-verification.md) → [قرارداد تبدیل](../architecture/contracts/category-axis-conversion.md) → [BDD](../bdd/features/experiences/category-axis-conversion.feature).
