---
id: FEAT-PAY-GATEWAY-RECOVERY-001
title: بازیابی نتیجه نامشخص و retry امن پرداخت
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: deferred
execution_authorized: false
---

# CV-PAY-GW-08 — بازیابی نتیجه نامشخص و retry امن پرداخت

Issue: [#162](https://github.com/MohamadParsa/marketplace/issues/162) · مسئول: `eliassarband` · معوق به دستور مالک تا آماده‌شدن لاگ‌گیری و انتخاب دوباره اجرا.

## مسئله و نتیجه

پیگیری پایدار تلاش‌ها با inquiry و verify در جای صحیح، زمان‌بندی retry تحت تنظیم هر درگاه، کنترل هم‌زمانی، restart و شکست ذخیره پس از پاسخ. عدم بازگشت مشتری تنها مسیر پیگیری نباشد؛ timeout درخواست بدون authority با دوباره‌ارسال کور به دریافت وجه دوم تبدیل نشود.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: پرداخت دوباره خودکار، معادل‌دانستن inquiry با verify یا حدس مهلت بانکی از مهلت رزرو سفارش.
- وابستگی: [#160](https://github.com/MohamadParsa/marketplace/issues/160)، [#161](https://github.com/MohamadParsa/marketplace/issues/161)

## مسیر و کنترل‌های لازم

پیش از شروع، پوشش لاگ آماده باشد؛ عملیات بیرونی با تلاش قابل بازیابی، اطلاعات ثابت همان تلاش و ثبت امن شروع/نتیجه انجام شود. callback دامنه اصلی سایت هنگام ساخت request تولید می‌شود.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-PAY-GATEWAY-RECOVERY-001-AC01:** قطع مرورگر، callback نرسیده، timeout verify و شکست DB پس از verify تا نتیجه قابل توضیح و اثر مالی یکتا بازیابی شوند.
- **FEAT-PAY-GATEWAY-RECOVERY-001-AC02:** inquiry فقط وضعیت را گزارش دهد؛ PAID/VERIFIED/IN_BANK/FAILED/REVERSED مطابق قرارداد تفسیر و نهایی‌سازی با کنترل مالی انجام شود.
- **FEAT-PAY-GATEWAY-RECOVERY-001-AC03:** آستانه retry و توقف/ارجاع بررسی با attempt، علت، زمان و correlation در لاگ مشخص باشند؛ دو worker یک تلاش را هم‌زمان مؤثر اجرا نکنند.
- **FEAT-PAY-GATEWAY-RECOVERY-001-AC04:** تغییر/غیرفعال‌سازی درگاه مانع پرداخت جدید شود ولی پیگیری تلاش قبلی با نسخه اتصال خودش ممکن بماند؛ نتیجه ناشناخته موفق یا شکست قطعی جعل نشود.

## روش ارزیابی و تحویل آینده

آزمون crash/restart/timeout/concurrency با clock و provider ساختگی؛ شاهد عدم credit/debit مضاعف.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

GATEWAY-Q03 برای مهلت، backoff و سقف پیگیری؛ GATEWAY-Q02 برای محدودیت ترمینال واقعی.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.
