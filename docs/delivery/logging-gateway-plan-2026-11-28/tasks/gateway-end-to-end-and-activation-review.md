---
id: FEAT-PAY-GATEWAY-ACCEPTANCE-001
title: پذیرش یکپارچه زرین‌پال و راهنمای فعال‌سازی
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: deferred
execution_authorized: false
---

# CV-PAY-GW-09 — پذیرش یکپارچه زرین‌پال و راهنمای فعال‌سازی

Issue: [#163](https://github.com/MohamadParsa/marketplace/issues/163) · مسئول: `eliassarband` · معوق به دستور مالک تا آماده‌شدن لاگ‌گیری و انتخاب دوباره اجرا.

## مسئله و نتیجه

شواهد end-to-end خرید و شارژ، پوشش کامل لاگ و راهنمای callback دامنه اصلی هر محیط، credential/واحد/ضریب/فعال‌سازی، مدیریت خطا و خاموش‌سازی. mock، Sandbox و محیط واقعی جدا گزارش شوند و هیچ ادعای آمادگی بانکی از تست ساختگی نشود.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: merge/deploy عمومی، دریافت وجه واقعی، قرارداد تجاری یا بستن خودکار گیت انتشار #12.
- وابستگی: [#159](https://github.com/MohamadParsa/marketplace/issues/159)، [#161](https://github.com/MohamadParsa/marketplace/issues/161)، [#162](https://github.com/MohamadParsa/marketplace/issues/162)

## مسیر و کنترل‌های لازم

پیش از شروع، پوشش لاگ آماده باشد؛ عملیات بیرونی با تلاش قابل بازیابی، اطلاعات ثابت همان تلاش و ثبت امن شروع/نتیجه انجام شود. callback دامنه اصلی سایت هنگام ساخت request تولید می‌شود.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC01:** ماتریس خرید/اعتبار/کوپن/مبلغ صفر/شارژ مستقل/لغو/دیررس/تکرار/عدم بازگشت، با trace قابل پیگیری و داده ساختگی کامل باشد.
- **FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC02:** برای همه تماس‌ها و اکشن‌های auth/payment/recovery، زمان/تلاش/نتیجه بدون اطلاعات حساس در ذخیره محلی جست‌وجو شود.
- **FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC03:** Sandbox فقط در محیط ایزوله و پس از انتخاب اجرا؛ تست واقعی فقط با شرایط GATEWAY-Q02/Q03 و دستور جدا، به همراه ثبت دامنه callback در پذیرنده.
- **FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC04:** راهنمای اپراتور و محدودیت‌های باقی‌مانده، commit/push/CI و پذیرش مالک ثبت شوند؛ شروع پردازش و استرداد فعلی با اتصال جدید سازگار بمانند.

## روش ارزیابی و تحویل آینده

API/race/PostgreSQL و مرورگر؛ گزارش تفکیک‌شده fixture/Sandbox/live با مجوز لازم در زمان اجرا.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

GATEWAY-Q01/Q02/Q03 و انتخاب مجدد مالک برای اجرا؛ این برنامه مجوز اتصال بانکی یا انتشار نیست.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.
