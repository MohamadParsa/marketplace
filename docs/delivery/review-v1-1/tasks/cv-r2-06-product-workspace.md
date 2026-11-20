---
id: CV-R2-06
title: اکشن سریع تنوع، قیمت و موجودی در محصول
status: in-review
owner: unassigned
updated: 2026-11-20
related: [DOM-CATALOG-001, FEAT-BO-001]
approved_by: []
implementation_status: todo
priority: P0
wave: W1
---

# CV-R2-06 — اکشن سریع تنوع، قیمت و موجودی در محصول

Issue: [#6](https://github.com/MohamadParsa/marketplace/issues/6) · پاسخ‌های مالک: [R2-Q02](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری)، [R2-Q03](https://github.com/MohamadParsa/marketplace/blob/main/docs/product/open-questions.md#پرسشهای-اصلاحات-جاری) · [ADR-005](../../../decisions/adr-005-review-v1-1-product-responses.md)

این مشخصات بازنویسی پس از #13 است و هنوز in-review است. اولویت و موج ترتیب پیشنهادی وابستگی‌اند؛ مسئول/موعد تخصیص نیافته‌اند. اجرای کار todo است.

## هدف بیزینسی و نتیجه

از گرید و جزئیات محصول به فرم واحد ترکیب و عرضه فروشنده دسترسی سریع فراهم شود.

## داخل دامنه

- Product محتوا، Variant ترکیب معتبر محورهای انتخاب و Offer قیمت/موجودی هر فروشنده است.
- اکشن گرید و tab جزئیات همان فرم/شناسه‌ها و مجوز را مصرف کنند؛ فروشنده فقط عرضه‌های scope خودش را ویرایش کند.
- انتخاب رنگ/گارانتی در PDP به همان ترکیب و عرضه برسد؛ تعداد محور از مثال hardcode نشود.

## راه‌حل و داده پیشنهادی

use case و component مشترک؛ Stock با reserved و کنترل version؛ مجوز products/variants/offers/stock مستقل. خطای هر ردیف، conflict و حفظ draft.

## دسترسی، خطا و حالات

مجوز هر query و عملیات در سرور و همراه scope انتساب کنترل شود؛ خطای اعتبارسنجی کنار فیلد، loading/empty/error/forbidden و conflict متناسب با عملیات دیده شوند. draft معتبر از بین نرود و retry اثر تکراری ایجاد نکند. برای کار معوق این‌ها معیار فاز فعال‌شده‌اند؛ وجود فرم نمونه شاهد اجرا نیست.

## وابستگی

- [CV-R2-05 · #5](cv-r2-05-shared-attributes.md)
- [CV-CAT-04](../../../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/tasks/cv-cat-04.md)
- [CV-CAT-06](../../../review-packages/cherriva-2026-11-17-v1-1/03-analysis-v1-0/tasks/cv-cat-06.md)

## بخش مسدود، خارج محدوده یا شرط شروع

تغییر آینده محور استفاده‌شده تابع R2-Q03 است؛ قیمت فرمولی واقعی طلا تا CHQ-03 فقط تنظیم draft/قرارداد روشن دارد. این کار مجوز حدس فرمول طلا نیست.

اصل پنل/سلر کامل با داده آزمون پیش از سایت حفظ شود. مخزن/پروژه، provider، مقررات و تصمیم باز از fixture یا رابط استنباط نشود؛ اتصال واقعی گیت مستقل دارد.

## معیار پذیرش

- CV-R2-06-AC01: ورودی گرید و جزئیات همان Variant/Offer را باز کنند و قیمت روی Offer ثبت شود.
- CV-R2-06-AC02: موجودی کمتر از رزرو و ویرایش عرضه فروشگاه دیگر با API مستقیم رد شود.
- CV-R2-06-AC03: ترکیب رنگ/گارانتی درست resolve شود؛ وقفه عرضه خواندنی از فروشنده باشد و محتوای عمومی ناخواسته تغییر نکند.

## شاهد پایان پیاده‌سازی

API و مرورگر از هر دو entry point؛ حداقل دو فروشگاه، دو ترکیب و stock/reserved؛ گزارش مثبت/منفی تغییر قیمت/موجودی.

گزارش تحویل: شناسه کار، فایل/قرارداد، رفتار قبل/بعد، داده آزمون و نتیجه هر AC، مجوز/scope، موارد باز و لینک شاهد. mock، integration/API، browser و provider مستقل گزارش شوند؛ هیچ نتیجه‌ای در مرحله تدوین این سند ثبت نشده است.

## ردیابی و سابقه

اصلاح تسک‌های پایه: CV-CAT-04، CV-CAT-05، CV-CAT-06.

[فهرست فعال](../README.md) · [پرسش‌های جاری](../../../product/open-questions.md) · [بسته تاریخی](../../../review-packages/cherriva-2026-11-17-v1-1/04-current-spec/task-index-fa.md). شناسه CV، شماره Issue و شناسه Q سه فضای متفاوت‌اند؛ مثلاً CV-R2-14، خود Issue فرایندی #14 نیست.
