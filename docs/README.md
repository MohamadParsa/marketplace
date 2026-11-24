# نقشه مستندات

[برنامه۳۰درخواست جدید؛ وضعیت اجرا و شواهد انتخاب‌های مالک](delivery/next-phase/README.md)

[صف۱۰تسک تجربه فروشگاه؛ وضعیت و شواهد اجرای انتخاب‌های مالک](delivery/storefront-followups/ux-followups.md)

[کش، URL و صف اصلاحات تجربه storefront](delivery/storefront-followups/README.md) · [گزارش و تصاویر #41](delivery/storefront-followups/issue-41-verification.md).

[داده تست واقعی کاتالوگ، JSON و پرامپت ورود](delivery/catalog-seed/README.md) · [نتیجه ورود کامل به لوکال](delivery/catalog-seed/local-import-2026-11-23.md)

[راهنمای اجرای لوکال از D:\Cherriva](local-development.md)

[سورس Go/React و راه‌اندازی در D:\Cherriva](../README.md) · [تطبیق نسخه سورس و مدارک](delivery/source-document-map.md) · [اصلاح خرید، ویزارد و استان/شهر](delivery/initial-launch/purchase-refinements-verification.md)

| نیاز خواننده | مرجع |
| --- | --- |
| مالکیت اجرا و رفع هم‌پوشانی #24/#25 با #28/#29 | [صف یکتا؛ چهار برش محلی تحویل و گیت عمومی باز است](delivery/initial-launch/README.md#execution-ownership) |
| صف مقدم انتشار اولیه | [چهار کار بک‌آفیس/API/سایت: خانه، PLP، PDP، خرید کارت‌به‌کارت و کیف پول](delivery/initial-launch/README.md) |
| پاسخ‌های الیاس برای تسک #13 و دامنه شروع نسخه۱.۱ | [متن کامل، مثال‌ها، پیشنهاد تغییر موبایل و اثر بر تسک‌ها](decisions/adr-005-review-v1-1-product-responses.md) |
| مرور مشترک طرح‌ها پیش از نهایی‌کردن تسک‌ها | [بسته یکپارچه سایت و بک‌آفیس، تصاویر و راهنمای مشاهده](review-packages/README.md) |
| فیلدهای تکمیلی دسته/محصول و SEO | [تطبیق مدل و تسک‌های #24/#25](domains/catalog/content-and-seo.md) |
| ادامه سرویس‌های نظر و پرسش‌وپاسخ پس از #25 | [مرز و پذیرش CV-WEB-15/16](delivery/product-interactions-followup.md) |
| شناخت محصول و مرز دامنه | [دامنه](product/scope.md)، [واژه‌نامه](product/glossary.md) |
| شناخت نقش و داده مجاز | [نقش‌ها](product/roles.md) |
| پیدا کردن قانون مشترک | [حوزه‌ها](domains/README.md) |
| طراحی تجربه یک محیط | [تجربه‌ها](experiences/README.md) |
| درک مرز فنی و روابط | [معماری](architecture/README.md)، [مدل منطقی](architecture/data-model.md) |
| طراحی ارتباطات و آزمون شکست | [قراردادها](architecture/contracts/README.md)، [کیفیت](architecture/quality.md) |
| تصمیم‌گیری روی ابهام | [پرسش‌ها](product/open-questions.md) |
| فهم اصلاح تصمیم قبلی | [دفتر تصمیم‌ها](decisions/README.md) |
| تسک‌های بازنویسی‌شده پس از پاسخ #13 | [۱۷ شرح فعال/معوق، وابستگی و AC و شاهد پایان](delivery/review-v1-1/README.md) |
| انتخاب مسیر ساخت | [برش‌ها](delivery/README.md)، [رودمپ](delivery/roadmap.md)، [آمادگی](delivery/readiness.md) |
| ردیابی تسک‌های تحلیل پایه | [نقشه ۹۱ تسک و موج‌های پیشنهادی](delivery/analysis-backlog.md) |
| اجرای سرویس و CI/CD | [راهنمای استقرار Liara](delivery/liara-deployment.md)، [ADR-003](decisions/adr-003-liara-deployment.md) |
| طراحی بصری سایت مشتری | [طرح منتخب متعادل و شیک چریوا، نسخه ۰.۴](experiences/storefront/visual-design.md) |
| طراحی بک‌آفیس و پنل فروشنده | [نقشه صفحات، فرم‌ها و مسیرها](experiences/backoffice/panel-design.md) |
| بررسی رابط سایت مشتری | [نسخه بررسی Cherriva](delivery/storefront-review.md) |
| پیگیری پوشش نیازمندی | [ردیابی](governance/traceability.md) |
| سناریوهای رفتاری و داده آزمون | [BDD](bdd/README.md) |
| افزودن سند | [راهنما](governance/documentation-guide.md)، [قالب‌ها](templates/README.md) |

[دسترسی حساب ادمین نسخه محلی تیم](delivery/local-review-access.md)

[URL عمومی chr/slug، نگاشت کامل، مهاجرت و شواهد #42](delivery/storefront-followups/issue-42-verification.md)

## مسیر مطالعه برای یک قابلیت

دامنه و نقش ← قواعد حوزه مرتبط ← تجربه محیط ← قرارداد و کیفیت ← پرسش‌های مرتبط ← آمادگی همان برش. خواندن کل مخزن برای هر تغییر لازم نیست.

مرجع اصلاح مهم قیمت‌گذاری، [ADR-001](decisions/adr-001-pricing-and-payment-boundaries.md) است؛ از تصمیم قدیمی حفظ اجباری قانون روی قیمت پایه جدید استفاده نکنید.

[گزارش خرید و کیف پول #30](delivery/initial-launch/issue-30-verification.md) · [تحویل انتشار به پارسا](delivery/initial-launch/production-handoff.md)

[اجرای محلی کوپن، اعتبار انتخابی کیف پول و گروه گارانتی #67/#68/#69/#78](delivery/next-phase/issues-67-68-69-78-verification.md)

[فعال‌سازی محلی افزودن به سبد PDP پس از پوش PR #94](delivery/storefront-followups/pdp-add-enabled-verification.md).

[ممیزی کل تسک‌ها، درخواست‌های تازه و دریافت کار پارسا](delivery/backlog-review-2026-11-24/README.md) · [صف یکتای GitHub #95](https://github.com/MohamadParsa/marketplace/issues/95).

[تحویل محلی #103/#104/#102؛ فاصله‌ها، Skeleton و شناسه عمومی chr](delivery/backlog-review-2026-11-24/issues-102-103-104-verification.md).
