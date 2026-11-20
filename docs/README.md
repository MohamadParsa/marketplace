# نقشه مستندات

| نیاز خواننده | مرجع |
| --- | --- |
| پاسخ‌های الیاس برای تسک #13 و دامنه شروع نسخه۱.۱ | [متن کامل، مثال‌ها، پیشنهاد تغییر موبایل و اثر بر تسک‌ها](decisions/adr-005-review-v1-1-product-responses.md) |
| مرور مشترک طرح‌ها پیش از نهایی‌کردن تسک‌ها | [بسته یکپارچه سایت و بک‌آفیس، تصاویر و راهنمای مشاهده](review-packages/README.md) |
| شناخت محصول و مرز دامنه | [دامنه](product/scope.md)، [واژه‌نامه](product/glossary.md) |
| شناخت نقش و داده مجاز | [نقش‌ها](product/roles.md) |
| پیدا کردن قانون مشترک | [حوزه‌ها](domains/README.md) |
| طراحی تجربه یک محیط | [تجربه‌ها](experiences/README.md) |
| درک مرز فنی و روابط | [معماری](architecture/README.md)، [مدل منطقی](architecture/data-model.md) |
| طراحی ارتباطات و آزمون شکست | [قراردادها](architecture/contracts/README.md)، [کیفیت](architecture/quality.md) |
| تصمیم‌گیری روی ابهام | [پرسش‌ها](product/open-questions.md) |
| فهم اصلاح تصمیم قبلی | [دفتر تصمیم‌ها](decisions/README.md) |
| انتخاب مسیر ساخت | [برش‌ها](delivery/README.md)، [رودمپ](delivery/roadmap.md)، [آمادگی](delivery/readiness.md) |
| ردیابی تسک‌های تحلیل پایه | [نقشه ۹۱ تسک و موج‌های پیشنهادی](delivery/analysis-backlog.md) |
| اجرای سرویس و CI/CD | [راهنمای استقرار Liara](delivery/liara-deployment.md)، [ADR-003](decisions/adr-003-liara-deployment.md) |
| طراحی بصری سایت مشتری | [طرح منتخب متعادل و شیک چریوا، نسخه ۰.۴](experiences/storefront/visual-design.md) |
| طراحی بک‌آفیس و پنل فروشنده | [نقشه صفحات، فرم‌ها و مسیرها](experiences/backoffice/panel-design.md) |
| بررسی رابط سایت مشتری | [نسخه بررسی Cherriva](delivery/storefront-review.md) |
| پیگیری پوشش نیازمندی | [ردیابی](governance/traceability.md) |
| سناریوهای رفتاری و داده آزمون | [BDD](bdd/README.md) |
| افزودن سند | [راهنما](governance/documentation-guide.md)، [قالب‌ها](templates/README.md) |

## مسیر مطالعه برای یک قابلیت

دامنه و نقش ← قواعد حوزه مرتبط ← تجربه محیط ← قرارداد و کیفیت ← پرسش‌های مرتبط ← آمادگی همان برش. خواندن کل مخزن برای هر تغییر لازم نیست.

مرجع اصلاح مهم قیمت‌گذاری، [ADR-001](decisions/adr-001-pricing-and-payment-boundaries.md) است؛ از تصمیم قدیمی حفظ اجباری قانون روی قیمت پایه جدید استفاده نکنید.
