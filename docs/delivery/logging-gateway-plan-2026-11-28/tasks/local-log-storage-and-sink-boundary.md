---
id: FEAT-LOG-LOCAL-SINK-001
title: ذخیره محلی لاگ، چرخش فایل و امکان تغییر مقصد
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-016, DOM-PAYMENTS-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# CV-LOG-02 — ذخیره محلی لاگ، چرخش فایل و امکان تغییر مقصد

Issue: [#156](https://github.com/MohamadParsa/marketplace/issues/156) · مسئول: `eliassarband` · تکمیل، پوش و با CI موفق بسته شد؛ دستور اجرا پس از تسک‌بندی صادر شد.

## مسئله و نتیجه

پیاده‌سازی مقصد محلی پیشنهادی JSON Lines پشت interface؛ مسیر خصوصی خارج Git/web-root، پوشش چند پردازه، چرخش بر اساس اندازه/زمان، سقف دیسک و نگهداری قابل تنظیم، کنترل دسترسی فایل، صف محدود و flush/shutdown. تحویل روش جست‌وجوی محلی با correlation/type/time/result و تست sink جایگزین.

## دامنه و منابع

- محیط: API، کارهای پس‌زمینه و رابط مرتبط storefront/بک‌آفیس؛ تغییر پنل فروشنده فقط در صورت ارتباط همین قابلیت.
- مجوز: [نقش‌ها](../../../product/roles.md) و scope موجود؛ دسترسی به log یا credentials از عضویت فروشگاه نتیجه نمی‌شود.
- مرجع مشترک: [قرارداد پیشنهادی لاگ](../../../architecture/contracts/operation-logging.md)، [پرداخت](../../../domains/payments/README.md) و [مرز مالی](../../../decisions/adr-001-pricing-and-payment-boundaries.md).
- خارج دامنه: نصب/اتصال Elastic، ClickHouse، سرویس ابری یا ساخت پنل جدید بدون انتخاب مالک.
- وابستگی: [#155](https://github.com/MohamadParsa/marketplace/issues/155)

## مسیر و کنترل‌های لازم

رویداد معنایی از context معتبر تولید، پیش از هر مقصد پالایش و با شناسه پیگیری و نتیجه ثبت شود؛ جزئیات نوع/سطح/فیلدها و تفکیک audit/ledger/behavior در قرارداد مشترک است.

حالت موفق، رد اعتبارسنجی/مجوز، نتیجه نامشخص/قطع سرویس، ارسال تکراری و تغییر هم‌زمان باید شاهد متناسب داشته باشند. ورودی مهم یا مالی از client معتبر فرض نشود. داده محرمانه در log، exception، خروجی عمومی یا cache ثبت نشود. رفتار شکست/نگهداریِ حل‌نشده مطابق دفتر پرسش‌ها تعیین شود، نه حدس پیاده‌ساز.

## پذیرش قابل بررسی

- **FEAT-LOG-LOCAL-SINK-001-AC01:** رویداد معتبر در فایل محلی قابل خواندن و پس از restart موجود باشد؛ مسیر پیش‌فرض خصوصی و از HTTP/استاتیک غیرقابل دریافت باشد.
- **FEAT-LOG-LOCAL-SINK-001-AC02:** چرخش، انقضا و سقف مصرف با داده ساختگی آزموده شوند؛ دو API روی یک دستگاه خط‌ها را خراب نکنند و هر رویداد مبدأ پردازه/سرویس داشته باشد.
- **FEAT-LOG-LOCAL-SINK-001-AC03:** دیسک پر، مجوز نامعتبر، صف پر و shutdown بدون نشت secret/بن‌بست و حذف خاموش داده مالی آزمایش شوند؛ سیاست شکست audit از telemetry جدا و در NEXT-Q09 تعیین شود.
- **FEAT-LOG-LOCAL-SINK-001-AC04:** با sink آزمایشی دوم، ثبت‌کننده‌های کسب‌وکار بدون تغییر کار کنند؛ schema ثابت بماند و راهنمای مشاهده/پاکسازی محدود و شاخص dropped/write-failure تحویل شود.

## روش ارزیابی و تحویل آینده

integration filesystem و concurrency/rotation/restart/failure؛ اعداد retention/queue/size پیش از اجرا مستند و قابل تنظیم باشند.

پس از انتخاب اجرای همان بخش: تست متناسب، مستندات، commit/push، CI و کامنت شواهد؛ سپس بستن همین Issue. تکمیل تسک‌بندی یا مستندات به معنی done پیاده‌سازی نیست؛ merge/deploy عمومی و اتصال واقعی مجوز جدا می‌خواهند.

## پرسش و بخش متوقف

NEXT-Q09؛ مدت/حجم نگهداری و رفتار شکست هر نوع لاگ هنوز تصمیم باز است، مقصد محلی انتخاب شده.

مرجع: [دفتر پرسش‌ها](../../../product/open-questions.md) · [برنامه و ترتیب](../README.md) · [BDD طراحی](../../../bdd/features/experiences/logging-gateway-plan.feature). BDD در این نوبت اجرا نشده و همه داده‌های تست آینده باید ساختگی و ایزوله باشند.

## شاهد اجرا

JSON Lines خصوصی، صف محدود، flush/close، قفل بین پردازه‌ای Windows/Unix، rotation و سقف مشترک فایل‌ها، نگهداری۷/۳۰روز و reader/CLI فقط‌خواندنی پیاده شدند. تست هم‌زمانی دو نویسنده، restart/صفحه‌بندی، rotation/retention/سقف، صف پر و مسیر خراب موفق‌اند. اتصال به boot برنامه و endpoint ادمین در #157/#75/#76 انجام می‌شود. [گزارش](../execution.md).
