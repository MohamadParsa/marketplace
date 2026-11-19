---
id: CV-R2-11
document-status: in-review
implementation-status: todo
priority: P0
wave: W3
owner: unassigned
updated: 2026-11-17
---

# CV-R2-11 — همگام‌سازی سایت با قراردادهای اصلاح‌شده

## هدف بیزینسی

کاربر سایت همان تنظیمات و قواعد پنل را با زبان ساده ببیند؛ طرح جدید رفتار خرید قبلی را حفظ کند.

این کار اصلاح/تکمیل تسک‌های CV-WEB-02، CV-WEB-03، CV-WEB-04، CV-WEB-05، CV-WEB-06، CV-WEB-07، CV-WEB-08 است؛ قابلیت مستقل تکراری یا پیاده‌سازی انجام‌شده محسوب نشود. قواعد صریح را از [مشخصات جاری](../storefront-impact-fa.md) و راه‌حل پیشنهادی را از متن زیر بخوانید.

## دامنه و راه‌حل فنی پیشنهادی

مصرف قرارداد مشترک auth/profile، آیکن fallback، binding کارت/چیپ، انتخاب offer و lead فروشنده. سبد با offer_id و variant_id، قیمت/stock server-authoritative؛ banner/rail published. سه جهت بصری فقط theme/layout هستند؛ منطق مشترک باشد.

## داده و قرارداد پیشنهادی

DTOهای عمومی بدون اطلاعات خصوصی؛ افزودن به سبد idempotent طبق قرارداد قبلی؛ هیچ محاسبه قطعی قیمت در browser.

نام endpoint و schema در این سند پیشنهاد قرارداد است؛ در kickoff با OpenAPI پروژه مقصد هماهنگ شود و بدون نسخه‌گذاری با مسیر قدیمی موازی نشود.

## دسترسی و محدوده

public read scoped؛ account/cart owner-only؛ سرور بررسی session/store و انتخاب offer.

## خطا و حالات رابط

NO_MATCHING_OFFER، OFFER_UNAVAILABLE، QUOTE_CHANGED، SESSION_EXPIRED؛ متن واضح و اقدام اصلاح.

حالت loading، خالی، خطای شبکه و تلاش مجدد امن، عدم دسترسی و conflict متناسب با همان عملیات لازم‌اند. خطا نباید draft معتبر کاربر را حذف یا موفقیت ساختگی اعلام کند.

## وابستگی و مانع دقیق

- وابستگی‌ها: [CV-R2-03](cv-r2-03-profile-email-password.md)، [CV-R2-05](cv-r2-05-shared-attribute-bindings.md)، [CV-R2-06](cv-r2-06-product-workspace.md)، [CV-R2-08](cv-r2-08-manual-product-rail.md)، [CV-R2-09](cv-r2-09-sticky-banner-section.md)، [CV-R2-10](cv-r2-10-seller-supply-lead.md).
- بعد از پذیرش پنل/سلر آزمون آغاز شود؛ host/session CHQ-08 و قیمت طلا CHQ-03 فقط مسیر وابسته را می‌بندند.
- مسئول فردی و موعد تعیین نشده‌اند؛ اندازه‌بندی در kickoff پس از بررسی سورس مقصد. اولویت P0؛ موج W3.

## معیار پذیرش

- CV-R2-11-AC01: گارانتی۲۴ پس از انتخاب در قیمت، لیست seller، سبد و checkout یکسان دیده شود.
- CV-R2-11-AC02: کارت بدون انتخاب تنوع به PDP مستقل برود؛ CTA خرید همان استایل مرجع حفظ شود.
- CV-R2-11-AC03: درخت PLP و برگ بدون زیرشاخه، سه جهت بصری و موبایل مرور شوند؛ قوانین طلا حفظ شوند.

## خروجی قابل تحویل تیم

قرارداد/کد محدود به این دامنه، migration در صورت نیاز با گزارش اثر داده، بررسی معنادار ACها و گزارش فایل‌ها/نتیجه/محدودیت. ابتدا پنل و سلر با داده آزمون؛ فلو سایت پس از گیت مرحله اول. بررسی mock و اتصال واقعی جدا گزارش شوند؛ این فایل مجوز شروع خودکار یا انتشار نیست.

[فهرست اصلاحات و ترتیب اجرا](../task-index-fa.md) · [راهنمای شروع](../../START-HERE-fa.md)
