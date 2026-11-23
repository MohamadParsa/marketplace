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
