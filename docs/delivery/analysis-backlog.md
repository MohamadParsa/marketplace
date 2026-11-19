# نقشه تسک‌های تحلیل پایه چریوا

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

۱۲ اصلاح زیر همراه با تسک‌های پایه متأثر بررسی می‌شوند. وضعیت ثبت‌شده در بسته برای همه آن‌ها `todo` است.

| اصلاح | موج | تسک‌های پایه متأثر | مرجع جاری |
| --- | --- | --- | --- |
| [CV-R2-01](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-01-single-user.md) | W0 | CV-IAM-04, CV-IAM-05, CV-IAM-07 | [سند](../domains/identity/README.md) |
| [CV-R2-02](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-02-phone-otp-link.md) | W0 | CV-IAM-01, CV-IAM-02 | [سند](../domains/identity/README.md) |
| [CV-R2-03](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-03-profile-email-password.md) | W0 | CV-IAM-02, CV-IAM-07, CV-WEB-07 | [سند](../domains/identity/README.md) |
| [CV-R2-04](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-04-category-brand-media.md) | W1 | CV-CAT-01, CV-CAT-03, CV-CAT-07 | [سند](../domains/catalog/README.md) |
| [CV-R2-05](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-05-shared-attribute-bindings.md) | W1 | CV-CAT-02, CV-CAT-05, CV-CAT-08 | [سند](../domains/catalog/README.md) |
| [CV-R2-06](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-06-product-workspace.md) | W1 | CV-CAT-04, CV-CAT-05, CV-CAT-06 | [سند](../experiences/backoffice/management.md) |
| [CV-R2-07](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-07-responsive-banner-section.md) | W2 | CV-CMS-01, CV-CMS-02 | [سند](../experiences/backoffice/management.md) |
| [CV-R2-08](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-08-manual-product-rail.md) | W2 | CV-CMS-01, CV-CMS-02, CV-WEB-03 | [سند](../experiences/backoffice/management.md) |
| [CV-R2-09](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-09-sticky-banner-section.md) | W2 | CV-CMS-01, CV-CMS-02, CV-WEB-03 | [سند](../experiences/backoffice/management.md) |
| [CV-R2-10](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-10-seller-supply-lead.md) | W2 | CV-IAM-06, CV-CAT-06, CV-OPS-05, CV-OPS-06 | [سند](../domains/shipping/README.md) |
| [CV-R2-11](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-11-storefront-alignment.md) | W3 | CV-WEB-02, CV-WEB-03, CV-WEB-04, CV-WEB-05, CV-WEB-06, CV-WEB-07, CV-WEB-08 | [سند](../experiences/storefront/discovery.md) |
| [CV-R2-12](../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/tasks/cv-r2-12-revision-acceptance.md) | W3 | CV-QA-01, CV-QA-04, CV-QA-06 | [سند](readiness.md) |
