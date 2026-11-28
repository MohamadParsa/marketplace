---
id: FEAT-PAY-GATEWAY-CALLBACK-001
title: ساخت callback هنگام پرداخت با دامنه اصلی سایت
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: deferred
execution_authorized: false
---

# CV-PAY-GW-04 — ساخت callback هنگام پرداخت با دامنه اصلی سایت

Issue: [#158](https://github.com/MohamadParsa/marketplace/issues/158) · مسئول: `eliassarband` · معوق به دستور مالک تا آماده‌شدن لاگ‌گیری و انتخاب دوباره اجرا.

## مسئله و نتیجه

در ساخت هر تلاش پرداخت، callback از origin معتبر و پیکربندی‌شده دامنه اصلی storefront همان محیط + مسیر ثابت کد ساخته و با تلاش snapshot شود؛ بدون فیلد ورود callback برای اپراتور و بدون وابستگی به دامنه بک‌آفیس/API یا Host کنترل‌نشده درخواست. request/callback lifecycle رویداد مجاز داشته باشد.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: انتخاب دامنه واقعی قرارداد، ساخت callback مالی مستقل از adapter و آغاز اتصال بانکی قبل از زیرساخت لاگ.
- وابستگی: [#157](https://github.com/MohamadParsa/marketplace/issues/157)

## مسیر و کنترل‌های لازم

پیش از شروع، پوشش لاگ آماده باشد؛ عملیات بیرونی با تلاش قابل بازیابی، اطلاعات ثابت همان تلاش و ثبت امن شروع/نتیجه انجام شود. callback دامنه اصلی سایت هنگام ساخت request تولید می‌شود.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-PAY-GATEWAY-CALLBACK-001-AC01:** در توسعه/استیج/پروداکشن ساخت URL بدون تغییر کد و در لحظه request انجام شود؛ در محیط عمومی origin سایت HTTPS باشد و API/backoffice origin جای آن ننشیند.
- **FEAT-PAY-GATEWAY-CALLBACK-001-AC02:** تغییر Host/Forwarded headers یا query مقصد callback را عوض نکند؛ origin نامعتبر موجب خطای مشخص پیش از تماس بیرونی شود.
- **FEAT-PAY-GATEWAY-CALLBACK-001-AC03:** تلاش آغازشده با callback/تنظیمات خودش بازیابی شود و تغییر دامنه/تنظیم جاری مسیر تلاش قبلی را بازنویسی نکند؛ handler روی دامنه سایت واقعاً route شود.
- **FEAT-PAY-GATEWAY-CALLBACK-001-AC04:** آزمون config/URL و request fixture دامنه درست را اثبات کند؛ log عمومی حاوی token/query حساس نباشد.

## روش ارزیابی و تحویل آینده

unit ساخت URL و integration درخواست/route با دامنه‌های ساختگی؛ GATEWAY-Q02 برای ثبت دامنه واقعی پذیرنده.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

GATEWAY-Q02؛ اصل دامنه اصلی و زمان ساخت صریحاً توسط مالک انتخاب شده است.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.
