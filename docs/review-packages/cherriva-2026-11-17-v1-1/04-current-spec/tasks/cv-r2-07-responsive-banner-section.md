---
id: CV-R2-07
document-status: in-review
implementation-status: todo
priority: P1
wave: W2
owner: unassigned
updated: 2026-11-17
---

# CV-R2-07 — سکشن بنر با تصویر دسکتاپ و موبایل

## هدف بیزینسی

بنر روی موبایل crop نامناسب نگیرد و فرم آن فیلد نامرتبط ریل محصول نداشته باشد.

این کار اصلاح/تکمیل تسک‌های CV-CMS-01، CV-CMS-02 است؛ قابلیت مستقل تکراری یا پیاده‌سازی انجام‌شده محسوب نشود. قواعد صریح را از [مشخصات جاری](../sections-fa.md) و راه‌حل پیشنهادی را از متن زیر بخوانید.

## دامنه و راه‌حل فنی پیشنهادی

discriminated schema برای hero/banner-grid؛ desktop_media_id و mobile_media_id و alt/link مستقل هر اسلاید. preview دو اندازه؛ draft/revision و publish موجود حفظ شود. script arbitrary در link یا متن ممنوع؛ template نسبت تصویر را اعلام کند.

## داده و قرارداد پیشنهادی

Section.type و payload معتبر برای همان type؛ picture/source با breakpoint قالب؛ GET page revision و publish اتمیک.

نام endpoint و schema در این سند پیشنهاد قرارداد است؛ در kickoff با OpenAPI پروژه مقصد هماهنگ شود و بدون نسخه‌گذاری با مسیر قدیمی موازی نشود.

## دسترسی و محدوده

pages.edit و pages.publish مستقل و scope main/gold/digi؛ upload مجوز رسانه.

## خطا و حالات رابط

DESKTOP_IMAGE_REQUIRED، MOBILE_IMAGE_REQUIRED، LINK_INVALID، REVISION_CONFLICT؛ انتشار ناقص متوقف شود.

حالت loading، خالی، خطای شبکه و تلاش مجدد امن، عدم دسترسی و conflict متناسب با همان عملیات لازم‌اند. خطا نباید draft معتبر کاربر را حذف یا موفقیت ساختگی اعلام کند.

## وابستگی و مانع دقیق

- وابستگی‌ها: [CV-CMS-01](../../03-analysis-v1-0/tasks/cv-cms-01.md)، [CV-CMS-02](../../03-analysis-v1-0/tasks/cv-cms-02.md)، [CV-R2-04](cv-r2-04-category-brand-media.md).
- اندازه نمونه پیشنهادی؛ نسبت/سقف محتوا نهایی CHQ-12؛ asset واقعی مستقل.
- مسئول فردی و موعد تعیین نشده‌اند؛ اندازه‌بندی در kickoff پس از بررسی سورس مقصد. اولویت P1؛ موج W2.

## معیار پذیرش

- CV-R2-07-AC01: فرم بنر دو preview مستقل دارد؛ تصویر تغییر یک اندازه دیگری را overwrite نکند.
- CV-R2-07-AC02: عرض390 و1440 تصویر درست را نشان دهند؛ alt و لینک قابل لمس باشند.
- CV-R2-07-AC03: ویرایش draft، صفحه منتشرشده یا کش عمومی را تا publish عوض نکند.

## خروجی قابل تحویل تیم

قرارداد/کد محدود به این دامنه، migration در صورت نیاز با گزارش اثر داده، بررسی معنادار ACها و گزارش فایل‌ها/نتیجه/محدودیت. ابتدا پنل و سلر با داده آزمون؛ فلو سایت پس از گیت مرحله اول. بررسی mock و اتصال واقعی جدا گزارش شوند؛ این فایل مجوز شروع خودکار یا انتشار نیست.

[فهرست اصلاحات و ترتیب اجرا](../task-index-fa.md) · [راهنمای شروع](../../START-HERE-fa.md)
