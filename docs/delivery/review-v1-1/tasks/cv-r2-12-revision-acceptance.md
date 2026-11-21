---
id: CV-R2-12
title: پذیرش اصلاحات و ردیابی با دامنه فعال
status: in-review
owner: unassigned
updated: 2026-11-21
related: [NFR-INDEX-001]
approved_by: []
implementation_status: todo
priority: P0
wave: W3
---

# CV-R2-12 — پذیرش اصلاحات و ردیابی با دامنه فعال

Issue: [#12](https://github.com/MohamadParsa/marketplace/issues/12) · پاسخ‌های مالک: [R2-Q01](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q02](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q03](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q04](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q05](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q06](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [Q-027](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-027)، [Q-028](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-028) · [ADR-005](../../../decisions/adr-005-review-v1-1-product-responses.md)

این مشخصات بازنویسی پس از #13 است و هنوز in-review است. اولویت و موج ترتیب پیشنهادی وابستگی‌اند؛ مسئول/موعد تخصیص نیافته‌اند. اجرای کار todo است.

## هدف بیزینسی و نتیجه

هر قابلیت فعال با شاهد واقعی همان سطح بررسی تحویل شود؛ کار معوق به‌جای done ثبت نشود.

## داخل دامنه

- ماتریس پاسخ→سند→Issue→AC→شاهد برای دامنه فعال؛ #7/#8/#9 از الزام پذیرش شروع کنارند و deferred باقی می‌مانند.
- پذیرش پنل و سلر کامل با داده آزمون مقدم بر سایت؛ جهت A برای پنل/سلر/سایت مرجع است.
- آزمون API دسترسی مستقیم و scoped، auth و اثر تاریخی قیمت/وقفه/تقویم؛ ظرفیت هم‌زمان و اعلان طبق سیاست منتخب.

## راه‌حل و داده پیشنهادی

گزارش آزمون با fixture بی‌اطلاعات شخصی و شواهد قابل تکرار؛ کنترل IDOR و تزریق scope، session و log محرمانه؛ قالب گزارش بر اساس ریسک همان قابلیت.

## دسترسی، خطا و حالات

مجوز هر query و عملیات در سرور و همراه scope انتساب کنترل شود؛ خطای اعتبارسنجی کنار فیلد، loading/empty/error/forbidden و conflict متناسب با عملیات دیده شوند. draft معتبر از بین نرود و retry اثر تکراری ایجاد نکند. برای کار معوق این‌ها معیار فاز فعال‌شده‌اند؛ وجود فرم نمونه شاهد اجرا نیست.

## وابستگی

- [CV-R2-01 · #1](cv-r2-01-single-user.md)
- [CV-R2-02 · #2](cv-r2-02-mobile-first-login.md)
- [CV-R2-03 · #3](cv-r2-03-profile-password-recovery.md)
- [CV-R2-04 · #4](cv-r2-04-category-brand-media.md)
- [CV-R2-05 · #5](cv-r2-05-shared-attributes.md)
- [CV-R2-06 · #6](cv-r2-06-product-workspace.md)
- [CV-R2-10 · #10](cv-r2-10-seller-lead-snapshot.md)
- [CV-R2-11 · #11](cv-r2-11-storefront-alignment.md)
- [CV-R2-13 · #15](cv-r2-13-general-settings.md)
- [CV-R2-14 · #16](cv-r2-14-shipment-slots.md)
- [CV-R2-15 · #17](cv-r2-15-stock-alert.md)

## بخش مسدود، خارج محدوده یا شرط شروع

فقط مسیر فعال و دارای سیاست نهایی پذیرفته شود؛ بخش باز بدون شاهد و تصمیم done نشود. اتصال واقعی پذیرش مستقل می‌خواهد؛ بستن #14 اداری شاهد تکمیل این کار نیست.

اصل پنل/سلر کامل با داده آزمون پیش از سایت حفظ شود. مخزن/پروژه، provider، مقررات و تصمیم باز از fixture یا رابط استنباط نشود؛ اتصال واقعی گیت مستقل دارد.

## معیار پذیرش

- CV-R2-12-AC01: هر AC فعال به گزارش/اجرا و نتیجه مثبت/منفی متناسب لینک شود؛ مرور screenshot جای تست backend/امنیت/مالی ننشیند.
- CV-R2-12-AC02: Issueهای معوق باز/معوق و خارج شرط شروع باشند؛ هیچ done با fixture یا build تنها ثبت نشود.
- CV-R2-12-AC03: لینک‌ها/شناسه‌ها/dependency graph و قراردادهای فعلی سازگار و سیاست‌های باقیمانده برای هر بخش blocked مشخص باشند.
- CV-R2-12-AC04: گزارش mock، API/integration، browser و sandbox/provider واقعی جدا و نواقص قابل پیگیری باشند.

## شاهد پایان پیاده‌سازی

ماتریس تکمیل‌شده AC/مدرک، گزارش API و مرورگر و تصمیم‌های باز؛ هیچ شاهد اجرای محصول در مرحله برنامه‌ریزی فعلی جعل نمی‌شود.

گزارش تحویل: شناسه کار، فایل/قرارداد، رفتار قبل/بعد، داده آزمون و نتیجه هر AC، مجوز/scope، موارد باز و لینک شاهد. mock، integration/API، browser و provider مستقل گزارش شوند؛ هیچ نتیجه‌ای در مرحله تدوین این سند ثبت نشده است.

## ردیابی و سابقه

اصلاح تسک‌های پایه: CV-QA-01، CV-QA-04، CV-QA-06.

[فهرست فعال](../README.md) · [پرسش‌های جاری](../../../product/open-questions.md) · [بسته تاریخی](../../../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/task-index-fa.md). شناسه CV، شماره Issue و شناسه Q سه فضای متفاوت‌اند؛ مثلاً CV-R2-14، خود Issue فرایندی #14 نیست.


## وابستگی تکمیلی محتوای کاتالوگ — ۲۰۲۶-۱۱-۲۱

[CV-R2-16 / #24](cv-r2-16-category-content-seo.md) و [CV-R2-17 / #25](cv-r2-17-product-content-seo.md) به وابستگی این کار اضافه شدند. [مرجع مشترک](../../../domains/catalog/content-and-seo.md) و R2-Q07 تا R2-Q09 مبنای اجرا هستند. #11 متادیتای HTML اختصاصی PLP/PDP، حالت robots، نمایش محتوا/گالری و تنظیم مستقل نظر/Q&A را با cache عمومی امن مصرف و بررسی می‌کند؛ #12 شواهد ACهای هر دو تسک و مصرف نهایی را در ماتریس پذیرش می‌آورد. وجود DTO پنل به معنی تکمیل سایت نیست.
