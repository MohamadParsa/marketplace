---
id: FEAT-LOG-FOUNDATION-001
title: قرارداد مشترک لاگ، شناسه پیگیری و حذف داده حساس
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: todo
execution_authorized: false
---

# CV-LOG-01 — قرارداد مشترک لاگ، شناسه پیگیری و حذف داده حساس

Issue: [#155](https://github.com/MohamadParsa/marketplace/issues/155) · مسئول: `eliassarband` · فقط برنامه؛ شروع کدنویسی در درخواست فعلی انتخاب نشده است.

## مسئله و نتیجه

تعریف schema نسخه‌دار و API مشترک ثبت رویداد برای API و workerها؛ اتصال به logger موجود Go و تفکیک log_type از severity؛ request/operation/parent-operation و انتشار context بین مراحل. اعتبارسنجی و allowlist فیلدها، پالایش مرکزی پیش از هر مقصد و سقف payload/ورودی بدون ساخت logger موازی.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: انتخاب Elastic/ClickHouse، ذخیره‌ساز نهایی، اتصال درگاه و داشبورد تحلیلی.
- وابستگی: ندارد؛ قراردادهای مجوز/داده موجود پیش‌شرط‌اند.

## مسیر و کنترل‌های لازم

رویداد معنایی از context معتبر تولید، پیش از هر مقصد پالایش و با شناسه پیگیری و نتیجه ثبت شود؛ جزئیات نوع/سطح/فیلدها و تفکیک audit/ledger/behavior در قرارداد مشترک است.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-LOG-FOUNDATION-001-AC01:** یک عملیات ساختگی با یک operation_id در درخواست، سرویس و کار پس‌زمینه دنبال شود؛ request_id هر درخواست مستقل و provider_attempt_id تلاش بیرونی را جدا کند.
- **FEAT-LOG-FOUNDATION-001-AC02:** نوع‌های technical/integration/business/security/behavior از سطح debug/info/warn/error و نتیجه started/succeeded/rejected/failed/pending/unknown/cancelled جدا باشند؛ زمان رخداد و ثبت، نسخه و service/environment/surface مشخص باشند.
- **FEAT-LOG-FOUNDATION-001-AC03:** در تست تو در تو و خطا، password/OTP/token/Authorization/cookie/client_secret/API-key/DSN/کارت کامل و بدنه خصوصی در هیچ مقصد یا fallback ظاهر نشوند؛ داده بزرگ و log injection کنترل شوند.
- **FEAT-LOG-FOUNDATION-001-AC04:** فیلدهای مشتری/مهمان/فروشگاه از context معتبر سرور باشند؛ داده مرورگر trusted تلقی نشود و logger مالک ledger یا مجوز نباشد.

## روش ارزیابی و تحویل آینده

unit قرارداد/پالایش/اعتبارسنجی، HTTP context و worker؛ داده آزمون ساختگی. پیش‌نیاز مشترک #75/#76 و زیرتسک #77.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

NEXT-Q09 برای جزئیات مجاز و نقش مشاهده؛ تصمیم‌های مستقل روشن قبل از اجرا قابل پیشبردند.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.
