# نقشه تسک‌های تحلیل پایه چریوا

## صف انتشار اولیه جاری

[برنامه انتشار و چهار برش مشترک](initial-launch/README.md) مرجع اجرای بعدی‌اند: #27 خانه، #28 PLP، #29 PDP و #30 خرید/کیف پول کارت‌به‌کارت، زیر والد #11. این‌ها برش اجرایی تسک‌های مادرند و به معنی تکرار نیازمندی‌های کاتالوگ نیستند. #12 گیت محدود را جدا از پذیرش کامل ثبت می‌کند؛ [ADR-007](../decisions/adr-007-initial-launch-sequence.md) تقدم بر شرط قبلی پنل کامل را توضیح می‌دهد.


تحلیل ۱.۰ مورخ ۲۰۲۶-۱۱-۱۷ شامل ۹۱ تسک در ۱۱ گروه است. وضعیت ثبت‌شده تمام تسک‌ها در نسخه بسته `todo` است. وضعیت اجرای فعلی هر قابلیت در رودمپ و گزارش تحویل همان محیط پیگیری می‌شود. شناسه اصلی، اندازه، موج، وابستگی، مانع و پذیرش هر تسک در فهرست مبدأ باقی می‌ماند.

[فهرست انسانی تسک‌ها](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/tasks/README.md) · [فهرست ماشین‌خوان](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/task-index.json) · [ماتریس ۲۹ ردیف پوشش](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/14-coverage.md) · [رودمپ جاری](roadmap.md)

