---
id: FEAT-PAY-GATEWAY-ADAPTER-001
title: اتصال request و verify زرین‌پال با تلاش نسخه‌دار
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: deferred
execution_authorized: false
---

# CV-PAY-GW-06 — اتصال request و verify زرین‌پال با تلاش نسخه‌دار

Issue: [#160](https://github.com/MohamadParsa/marketplace/issues/160) · مسئول: `eliassarband` · معوق به دستور مالک تا آماده‌شدن لاگ‌گیری و انتخاب دوباره اجرا.

## مسئله و نتیجه

adapter V4 و ثبت تلاش شامل provider/environment/version، مبلغ/واحد/ضریب/شرح وابسته به عملیات، authority/ref و نتیجه؛ request، redirect، callback و verify سروری با اطلاعات ثابت همان تلاش. نتیجه مالی از query مرورگر گرفته نشود و HTTP بیرونی داخل تراکنش طولانی DB نباشد؛ credential لازم هر عملیات کنترل شود.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: OAuth مدیریتی غیرضروری به‌عنوان پیش‌شرط عمومی خرید، انتشار/پرداخت واقعی و درج ledger بدون کنترل تسک هماهنگ‌سازی.
- وابستگی: [#157](https://github.com/MohamadParsa/marketplace/issues/157)، [#158](https://github.com/MohamadParsa/marketplace/issues/158)

## مسیر و کنترل‌های لازم

پیش از شروع، پوشش لاگ آماده باشد؛ عملیات بیرونی با تلاش قابل بازیابی، اطلاعات ثابت همان تلاش و ثبت امن شروع/نتیجه انجام شود. callback دامنه اصلی سایت هنگام ساخت request تولید می‌شود.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-PAY-GATEWAY-ADAPTER-001-AC01:** request ساختگی با مبلغ تبدیل‌شده دقیق، شرح خرید/شارژ و callback دامنه سایت ساخته شود؛ authority پیش از redirect ثبت و request تکراری تلاش ناخواسته دوم نسازد.
- **FEAT-PAY-GATEWAY-ADAPTER-001-AC02:** callback با correlation سروری بدون نیاز به نشست مرورگر به تلاش درست برسد؛ verify با authority/مبلغ/اتصال ثبت‌شده انجام شود.
- **FEAT-PAY-GATEWAY-ADAPTER-001-AC03:** کد100/101، رفض/لغو، پاسخ نامعتبر، timeout و تغییر تنظیمات وسط خرید پوشش داده شوند؛101 مجوز credit دوباره نباشد.
- **FEAT-PAY-GATEWAY-ADAPTER-001-AC04:** تمام اکشن‌ها و تماس‌ها لاگ نوع integration/technical با outcome درست و بدون secret داشته باشند؛ Sandbox به دفتر واقعی متصل نشود.

## روش ارزیابی و تحویل آینده

HTTP mock و PostgreSQL/رقابت/قطع پاسخ؛ اتصال بانکی واقعی فقط در تسک پذیرش بعدی.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

GATEWAY-Q02/Q03 برای شرایط واقعی؛ اجرای این تسک پس از انتخاب دوباره مالک.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.
