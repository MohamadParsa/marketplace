---
id: FEAT-LOG-OPERATIONS-001
title: پوشش مسیر خرید، عملیات حساس و تماس سرویس‌های بیرونی
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# CV-LOG-03 — پوشش مسیر خرید، عملیات حساس و تماس سرویس‌های بیرونی

Issue: [#157](https://github.com/MohamadParsa/marketplace/issues/157) · مسئول: `eliassarband` · فقط برنامه؛ شروع کدنویسی در درخواست فعلی انتخاب نشده است.

## مسئله و نتیجه

ممیزی جدول رویداد/سورس و اتصال زیرساخت مشترک به مسیرهای فعلی: محاسبه/کوپن/اعتبار، PlaceOrder/رزرو/پرداخت صفر، شارژ/مصرف، مراحل سفارش/مرسوله، کنسلی/استرداد، auth/دسترسی و تماس‌های موجود پیامک/نقشه. wrapper تماس بیرونی با شروع/پایان، provider/operation، elapsed/status/error-code، timeout/retry/attempt و correlation؛ تمام اکشن‌های درگاه آینده همین قرارداد را مصرف کنند.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: پیاده‌سازی تماس بانکی/OAuth یا فعال‌سازی سرویس واقعی تازه؛ ثبت رفتار مشتری مالک #76 و audit تغییرات مالک #75 است.
- وابستگی: [#155](https://github.com/MohamadParsa/marketplace/issues/155)، [#156](https://github.com/MohamadParsa/marketplace/issues/156)

## مسیر و کنترل‌های لازم

رویداد معنایی از context معتبر تولید، پیش از هر مقصد پالایش و با شناسه پیگیری و نتیجه ثبت شود؛ جزئیات نوع/سطح/فیلدها و تفکیک audit/ledger/behavior در قرارداد مشترک است.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-LOG-OPERATIONS-001-AC01:** برای مسیر موفق، رد اعتبارسنجی، مجوز، conflict، timeout و نتیجه نامشخص رویداد معنایی درست با operation مشترک ثبت شود؛ کلیک یا HTTP200 معادل موفقیت مالی نباشد.
- **FEAT-LOG-OPERATIONS-001-AC02:** هر تماس واقعی به provider ثبت مستقل شروع/پایان یا شروع بدون پایان قابل بازیابی، زمان و attempt داشته باشد؛ درخواست/پاسخ فقط فیلدهای مجاز و نه dump خام باشند.
- **FEAT-LOG-OPERATIONS-001-AC03:** نمونه خرید از PlaceOrder تا credit/debit یا شکست با شناسه خرید/سفارش/تلاش و منشأ خطا دنبال شود؛ replay با رخداد replay از اثر تجاری تازه جدا شود.
- **FEAT-LOG-OPERATIONS-001-AC04:** پوشش اکشن‌های آماده درگاه (initialize/token/refresh/activation/request/redirect/callback/verify/inquiry/retry/reconciliation و موارد مالی آینده در صورت پیاده‌سازی) در ماتریس ثبت شود؛ هیچ تماس فعال بدون instrumentation پذیرفته نشود.

## روش ارزیابی و تحویل آینده

HTTP ساختگی بیرونی و integration خرید/مجوز، پیش‌نویس ماتریس پوشش سورس و تست؛ برای سرویس اجرا‌نشده فقط قرارداد/fixture، نه ادعای اتصال واقعی.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

NEXT-Q09؛ مسئول لاگ فنی #77 است و این تسک اجرای پوشش فنی آن است.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.

## شاهد اجرا

ثبت runtime/logger مشترک به lifecycle برنامه، context HTTP، actor معتبر احراز، عملیات commerce/preview/command و تماس واقعی adapterهای پیامک/نشان وصل شد. metadata و نتیجه سرویس ثبت و body/URL/توکن خصوصی ذخیره نمی‌شوند. تست HTTP correlation/replay/حریم، خطای فایل بدون توقف سبد، provider با HTTP200 ولی نتیجه ناموفق و همه Go/PG/Redis و vet موفق‌اند؛ درگاه جدیدی ساخته نشد. [ماتریس پوشش](../coverage.md).
