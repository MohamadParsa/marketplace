# ردیابی قابلیت تا پذیرش

این جدول فقط نقشه لینک‌هاست؛ قانون و سناریو در محل مرجع نگهداری می‌شوند. همه قابلیت‌ها فعلاً در مرحله مستندات‌اند؛ کد، PR یا آزمون اجراشده‌ای وجود ندارد. شناسه‌های `AC` داخل سند قابلیت، شواهد طراحی پذیرش‌اند، نه گزارش آزمون موفق.

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
