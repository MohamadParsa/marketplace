---
id: FEAT-NEXT-016
title: لاگ فنی عملیات حساس، خطا و شکست با correlation
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [DEL-NEXT-001]
approved_by: []
implementation_status: done
priority: P1
wave: W0
execution_authorized: true
---

# CV-NEXT-16 — لاگ فنی عملیات حساس، خطا و شکست با correlation

Issue: [#77](https://github.com/MohamadParsa/marketplace/issues/77) · اساین: `eliassarband` · وضعیت اجرا: **پذیرش محلی تکمیل؛ بستن پس از CI** · اجرای غیر درگاه با دستور صریح مالک.

## مسئله و نتیجه بیزینسی

خطای یک خرید یا عملیات حساس بین مراحل قابل دنبال‌کردن باشد.

## خواسته تأییدشده مالک

برای عملیات مهم و حساس، خطا و شکست لاگ مناسب داشته باشند.

## طراحی فنی پیشنهادی و داخل دامنه

بازبینی structured logging موجود؛ request/operation/trace ID، خطای typed، latency و outcome برای auth/checkout/اعتبار/کوپن/درگاه/موجودی. masking و سطح دسترسی، خطای زیرساخت و retry؛ متفاوت از ActivityLog و Analytics، بدون جایگزینی ledger.

نوع کار: `audit-then-implementation`. برای audit/design/analysis-first، گزارش و تصمیم پیش از اصلاح اجرایی است؛ کار بزرگ پس از تحلیل به زیرتسک قابل تحویل ریز شود. این بسته دامنه کل محصول یا مجوز سرویس بیرونی را گسترش نمی‌دهد.

## وابستگی و مرز مالکیت

بدون وابستگی سخت داخل این بسته؛ قواعد/مجوزهای موجود پیش‌شرط‌اند.

- [قواعد payments](../../../domains/payments/README.md)

## پذیرش قابل بررسی

- **FEAT-NEXT-016-AC01:** یک شکست ساختگی checkout با correlation در تمام مراحل دنبال شود.
- **FEAT-NEXT-016-AC02:** OTP/رمز/token/DSN/شماره کارت کامل و بدنه خصوصی در لاگ نباشند.
- **FEAT-NEXT-016-AC03:** خطا، رد اعتبارسنجی و نتیجه نامشخص به‌درستی تفکیک شوند.
- **FEAT-NEXT-016-AC04:** مرجع مشاهده، سطح لاگ و retention با بار نمونه و مجوز بررسی شوند.

## پرسش و بخش متوقف

NEXT-Q09 برای برش محلی با پاسخ مالک حل‌شده است؛ سرویس بیرونی انتخاب نشده و درگاه‌ها معوق‌اند.

[دفتر پرسش‌های همین مرحله](../../../product/open-questions.md#next-phase-questions) · [برنامه/تفکیک مالکیت](../README.md) · [سناریوهای رفتاری مستند](../../../bdd/features/experiences/next-phase-planning.feature).

## مرز مشترک و شرط شروع

برش لاگ محلی با دستور مستقیم مالک اجرا شده است؛ وضعیت سند in-review با done اجرا متفاوت است. طرح **A — متعادل و شیک** و نشان C چریوا مرجع‌اند. تغییر پنل و API و storefront در محدوده همین قابلیت با هم دیده شود؛ قسمت نامرتبط با دلیل کنار گذاشته شود.

فرم بدون مجوز server-side کافی نیست: actor، store/user scope، version و idempotency متناسب حفظ شوند. loading/empty/error/forbidden/conflict و حفظ ورودی در خطا مشخص باشند؛ داده خصوصی وارد کش عمومی نشود و محتوای عمومی بعد از تغییر معتبر طبق سیاست کش تجدید شود. در UI کیبورد/RTL/موبایل و الگوی بارگذاری مشترک رعایت شوند؛ dropdown تازه فقط با نیاز صریح پذیرفته‌شده، مانند استان/شهر، مطرح شود.

## شواهد پایان در مرحله اجرا

ACهای همین قابلیت به تست مناسب API/دامنه/مرورگر وصل شوند؛ fixture، اتصال محلی و provider واقعی جدا گزارش شوند. تکمیل سند، build یا اسکرین‌شات به‌تنهایی اثبات عملیات مالی/مجوز نیست. پس از اجرای انتخاب‌شده: سورس و گزارش/تصاویر کامیت و پوش، بررسی‌های لازم، کامنت شواهد و سپس بستن Issue. انتشار لیارا همچنان با پارساست؛ پس از تست و CI، همین Issue با کامنت شاهد بسته می‌شود.

## تفکیک اجرای لاگ محلی پیش از درگاه — ۲۰۲۶-۱۱-۲۸

#77 مالک فنی/ردیابی و پذیرش یکپارچه باقی می‌ماند. اجرای مشترک به [#155 قرارداد/correlation/پالایش](../../logging-gateway-plan-2026-11-28/tasks/logging-contract-correlation-redaction.md)، [#156 مقصد محلی و rotation](../../logging-gateway-plan-2026-11-28/tasks/local-log-storage-and-sink-boundary.md) و [#157 نقاط حساس و تماس‌های بیرونی](../../logging-gateway-plan-2026-11-28/tasks/instrument-critical-actions-and-external-services.md) شکسته شد؛ در این والد همان کد دوباره ساخته نشود. همه اکشن‌های درگاه آینده پیش از پذیرش نیازمند پوشش لاگ‌اند. logger JSON/stdout فعلی ممیزی و تکمیل شود؛ لاگ فنی، audit #75 و رفتار #76 با مالکیت جدا از [قرارداد مشترک](../../../architecture/contracts/operation-logging.md) استفاده کنند. مقصد فعلاً محلی و تغییر به ابزار دیگر آینده است. NEXT-Q09 با پاسخ مالک برای برش محلی حل شده است. [برنامه](../../logging-gateway-plan-2026-11-28/README.md) و [گزارش اجرا](../../logging-gateway-plan-2026-11-28/execution.md) وضعیت واقعی شش تسک را جدا از درگاه معوق نگه می‌دارند.


## شواهد پایان

قرارداد/ذخیره/instrumentation در #155..157 و audit/رفتار در #75/#76 تکمیل شده‌اند؛ در این تسک دوباره ساخته نشدند. تست شکست checkout با correlation مشترک، عدم تغییر مالی و حذف داده خصوصی؛ تست مسیر عمومی از پشت symlink؛ lint/build و full Go/PostgreSQL/Redis/vet موفق‌اند. race در CI لینوکس بررسی می‌شود. [گزارش و تصاویر](../../logging-gateway-plan-2026-11-28/execution.md) · [راهنمای محلی و مرجع مشاهده](../../logging-gateway-plan-2026-11-28/local-logging-guide.md). Gateway/OAuth در #158..163 اجرا نشده است.
