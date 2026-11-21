# تسک‌های فعال نسخه۱.۱ پس از پاسخ #13

## مسیر اجرای مقدم برای انتشار اولیه

به دستور مالک۲۰۲۶-۱۱-۲۱، [برنامه چهارمرحله‌ای](../initial-launch/README.md) و [ADR-007](../../decisions/adr-007-initial-launch-sequence.md) مرجع صف اجرا هستند. #11 والد #27 خانه → #28 PLP → #29 PDP → #30 خرید/کیف پول کارت‌به‌کارت است. هر تسک بک‌آفیس+API+سایت را شامل می‌شود. فهرست۱۷تایی پایین تسک‌های مادر/اصلاحات را حفظ می‌کند؛ وابستگی‌های W3 یا پنل کامل در متن قدیمی نباید گیت این انتشار فرض شوند. طرح A و C حرف اول Cherriva قطعی‌اند؛ ثبت رسید، شارژ کیف پول و تأیید سفارش سه رویداد مستقل‌اند.


تاریخ:۲۰۲۶-۱۱-۲۱ | وضعیت مشخصات: in-review | مالک بررسی: محصول و راهبر فنی؛ فرد/موعد انتساب نشده | [فرایند #14](https://github.com/MohamadParsa/marketplace/issues/14)

این پوشه بازنویسی اجرایی تسک‌های #1 تا #12 با پاسخ‌های [ADR-005](../../decisions/adr-005-review-v1-1-product-responses.md) است. اصل snapshot۱.۱ و manifest آن محفوظ‌اند. ۱۷ شرح مرجع داریم: ۱۵ ردیف غیرمعوق (۹todo،۱done،۱in-progress و۴review) و۲ ردیف معوق. این شمار، تعداد کار صف انتشار نیست؛ #11 والد پیگیری و #24/#25 مرجع نیازمندی‌اند. در صف جاری #27/#28 done هستند؛ #29 بعدی و #30 پس از آن است. سه کار جدید تنظیمات عمومی، ظرفیت اسلات و خبرم‌کن از پاسخ مالک تفکیک شده‌اند.

## ترتیب و وابستگی

صف اجرایی جاری: #27 done → #28 done → **#29 کار بعدی** → #30. #24/#25 بدون رتبه اجرای مستقل‌اند؛ [مرز یکتای هر کار](../initial-launch/README.md#execution-ownership). موج‌های W0 تاW3 جدول پایین سابقه برنامه گسترده‌ترند و برای انتشار اول گیت موازی ایجاد نمی‌کنند.

| تسک / Issue | خروجی | اولویت / موج پیشنهادی | وضعیت اجرا | وابستگی |
| --- | --- | --- | --- | --- |
| [CV-R2-01](tasks/cv-r2-01-single-user.md) · [#1](https://github.com/MohamadParsa/marketplace/issues/1) | هویت واحد، نقش admin و عضویت چندبه‌چند | P0 / W0 | todo | CV-FND-04، CV-IAM-03 |
| [CV-R2-02](tasks/cv-r2-02-mobile-first-login.md) · [#2](https://github.com/MohamadParsa/marketplace/issues/2) | ورود موبایل، OTP و لینک رمز پس از شماره | P0 / W0 | todo | CV-R2-01، CV-IAM-01، CV-IAM-02 |
| [CV-R2-03](tasks/cv-r2-03-profile-password-recovery.md) · [#3](https://github.com/MohamadParsa/marketplace/issues/3) | پروفایل، رمز و بازیابی OTP با شروع ورود از موبایل | P0 / W0 | todo | CV-R2-02 |
| [CV-R2-04](tasks/cv-r2-04-category-brand-media.md) · [#4](https://github.com/MohamadParsa/marketplace/issues/4) | آیکن دسته و برند با پیش‌فرض | P1 / W1 | review | CV-CAT-01، CV-CAT-03، CV-CAT-07 |
| [CV-R2-05](tasks/cv-r2-05-shared-attributes.md) · [#5](https://github.com/MohamadParsa/marketplace/issues/5) | کتابخانه خصوصیت، انتساب صریح و سقف کارت | P0 / W1 | review | CV-CAT-01، CV-CAT-02، CV-R2-13 |
| [CV-R2-06](tasks/cv-r2-06-product-workspace.md) · [#6](https://github.com/MohamadParsa/marketplace/issues/6) | اکشن سریع تنوع، قیمت و موجودی در محصول | P0 / W1 | review | CV-R2-05، CV-CAT-04، CV-CAT-06 |
| [CV-R2-07](tasks/cv-r2-07-responsive-banner.md) · [#7](https://github.com/MohamadParsa/marketplace/issues/7) | سکشن بنر با تصویر دسکتاپ و موبایل | P1 / W1 | review | CV-CMS-01، CV-CMS-02، CV-R2-04 |
| [CV-R2-08](tasks/cv-r2-08-manual-product-rail.md) · [#8](https://github.com/MohamadParsa/marketplace/issues/8) | کروسل اختصاصی با انتخاب و ترتیب دستی کالا | P1 / later | deferred | CV-CMS-01، CV-CMS-02، CV-CAT-09، CV-PRC-07 |
| [CV-R2-09](tasks/cv-r2-09-sticky-banner.md) · [#9](https://github.com/MohamadParsa/marketplace/issues/9) | سکشن استیکی با کنترل نمایش و بستن | P1 / later | deferred | CV-R2-07 |
| [CV-R2-10](tasks/cv-r2-10-seller-lead-snapshot.md) · [#10](https://github.com/MohamadParsa/marketplace/issues/10) | وقفه تأمین فروشنده، تقویم و حفظ وعده سفارش | P0 / W2 | todo | CV-IAM-06، CV-CAT-06، CV-OPS-01، CV-R2-13 |
| [CV-R2-11](tasks/cv-r2-11-storefront-alignment.md) · [#11](https://github.com/MohamadParsa/marketplace/issues/11) | والد انتشار اولیه | P0 / L0 | in-progress | CV-LCH-01 تا04 |
| [CV-R2-12](tasks/cv-r2-12-revision-acceptance.md) · [#12](https://github.com/MohamadParsa/marketplace/issues/12) | پذیرش اصلاحات و ردیابی با دامنه فعال | P0 / W3 | todo | CV-R2-01، CV-R2-02، CV-R2-03، CV-R2-04، CV-R2-05، CV-R2-06، CV-R2-10، CV-R2-11، CV-R2-13، CV-R2-14، CV-R2-15، CV-R2-16، CV-R2-17 |
| [CV-R2-13](tasks/cv-r2-13-general-settings.md) · [#15](https://github.com/MohamadParsa/marketplace/issues/15) | تنظیمات عمومی دسته‌بندی‌شده سایت | P0 / W1 | todo | CV-R2-01، CV-FND-03، CV-FND-04 |
| [CV-R2-14](tasks/cv-r2-14-shipment-slots.md) · [#16](https://github.com/MohamadParsa/marketplace/issues/16) | اسلات ارسال ظرفیت‌دار و کنترل رزرو | P0 / W2 | todo | CV-R2-10، CV-R2-13، CV-OPS-05، CV-OPS-06، CV-FIN-03 |
| [CV-R2-15](tasks/cv-r2-15-stock-alert.md) · [#17](https://github.com/MohamadParsa/marketplace/issues/17) | درخواست اطلاع موجودشدن همان تنوع | P1 / W2 | todo | CV-CAT-05، CV-CAT-06، CV-FND-06 |
| [CV-R2-16](tasks/cv-r2-16-category-content-seo.md) · [#24](https://github.com/MohamadParsa/marketplace/issues/24) | مرجع نیازمندی دسته؛ اجرا فقط در #28 | P0 نیاز / بدون صف مستقل | done / مرجع | هر چهار AC با [شاهد #28](../initial-launch/issue-28-verification.md)، بدون اجرای مستقل |
| [CV-R2-17](tasks/cv-r2-17-product-content-seo.md) · [#25](https://github.com/MohamadParsa/marketplace/issues/25) | مرجع نیازمندی محصول؛ اجرا فقط در #29 | P0 نیاز / بدون صف مستقل | todo / مرجع | شاهد #29؛ معیار شرطی تعامل تا آمادگی قابلیت باز است |

W0: هویت/مجوز/ورود و پروفایل. W1: تنظیمات عمومی، کاتالوگ و فضای محصول. W2: وقفه/وعده، ظرفیت اسلات و سرویس خبرم‌کن با آداپتر آزمون؛ در کنار تمام عملیات/مالی پنل پایه. W3: سایت اصلی با طرح A و پذیرش یکپارچه، پس از قبولی کل دامنه پنل/سلر با آزمون. later: CMS معوق پس از فعال‌سازی صریح مالک. وابستگی داخل موج هم رعایت شود؛ P0 به معنی دورزدن dependency نیست.

#7 به دستور مالک فقط در برش بنر فعال شد و #8/#9 معوق‌اند؛ این سه در شرط اجباری شروع #11/#12 نیستند؛ خانه چیدمان ثابت و داده خرید معتبر دارد. زیرسایت‌ها و نشست میان‌میزبانی در شروع نیستند. کارهای تکمیلی، خلأ دامنه را روشن می‌کنند؛ اولویت/اندازه نهایی و مسئول در kickoff مشخص شوند و کار بزرگ با حفظ شناسه مادر خرد شود.

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

## اجرای Issue#6

فرم مشترک تنوع و عرضه از گرید/جزئیات، stock/reserved و scope فروشگاه پیاده شد؛ [گزارش و تصاویر](issue-6-verification.md). وضعیت کد review است؛ بسته‌شدن Issue به دستور مالک، ادغام یا استقرار را اعلام نمی‌کند. این برش روی شاخه #5 است؛ ترتیب ادغام #18 سپس #19 و PR این کار رعایت شود.

## اجرای Issue#7

[ADR-006](../../decisions/adr-006-responsive-banner-activation.md) فعال‌سازی محدود مالک را ثبت می‌کند. بنر دو اندازه، draft/private preview و انتشار دستی نسخه‌دار در نسخه آزمون اجرا شد؛ [گزارش و تصاویر](issue-7-verification.md). #8/#9 معوق‌اند؛ اتصال خانه، storage واقعی و پذیرش تولید جداست.


## تکمیل مدل محتوا و SEO — ۲۰۲۶-۱۱-۲۱

[تطبیق مدل فعلی با درخواست و فهرست کامل فیلدها](../../domains/catalog/content-and-seo.md) ثبت شد. [#24](https://github.com/MohamadParsa/marketplace/issues/24) برای دسته و [#25](https://github.com/MohamadParsa/marketplace/issues/25) برای محصول، مراجع نیازمندی بدون کار اجرایی مستقل‌اند؛ هر چهار AC #24 در #28 بررسی شده‌اند و #25 برای اجرای #29 و موارد شرطی باز است. #11 فقط پیگیری صف و #12 جمع‌بندی گیت است؛ شواهد تکرار نمی‌شوند. مدل‌های قبلی #4/#5/#6 محدود به آیکن/خصوصیات/تنوع/عرضه بوده‌اند و انجام آن‌ها تکمیل محتوای محصول نیست. R2-Q07 تا R2-Q09 در دفتر پرسش‌ها جزئیات باز را جدا می‌کنند؛ کد محصول در این تغییر دست نخورده است.
