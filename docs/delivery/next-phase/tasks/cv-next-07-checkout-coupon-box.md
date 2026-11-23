---
id: FEAT-NEXT-007
title: باکس مستقل کد تخفیف در مرحله پرداخت خرید
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [DEL-NEXT-001]
approved_by: []
implementation_status: todo
priority: P1
wave: W3
execution_authorized: false
---

# CV-NEXT-07 — باکس مستقل کد تخفیف در مرحله پرداخت خرید

Issue: [#68](https://github.com/MohamadParsa/marketplace/issues/68) · اساین: `eliassarband` · وضعیت GitHub: **open** · شروع اجرا: انتخاب نشده.

## مسئله و نتیجه بیزینسی

مشتری کوپن را اعمال/حذف کند و تفکیک مبلغ جدید را ببیند.

## خواسته تأییدشده مالک

ورود کد در مرحله پرداخت، باکس مجزا و به‌روزشدن مبلغ نهایی.

## طراحی فنی پیشنهادی و داخل دامنه

پیشنهاد: کد/usage در quote سرور؛ apply/remove فقط پس از محاسبه معتبر، consent مبلغ با تغییر جمع بازنشانی شود. درخواست دیررس جمع تازه را بازنویسی نکند؛ مصرف نهایی همراه PlaceOrder قطعی شود.

نوع کار: `implementation`. برای audit/design/analysis-first، گزارش و تصمیم پیش از اصلاح اجرایی است؛ کار بزرگ پس از تحلیل به زیرتسک قابل تحویل ریز شود. این بسته دامنه کل محصول یا مجوز سرویس بیرونی را گسترش نمی‌دهد.

## وابستگی و مرز مالکیت

[CV-NEXT-05](cv-next-05-checkout-amounts-usage-contract.md)، [CV-NEXT-06](cv-next-06-coupon-management.md)، [CV-NEXT-09](cv-next-09-atomic-place-order.md)، [CV-NEXT-12](cv-next-12-sales-method-markup.md)

- [purchase-experience](../../../architecture/contracts/purchase-experience.md)
- [checkout-wallet-launch](../../../architecture/contracts/checkout-wallet-launch.md)

## پذیرش قابل بررسی

- **FEAT-NEXT-007-AC01:** اعمال/حذف کوپن، سهم کالا/ارسال و مانده قابل پرداخت را دقیق به‌روز کند.
- **FEAT-NEXT-007-AC02:** کد نامعتبر/منقضی/غیرمجاز پیام روشن و ورودی/مرحله خرید حفظ شوند.
- **FEAT-NEXT-007-AC03:** پاسخ قدیمی و دوبار کلیک تخفیف یا مصرف مضاعف نسازند.
- **FEAT-NEXT-007-AC04:** کار با کیف پول و روش فروش در مثال‌های قرارداد مبلغ و snapshot سفارش بررسی شود.

## پرسش و بخش متوقف

شروع اجرا وابسته به قرارداد مصوب usage و PlaceOrder؛ ابهام CHQ-04 را کلاینت حل نکند.

[دفتر پرسش‌های همین مرحله](../../../product/open-questions.md#next-phase-questions) · [برنامه/تفکیک مالکیت](../README.md) · [سناریوهای رفتاری مستند](../../../bdd/features/experiences/next-phase-planning.feature).

## مرز مشترک و شرط شروع

این فایل **شرح کار آینده** است: وضعیت سند in-review، پیاده‌سازی todo و مجوز اجرای کد هنوز صادر نشده است. انتخاب زمان شروع با مالک است؛ اولویت/موج پیشنهادند، نه برنامه زمانی قطعی. طرح **A — متعادل و شیک** و نشان C چریوا مرجع‌اند. تغییر پنل و API و storefront در محدوده همین قابلیت با هم دیده شود؛ قسمت نامرتبط با دلیل کنار گذاشته شود.

فرم بدون مجوز server-side کافی نیست: actor، store/user scope، version و idempotency متناسب حفظ شوند. loading/empty/error/forbidden/conflict و حفظ ورودی در خطا مشخص باشند؛ داده خصوصی وارد کش عمومی نشود و محتوای عمومی بعد از تغییر معتبر طبق سیاست کش تجدید شود. در UI کیبورد/RTL/موبایل و الگوی بارگذاری مشترک رعایت شوند؛ dropdown تازه فقط با نیاز صریح پذیرفته‌شده، مانند استان/شهر، مطرح شود.

## شواهد پایان در مرحله اجرا

ACهای همین قابلیت به تست مناسب API/دامنه/مرورگر وصل شوند؛ fixture، اتصال محلی و provider واقعی جدا گزارش شوند. تکمیل سند، build یا اسکرین‌شات به‌تنهایی اثبات عملیات مالی/مجوز نیست. پس از اجرای انتخاب‌شده: سورس و گزارش/تصاویر کامیت و پوش، بررسی‌های لازم، کامنت شواهد و سپس بستن Issue. انتشار لیارا همچنان با پارساست؛ در این مرحله هیچ Issue پیاده‌سازی بسته نمی‌شود.
