# کش، امنیت، SEO، لاگ و آمادگی عملیاتی

نسخه: ۱.۰ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: تحلیل و طراحی برای بازبینی تیم فنی | مالک بررسی: محصول و راهبر فنی؛ انتساب شخص هنوز انجام نشده

## سیاست کش پیشنهادی

| داده / مرحله | سیاست | ابطال / کنترل |
| --- | --- | --- |
| فایل تصویر/فونت/asset دارای hash | CDN بلندمدت immutable | نام نسخه جدید؛ محدودیت mime و origin |
| صفحه و سکشن عمومی | کش با scope+page+publication_revision | انتشار، تغییر ترتیب و مجازبودن سکشن؛ preview خصوصی no-store |
| دسته/برند/خصوصیت | کش عمومی versioned | ویرایش ساختار/آرشیو/خصوصیت؛ کلید شامل قلمرو |
| PLP/search/suggest | TTL کوتاه پیشنهادی، query+filter+sort+page+scope+نسخه داده | حذف عرضه/قیمت/تخفیف در زمان معتبر؛ PII وارد کلید عمومی نشود |
| PDP محتوایی | محتوای نسخه‌دار؛ قیمت/موجودی/eligibility جدا و تازه | quote سروری مجوز خرید؛ قیمت کهنه مجوز checkout نیست |
| نرخ طلا و quote | version+source+seller+method و محدود به valid_to | زمان دقیق CHQ-03؛ stale-rate به خرید نرود |
| سبد/checkout/OTP/profile/wallet/پنل | private/no-store برای پاسخ حساس | logout/تعویض حساب پاک‌سازی؛ CDN/SW کش نکنند |
| PWA | app shell و داده عمومی منتخب؛ نوشتن/مالی offline مسدود | نسخه SW و پاک‌کردن داده حساب قبلی؛ الزام PWA سلر با پنل مشترک حفظ شود |

TTLهای عددی production بعد از اندازه‌گیری و تعیین اعتبار نرخ انتخاب شوند؛ مقدار کش پروژه گوشی‌شاپ به طلا تعمیم داده نشود. db برای صحت مالی مرجع است؛ Redis قفل/ledger جایگزین نیست. هدف عملی ابطال، cache namespace/version مرتبط است؛ flush کل سامانه در هر ویرایش لازم نیست.

## SEO و مسیرها

HTML قابل crawl برای Home/PLP/PDP، title/description، breadcrumb، structured data معتبر محصول/عرضه، canonical، sitemap فقط URLهای منتشر، robots و noindex حساب/سبد/جست‌وجوی کم‌ارزش/فیلترهای تکراری طبق سیاست. URL slug پایدار و redirect تغییر slug؛ 404/410 مطابق وضعیت واقعی. قیمت/موجودی schema با عرضه واقعی یکسان باشد؛ امتیاز یا فروش ساختگی داده نشود. canonical مادر/زیرسایت و product URL در CHQ-08 پیش از انتشار نهایی شوند. SEO زیرساخت بخشی از طراحی است و به پایان کار موکول نشود.

## لاگ، audit و گزارش

structured log با correlation_id، actor_ref، resource_ref، event، duration، result_code؛ OTP، session، token، اطلاعات کارت و payload کامل مشتری ذخیره نشود. audit برای تغییر نرخ/تخفیف/مجوز/قیمت/موجودی/سفارش/خروج وجه با عامل، قبل/بعد محدود، علت و زمان. export هم permission و scope دارد. log عملیاتی جای ledger مالی نیست. retention و دسترسی audit/backup در CHQ-13 است.

شاخص‌ها: latency/error جست‌وجو، نرخ quote نامعتبر، رزرو منقضی، صف pending/unknown، اختلاف درگاه/ledger، تأخیر outbox، fail پیامک، کمبود تأمین و مانده تسویه. dashboard فروش فقط از پرداخت قطعی و قوانین refund مشخص محاسبه کند؛ سنجه نمونه به KPI واقعی تبدیل نشود.

## امنیت و قابلیت نگهداری

TLS، secret خارج repo، least privilege DB/provider، rate limit OTP/search/write، schema validation، parameterized SQL، upload محدود و sanitize HTML، CSRF/CORS/host allowlist. فروشگاه/کاربر/resource از داده سرور resolve شوند. role revocation و logout بر API مؤثر باشند؛ cached role نشستی بدون بازبینی اختیار حساس را زنده نگذارد. تهدیدهای IDOR، تزریق، callback جعلی، replay، فایل مخرب و خروجی CSV قابل اجرای فرمول بررسی شوند.

## استقرار و مجوزها

محیط local/test/sandbox/production تفکیک؛ fixture و درگاه آزمون در production غیرفعال. تنظیمات validated و healthcheck امن؛ migration مرحله‌ای/backup/restore/rollback، runbook reconciliation و پشتیبانی. دامنه ثبت‌شده به معنی DNS/TLS/اینماد فعال نیست. نماد/صلاحیت طلای فیزیکی، مالیات/فاکتور، پذیرندگی و حمل واقعی به مسئول کسب‌وکار سپرده شوند؛ سند حکم قانونی صادر نمی‌کند. آدرس host/provider/secret تولیدی در نمونه ساخته نشود.

برای کارایی، حجم dataset، کاربر هم‌زمان، P95 و زمان بازیابی را تیم با محصول در CHQ-13 انتخاب کند. بودجه سرور و SLO قطعی حدس زده نشده؛ تست باید با dataset قابل تکرار و هدف ثبت‌شده اجرا شود.
