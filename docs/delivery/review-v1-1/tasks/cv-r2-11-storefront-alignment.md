---
id: CV-R2-11
title: همگام‌سازی سایت اصلی با تصمیم‌های نسخه۱.۱
status: in-review
owner: unassigned
updated: 2026-11-21
related: [FEAT-SF-001, DOM-CHECKOUT-001]
approved_by: []
implementation_status: todo
priority: P0
wave: W3
---

# CV-R2-11 — همگام‌سازی سایت اصلی با تصمیم‌های نسخه۱.۱

Issue: [#11](https://github.com/MohamadParsa/marketplace/issues/11) · پاسخ‌های مالک: [R2-Q01](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q03](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q04](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q05](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q06](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [Q-027](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-027)، [Q-028](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#q-028) · [ADR-005](../../../decisions/adr-005-review-v1-1-product-responses.md)

این مشخصات بازنویسی پس از #13 است و هنوز in-review است. اولویت و موج ترتیب پیشنهادی وابستگی‌اند؛ مسئول/موعد تخصیص نیافته‌اند. اجرای کار todo است.

## هدف بیزینسی و نتیجه

سایت اصلی با قالب A و قراردادهای مشترک، خرید سازگار و رفتار ناموجود روشن داشته باشد.

## داخل دامنه

- فقط cherriva.ir برای همه دسته‌ها؛ خانه با سکشن‌های ثابت و داده معتبر، بدون dependency به #7/#8/#9.
- خانه ناموجود را نشان ندهد؛ PLP ناموجود را با برچسب نگه دارد؛ تنوع ناموجود در PDP اطلاعات ولی بدون قیمت/افزودن و CTA خبرم‌کن داشته باشد.
- کارت خانه/PLP مستقیم به PDP مستقل برود؛ خرید مستقیم از کارت فعال نیست. شرط popup سبد فقط برای قابلیت آینده است.
- ورود موبایل/رمز و ویژگی‌های کارت از تنظیم عمومی/دسته مصرف شوند؛ offer/variant منتخب در PDP، cart و checkout یکسان بمانند.
- ارسال از #10 و CV-R2-14، درخواست اطلاع از CV-R2-15؛ ادامه خرید در همان سایت حفظ شود.

## راه‌حل و داده پیشنهادی

قراردادهای عمومی کاتالوگ/quote authoritative؛ cart با offer/variant/quantity؛ خطاهای تغییر quote، ناموجودی، نشست، ظرفیت و اعلان قابل اصلاح؛ داده خصوصی cache نشود.

## دسترسی، خطا و حالات

مجوز هر query و عملیات در سرور و همراه scope انتساب کنترل شود؛ خطای اعتبارسنجی کنار فیلد، loading/empty/error/forbidden و conflict متناسب با عملیات دیده شوند. draft معتبر از بین نرود و retry اثر تکراری ایجاد نکند. برای کار معوق این‌ها معیار فاز فعال‌شده‌اند؛ وجود فرم نمونه شاهد اجرا نیست.

## وابستگی

- [CV-R2-03 · #3](cv-r2-03-profile-password-recovery.md)
- [CV-R2-04 · #4](cv-r2-04-category-brand-media.md)
- [CV-R2-05 · #5](cv-r2-05-shared-attributes.md)
- [CV-R2-06 · #6](cv-r2-06-product-workspace.md)
- [CV-R2-10 · #10](cv-r2-10-seller-lead-snapshot.md)
- [CV-R2-13 · #15](cv-r2-13-general-settings.md)
- [CV-R2-14 · #16](cv-r2-14-shipment-slots.md)
- [CV-R2-15 · #17](cv-r2-15-stock-alert.md)

## بخش مسدود، خارج محدوده یا شرط شروع

شروع سایت پس از پذیرش تمام دامنه پنل/سلر آزمون است. خاموش‌بودن خود PDP، خرید مستقیم صفحه فروشگاه، مقصد داخلی بازگشت، ظرفیت و اعلان در بخش وابسته منتظر سیاست‌اند. زیرسایت/نشست بین‌میزبانی و CMS معوق‌اند، نه مانع بخش روشن.

اصل پنل/سلر کامل با داده آزمون پیش از سایت حفظ شود. مخزن/پروژه، provider، مقررات و تصمیم باز از fixture یا رابط استنباط نشود؛ اتصال واقعی گیت مستقل دارد.

## معیار پذیرش

- CV-R2-11-AC01: طرح A در۳۹۰/۱۴۴۰ و کارت→PDP مستقل مرور شود؛ modal محصول و خرید مستقیم کارت وارد مسیر نشوند.
- CV-R2-11-AC02: ناموجودی یک تنوع، تنوع قابل خرید دیگر را از فروش خارج نکند؛ قیمت/CTA حالت انتخاب‌شده درست باشد.
- CV-R2-11-AC03: همان offer/variant و quote سرور در سبد/خرید دیده و تغییر قیمت/stock با اقدام روشن بازبینی شود.
- CV-R2-11-AC04: خانه ثابت به CMS معوق وابسته نباشد و فقط موجود را نشان دهد؛ PLP با سقف/ترتیب کارت از تنظیم مصرف کند.
- CV-R2-11-AC05: مسیر انتخاب اسلات دارای ظرفیت و ثبت خبرم‌کن با backend آزمون و قرارداد تصمیم‌گرفته تکمیل شود.

## شاهد پایان پیاده‌سازی

مرورگر و integration سایت اصلی با دو فروشنده، دو تنوع و یک ناموجود؛ مسیر موفق/خطا/برگشت، تقویم و اسلات و خبرم‌کن طبق قرارداد؛ mock و provider جدا.

گزارش تحویل: شناسه کار، فایل/قرارداد، رفتار قبل/بعد، داده آزمون و نتیجه هر AC، مجوز/scope، موارد باز و لینک شاهد. mock، integration/API، browser و provider مستقل گزارش شوند؛ هیچ نتیجه‌ای در مرحله تدوین این سند ثبت نشده است.

## ردیابی و سابقه

اصلاح تسک‌های پایه: CV-WEB-02، CV-WEB-03، CV-WEB-04، CV-WEB-05، CV-WEB-06، CV-WEB-07، CV-WEB-08.

[فهرست فعال](../README.md) · [پرسش‌های جاری](../../../product/open-questions.md) · [بسته تاریخی](../../../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/task-index-fa.md). شناسه CV، شماره Issue و شناسه Q سه فضای متفاوت‌اند؛ مثلاً CV-R2-14، خود Issue فرایندی #14 نیست.


## وابستگی تکمیلی محتوای کاتالوگ — ۲۰۲۶-۱۱-۲۱

[CV-R2-16 / #24](cv-r2-16-category-content-seo.md) و [CV-R2-17 / #25](cv-r2-17-product-content-seo.md) به وابستگی این کار اضافه شدند. [مرجع مشترک](../../../domains/catalog/content-and-seo.md) و R2-Q07 تا R2-Q09 مبنای اجرا هستند. #11 متادیتای HTML اختصاصی PLP/PDP، حالت robots، نمایش محتوا/گالری و تنظیم مستقل نظر/Q&A را با cache عمومی امن مصرف و بررسی می‌کند؛ #12 شواهد ACهای هر دو تسک و مصرف نهایی را در ماتریس پذیرش می‌آورد. وجود DTO پنل به معنی تکمیل سایت نیست.
