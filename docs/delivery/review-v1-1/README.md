# تسک‌های فعال نسخه۱.۱ پس از پاسخ #13

تاریخ:۲۰۲۶-۱۱-۲۰ | وضعیت مشخصات: in-review | مالک بررسی: محصول و راهبر فنی؛ فرد/موعد انتساب نشده | [فرایند #14](https://github.com/MohamadParsa/marketplace/issues/14)

این پوشه بازنویسی اجرایی تسک‌های #1 تا #12 با پاسخ‌های [ADR-005](../../decisions/adr-005-review-v1-1-product-responses.md) است. اصل snapshot۱.۱ و manifest آن محفوظ‌اند. ۱۵ شرح کار داریم: ۱۲ کار فعال (۱۰todo و۲review) و۳ کار معوق؛ هیچ پیاده‌سازی با تدوین این سند done نشده است. سه کار جدید تنظیمات عمومی، ظرفیت اسلات و خبرم‌کن از پاسخ مالک تفکیک شده‌اند.

## ترتیب و وابستگی

| تسک / Issue | خروجی | اولویت / موج پیشنهادی | وضعیت اجرا | وابستگی |
| --- | --- | --- | --- | --- |
| [CV-R2-01](tasks/cv-r2-01-single-user.md) · [#1](https://github.com/MohamadParsa/marketplace/issues/1) | هویت واحد، نقش admin و عضویت چندبه‌چند | P0 / W0 | todo | CV-FND-04، CV-IAM-03 |
| [CV-R2-02](tasks/cv-r2-02-mobile-first-login.md) · [#2](https://github.com/MohamadParsa/marketplace/issues/2) | ورود موبایل، OTP و لینک رمز پس از شماره | P0 / W0 | todo | CV-R2-01، CV-IAM-01، CV-IAM-02 |
| [CV-R2-03](tasks/cv-r2-03-profile-password-recovery.md) · [#3](https://github.com/MohamadParsa/marketplace/issues/3) | پروفایل، رمز و بازیابی OTP با شروع ورود از موبایل | P0 / W0 | todo | CV-R2-02 |
| [CV-R2-04](tasks/cv-r2-04-category-brand-media.md) · [#4](https://github.com/MohamadParsa/marketplace/issues/4) | آیکن دسته و برند با پیش‌فرض | P1 / W1 | review | CV-CAT-01، CV-CAT-03، CV-CAT-07 |
| [CV-R2-05](tasks/cv-r2-05-shared-attributes.md) · [#5](https://github.com/MohamadParsa/marketplace/issues/5) | کتابخانه خصوصیت، انتساب صریح و سقف کارت | P0 / W1 | review | CV-CAT-01، CV-CAT-02، CV-R2-13 |
| [CV-R2-06](tasks/cv-r2-06-product-workspace.md) · [#6](https://github.com/MohamadParsa/marketplace/issues/6) | اکشن سریع تنوع، قیمت و موجودی در محصول | P0 / W1 | todo | CV-R2-05، CV-CAT-04، CV-CAT-06 |
| [CV-R2-07](tasks/cv-r2-07-responsive-banner.md) · [#7](https://github.com/MohamadParsa/marketplace/issues/7) | سکشن بنر با تصویر دسکتاپ و موبایل | P1 / later | deferred | CV-CMS-01، CV-CMS-02، CV-R2-04 |
| [CV-R2-08](tasks/cv-r2-08-manual-product-rail.md) · [#8](https://github.com/MohamadParsa/marketplace/issues/8) | کروسل اختصاصی با انتخاب و ترتیب دستی کالا | P1 / later | deferred | CV-CMS-01، CV-CMS-02، CV-CAT-09، CV-PRC-07 |
| [CV-R2-09](tasks/cv-r2-09-sticky-banner.md) · [#9](https://github.com/MohamadParsa/marketplace/issues/9) | سکشن استیکی با کنترل نمایش و بستن | P1 / later | deferred | CV-R2-07 |
| [CV-R2-10](tasks/cv-r2-10-seller-lead-snapshot.md) · [#10](https://github.com/MohamadParsa/marketplace/issues/10) | وقفه تأمین فروشنده، تقویم و حفظ وعده سفارش | P0 / W2 | todo | CV-IAM-06، CV-CAT-06، CV-OPS-01، CV-R2-13 |
| [CV-R2-11](tasks/cv-r2-11-storefront-alignment.md) · [#11](https://github.com/MohamadParsa/marketplace/issues/11) | همگام‌سازی سایت اصلی با تصمیم‌های نسخه۱.۱ | P0 / W3 | todo | CV-R2-03، CV-R2-04، CV-R2-05، CV-R2-06، CV-R2-10، CV-R2-13، CV-R2-14، CV-R2-15 |
| [CV-R2-12](tasks/cv-r2-12-revision-acceptance.md) · [#12](https://github.com/MohamadParsa/marketplace/issues/12) | پذیرش اصلاحات و ردیابی با دامنه فعال | P0 / W3 | todo | CV-R2-01، CV-R2-02، CV-R2-03، CV-R2-04، CV-R2-05، CV-R2-06، CV-R2-10، CV-R2-11، CV-R2-13، CV-R2-14، CV-R2-15 |
| [CV-R2-13](tasks/cv-r2-13-general-settings.md) · [#15](https://github.com/MohamadParsa/marketplace/issues/15) | تنظیمات عمومی دسته‌بندی‌شده سایت | P0 / W1 | todo | CV-R2-01، CV-FND-03، CV-FND-04 |
| [CV-R2-14](tasks/cv-r2-14-shipment-slots.md) · [#16](https://github.com/MohamadParsa/marketplace/issues/16) | اسلات ارسال ظرفیت‌دار و کنترل رزرو | P0 / W2 | todo | CV-R2-10، CV-R2-13، CV-OPS-05، CV-OPS-06، CV-FIN-03 |
| [CV-R2-15](tasks/cv-r2-15-stock-alert.md) · [#17](https://github.com/MohamadParsa/marketplace/issues/17) | درخواست اطلاع موجودشدن همان تنوع | P1 / W2 | todo | CV-CAT-05، CV-CAT-06، CV-FND-06 |

W0: هویت/مجوز/ورود و پروفایل. W1: تنظیمات عمومی، کاتالوگ و فضای محصول. W2: وقفه/وعده، ظرفیت اسلات و سرویس خبرم‌کن با آداپتر آزمون؛ در کنار تمام عملیات/مالی پنل پایه. W3: سایت اصلی با طرح A و پذیرش یکپارچه، پس از قبولی کل دامنه پنل/سلر با آزمون. later: CMS معوق پس از فعال‌سازی صریح مالک. وابستگی داخل موج هم رعایت شود؛ P0 به معنی دورزدن dependency نیست.

#7/#8/#9 در شرط شروع #11/#12 نیستند؛ خانه چیدمان ثابت و داده خرید معتبر دارد. زیرسایت‌ها و نشست میان‌میزبانی در شروع نیستند. سه کار جدید، خلأ دامنه را روشن می‌کنند؛ اولویت/اندازه نهایی و مسئول در kickoff مشخص شوند و کار بزرگ با حفظ شناسه مادر خرد شود.

## پاسخ به Issue و سند متأثر

| پاسخ | کارها و Issueهای متأثر | مرجع جاری |
| --- | --- | --- |
| [R2-Q01](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-01 / #1](tasks/cv-r2-01-single-user.md)، [CV-R2-02 / #2](tasks/cv-r2-02-mobile-first-login.md)، [CV-R2-03 / #3](tasks/cv-r2-03-profile-password-recovery.md)، [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md) | [سند حوزه/تجربه](../../domains/identity/README.md) |
| [R2-Q02](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-01 / #1](tasks/cv-r2-01-single-user.md)، [CV-R2-02 / #2](tasks/cv-r2-02-mobile-first-login.md)، [CV-R2-06 / #6](tasks/cv-r2-06-product-workspace.md)، [CV-R2-12 / #12](tasks/cv-r2-12-revision-acceptance.md) | [سند حوزه/تجربه](../../product/roles.md) |
| [R2-Q03](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-05 / #5](tasks/cv-r2-05-shared-attributes.md)، [CV-R2-06 / #6](tasks/cv-r2-06-product-workspace.md)، [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md)، [CV-R2-13 / #15](tasks/cv-r2-13-general-settings.md) | [سند حوزه/تجربه](../../domains/catalog/README.md) |
| [R2-Q04](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-04 / #4](tasks/cv-r2-04-category-brand-media.md)، [CV-R2-07 / #7](tasks/cv-r2-07-responsive-banner.md)، [CV-R2-08 / #8](tasks/cv-r2-08-manual-product-rail.md)، [CV-R2-09 / #9](tasks/cv-r2-09-sticky-banner.md)، [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md)، [CV-R2-12 / #12](tasks/cv-r2-12-revision-acceptance.md) | [سند حوزه/تجربه](../../experiences/backoffice/management.md) |
| [R2-Q05](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-10 / #10](tasks/cv-r2-10-seller-lead-snapshot.md)، [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md)، [CV-R2-13 / #15](tasks/cv-r2-13-general-settings.md)، [CV-R2-14 / #16](tasks/cv-r2-14-shipment-slots.md) | [سند حوزه/تجربه](../../domains/shipping/README.md) |
| [R2-Q06](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) | [CV-R2-02 / #2](tasks/cv-r2-02-mobile-first-login.md)، [CV-R2-03 / #3](tasks/cv-r2-03-profile-password-recovery.md)، [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md) | [سند حوزه/تجربه](../../product/scope.md) |
| [Q-027](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-027) | [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md)، [CV-R2-12 / #12](tasks/cv-r2-12-revision-acceptance.md)، [CV-R2-15 / #17](tasks/cv-r2-15-stock-alert.md) | [سند حوزه/تجربه](../../experiences/storefront/discovery.md) |
| [Q-028](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-028) | [CV-R2-11 / #11](tasks/cv-r2-11-storefront-alignment.md)، [CV-R2-12 / #12](tasks/cv-r2-12-revision-acceptance.md) | [سند حوزه/تجربه](../../experiences/backoffice/panel-design.md) |
| تنظیمات عمومی | [CV-R2-13 / #15](tasks/cv-r2-13-general-settings.md) | [مدیریت](../../experiences/backoffice/management.md) |

تنظیمات عمومی نیز به [مدیریت](../../experiences/backoffice/management.md#تنظیمات-عمومی-دستهبندیشده)، ظرفیت به [ارسال](../../domains/shipping/README.md) و خبرم‌کن به [اطلاع‌رسانی](../../domains/notifications/README.md) پیوند دارند. هر فایل کار، پاسخ/دامنه، شرط شروع بخش مسدود، AC و شاهد بستن را دارد.

## قرارداد بستن کار

Issue پیاده‌سازی فقط با نتیجه و شواهد AC همان کار بسته شود؛ کار معوق انجام‌شده محسوب نشود. پیشنهاد تغییر موبایل، سقف اولیه کارت، تغییر آینده محور، سیاست ظرفیت و تأیید/ارسال اعلان تا پاسخ مالک/قرارداد منتخب در همان بخش blocked هستند؛ کار مستقل روشن ادامه یابد. گیت اتصال واقعی از آداپتر آزمون جداست.

[گزارش همگام‌سازی فرایند #14](verification.md) و [JSON تسک‌ها](task-index.json) برای کنترل نسخه و خواندن agent هستند؛ این گزارش شاهد آزمون نرم‌افزار نیست.

## اجرای Issue#4

آیکن دسته/برند پیاده‌سازی و محلی بررسی شد؛ [گزارش و شواهد](issue-4-verification.md). این کار review و منتظر ادغام است؛ گزارش همگام‌سازی #14 سابقه برنامه‌ریزی همان تاریخ است و وضعیت اجرای تازه را جایگزین نمی‌کند.

## اجرای Issue#5

خصوصیات مشترک، انتساب مستقل و نمایش کارت با سقف عمومی در نسخه آزمون پیاده شد؛ [گزارش و تصاویر](issue-5-verification.md). وضعیت کد review است. فقط برش تنظیم سقف از #15 اجرا شده؛ سایر تنظیمات هنوز todo هستند. Issue#4 به درخواست مالک بسته شد ولی PR #18 هنوز ادغام نشده است؛ بسته‌شدن Issue با ادغام و استقرار یکی نیست.