| گروه تسک | تعداد | موضوع و مرجع فعلی |
| --- | ---: | --- |
| CV-FND | ۷ | مرز ماژول، مدل، قرارداد خطا، تغییر نسخه، داده آزمون و پوسته پنل. [معماری](../architecture/README.md) · [قرارداد منطقی](../architecture/contracts/interactions.md) |
| CV-IAM | ۷ | ورود، نشست، مجوز عملیات، محدوده فروشگاه، مشتری و نشانی. [هویت](../domains/identity/README.md) · [نقش‌ها](../product/roles.md) |
| CV-CAT | ۱۰ | دسته، خصوصیت، برند، تنوع، عرضه، انتشار و جست‌وجو. [کاتالوگ](../domains/catalog/README.md) · [کشف](../experiences/storefront/discovery.md) |
| CV-PRC | ۸ | قیمت، نرخ طلا، روش فروش، مارکاپ، کمیسیون و عرضه نماینده. [قیمت‌گذاری](../domains/pricing/README.md) |
| CV-DSC | ۵ | تخفیف، کد، محدودیت مصرف و سهم هر قلم در استرداد. [قیمت‌گذاری](../domains/pricing/README.md) · [پرداخت](../domains/payments/README.md) |
| CV-CMS | ۶ | صفحه، سکشن، تنظیمات ویترین و انتشار. [مدیریت بک‌آفیس](../experiences/backoffice/management.md) |
| CV-OPS | ۹ | سفارش، تأمین، پرونده، واگذاری و مرسوله. [سفارش](../domains/orders/README.md) · [ارسال](../domains/shipping/README.md) |
| CV-FIN | ۸ | کیف پول، رزرو، برداشت، استرداد، تطبیق و تسویه. [پرداخت](../domains/payments/README.md) · [تسویه](../domains/settlements/README.md) |
| CV-WEB | ۱۸ | کشف، حساب، سبد، خرید، پیگیری، SEO و تعامل محتوایی. [تجربه سایت](../experiences/storefront/README.md) · [پرسش‌های محصول](../product/open-questions.md#پرسشهای-تحلیل-پایه) |
| CV-INT | ۷ | اتصال درگاه‌ها، پیامک، عملیات، رسانه و انتشار. [قراردادها](../architecture/contracts/README.md) · [محیط‌ها](environments.md) |
| CV-QA | ۶ | پذیرش دسترسی، قیمت، هم‌زمانی، عملیات، رابط و انتشار. [کیفیت](../architecture/quality.md) · [BDD](../bdd/README.md) |

## موج‌ها و ترتیب تحویل

| موج پیشنهادی بسته | تعداد | خروجی قابل مرور |
| --- | ---: | --- |
| W0 | ۱۰ | قراردادها و زیربنای مشترک |
| W1 | ۳۰ | کاتالوگ و تنظیمات پنل |
| W2 | ۲۱ | عملیات سفارش و مالی پنل با داده آزمون |
| W3 | ۲۲ | سایت مشتری و اتصال داخلی |
| W4 | ۸ | اتصال ارائه‌دهندگان و آمادگی انتشار |

موج‌ها ترتیب وابستگی پیشنهادی بسته را بیان می‌کنند. اولویت محصول طبق برنامه تحویل، تکمیل دامنه پنل‌ها با داده آزمون، سپس سایت مشتری و بررسی جداگانه اتصال واقعی است. هنگام استفاده از تسک، دامنه، وابستگی، پرسش‌های همان مسیر و اصلاحات صریح مرتبط را مرور کنید. [برنامه تحویل](README.md) · [ترتیب موج‌های بسته](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/12-delivery-plan.md) · [اصلاحات مرتبط](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/task-index-fa.md)

## ثبت نتیجه تسک

در کار اجرایی، شناسه تسک، قابلیت و سناریوی پذیرش مرتبط، پاسخ پرسش مانع و شاهد نتیجه در ردیف تحویل ثبت شوند. برای تغییر مالی و عملیاتی، شاهد شامل مسیر موفق، شکست یا نتیجه نامشخص و کنترل اثر تکراری مرتبط با همان تسک است. نوع شاهد را با عنوان بررسی رابط، آزمون با داده ساختگی، آزمون اتصال یا پذیرش عملیاتی مشخص کنید. [معیار آمادگی](readiness.md) · [راهنمای کار روی تسک](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/16-agent-delivery.md) · [گزارش کنترل فایل‌های بسته](../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/validation.md)

## اصلاحات نسخه ۱.۱

پس از پاسخ #13، [فهرست و شرح‌های فعال](review-v1-1/README.md) مرجع دامنه، وابستگی، AC و شاهد پایان هر کار است. ۱۲ Issue اصلی بازنویسی و سه نیاز تازه به #15/#16/#17 تفکیک شدند. snapshot نسخه۱.۱ بدون تغییر مرجع تاریخی است. وضعیت ۱۵ کار فعال پس از بازگشایی محدود #7 در فهرست جاری پیگیری می‌شود؛ پس از فعال‌سازی محدود #7 طبق [ADR-006](../decisions/adr-006-responsive-banner-activation.md)، فقط #8/#9 از این سه کار deferred هستند؛ همگام‌سازی #14 به معنی تکمیل محصول نیست.

| کار و Issue | موج پیشنهادی | وضعیت اجرا | اصلاح پایه/منشأ |
| --- | --- | --- | --- |
| [CV-R2-01](review-v1-1/tasks/cv-r2-01-single-user.md) · [#1](https://github.com/MohamadParsa/marketplace/issues/1) | W0 | todo | CV-IAM-04، CV-IAM-05، CV-IAM-07 |
| [CV-R2-02](review-v1-1/tasks/cv-r2-02-mobile-first-login.md) · [#2](https://github.com/MohamadParsa/marketplace/issues/2) | W0 | todo | CV-IAM-01، CV-IAM-02 |
| [CV-R2-03](review-v1-1/tasks/cv-r2-03-profile-password-recovery.md) · [#3](https://github.com/MohamadParsa/marketplace/issues/3) | W0 | todo | CV-IAM-02، CV-IAM-07، CV-WEB-07 |
| [CV-R2-04](review-v1-1/tasks/cv-r2-04-category-brand-media.md) · [#4](https://github.com/MohamadParsa/marketplace/issues/4) | W1 | review | CV-CAT-01، CV-CAT-03، CV-CAT-07 |
| [CV-R2-05](review-v1-1/tasks/cv-r2-05-shared-attributes.md) · [#5](https://github.com/MohamadParsa/marketplace/issues/5) | W1 | review | CV-CAT-02، CV-CAT-05، CV-CAT-08 |
| [CV-R2-06](review-v1-1/tasks/cv-r2-06-product-workspace.md) · [#6](https://github.com/MohamadParsa/marketplace/issues/6) | W1 | review | CV-CAT-04، CV-CAT-05، CV-CAT-06 |
| [CV-R2-07](review-v1-1/tasks/cv-r2-07-responsive-banner.md) · [#7](https://github.com/MohamadParsa/marketplace/issues/7) | W1 | review | CV-CMS-01، CV-CMS-02 |
| [CV-R2-08](review-v1-1/tasks/cv-r2-08-manual-product-rail.md) · [#8](https://github.com/MohamadParsa/marketplace/issues/8) | later | deferred | CV-CMS-01، CV-CMS-02، CV-WEB-03 |
| [CV-R2-09](review-v1-1/tasks/cv-r2-09-sticky-banner.md) · [#9](https://github.com/MohamadParsa/marketplace/issues/9) | later | deferred | CV-CMS-01، CV-CMS-02، CV-WEB-03 |
| [CV-R2-10](review-v1-1/tasks/cv-r2-10-seller-lead-snapshot.md) · [#10](https://github.com/MohamadParsa/marketplace/issues/10) | W2 | todo | CV-IAM-06، CV-CAT-06، CV-OPS-05، CV-OPS-06 |
| [CV-R2-11](review-v1-1/tasks/cv-r2-11-storefront-alignment.md) · [#11](https://github.com/MohamadParsa/marketplace/issues/11) | W3 | todo | CV-WEB-02، CV-WEB-03، CV-WEB-04، CV-WEB-05، CV-WEB-06، CV-WEB-07، CV-WEB-08 |
| [CV-R2-12](review-v1-1/tasks/cv-r2-12-revision-acceptance.md) · [#12](https://github.com/MohamadParsa/marketplace/issues/12) | W3 | todo | CV-QA-01، CV-QA-04، CV-QA-06 |
| [CV-R2-13](review-v1-1/tasks/cv-r2-13-general-settings.md) · [#15](https://github.com/MohamadParsa/marketplace/issues/15) | W1 | todo | تفکیک جدید از پاسخ مالک |
| [CV-R2-14](review-v1-1/tasks/cv-r2-14-shipment-slots.md) · [#16](https://github.com/MohamadParsa/marketplace/issues/16) | W2 | todo | تفکیک جدید از پاسخ مالک |
| [CV-R2-15](review-v1-1/tasks/cv-r2-15-stock-alert.md) · [#17](https://github.com/MohamadParsa/marketplace/issues/17) | W2 | todo | تفکیک جدید از پاسخ مالک |
| [CV-R2-16](review-v1-1/tasks/cv-r2-16-category-content-seo.md) · [#24](https://github.com/MohamadParsa/marketplace/issues/24) | reference؛ اجرا در #28 | done / مرجع | چهار AC با [شاهد #28](initial-launch/issue-28-verification.md)؛ بدون اجرای مستقل |
| [CV-R2-17](review-v1-1/tasks/cv-r2-17-product-content-seo.md) · [#25](https://github.com/MohamadParsa/marketplace/issues/25) | reference؛ اجرا در #29 | todo / مرجع | کنترل پوشش محصول؛ موارد شرطی تعامل باز می‌مانند |

<a id="issue-13-scope"></a>

## پاسخ مالک پس از #13 — محدوده آماده بازنویسی تسک

[ADR-005](../decisions/adr-005-review-v1-1-product-responses.md#اثر-بر-تسکها-و-مرحله-بعد) مرجع پاسخ و جدول کارهای متأثر است. اعداد موج/اولویت در snapshot بالا، برنامه نهایی پس از پاسخ نیستند. اثرهای زیر در شرح‌های فعال و Issueهای #14 اعمال می‌شوند؛ معیارهای این مرحله از اجرای نرم‌افزار جدا هستند:

- #7/#8/#9 و بخش CMS معوق علامت بخورند؛ برای #11/#12 dependency الزام CMS حذف یا با قالب ثابت خانه جایگزین شود. فرم سکشنِ معوق مانع شروع نیست.
- #1..3 فلو موبایل و لینک رمز، بازیابی و union مجوز scoped/عضویت چندبه‌چند را بگیرند؛ پیشنهاد تغییر موبایل تا تصویب فعال نشود.
- #5/#6/#11 انتساب بدون ارث و سقف کارت تنظیمی را بگیرند؛ مهاجرت تاریخی ساختگی به پروژه صفر اضافه نشود.
- #10/#11 روش اسلات‌دار و بازه‌ای را تفکیک کنند. ظرفیت، اعلان موجودشدن تنوع و تنظیمات عمومی به کارهای مستقل با AC و سؤال مربوط شکسته شوند؛ این نیازها اکنون در #15 تنظیمات، #16 ظرفیت و #17 خبرم‌کن تفکیک شده‌اند؛ اولویت‌ها پیشنهادی‌اند و اجرا شروع نشده است.
- اتصال میان زیرسایت‌ها از محدوده شروع خارج و A پنل در پذیرش ثبت شود. رفتار خاموش‌بودن خود PDP و خرید مستقیم صفحه فروشگاه در Q-027 باز باقی است.

تغییرات حوزه به معنی تکمیل کد/API یا پذیرش سناریوهای قبلی نیستند. مرحله۲ باید سناریوها، قراردادها، وابستگی و شواهد لازم بستن هر کار را یکپارچه کند.


تکمیل۲۰۲۶-۱۱-۲۱: دو تسک محتوای دسته/محصول به درخواست مالک افزوده شد؛ [مرجع تطبیق و فیلدها](../domains/catalog/content-and-seo.md). فهرست جاری۱۷ کار (۱۵ فعال،۲معوق) دارد؛ شمار۹۱ تسک snapshot تاریخی تغییر نکرده است.
