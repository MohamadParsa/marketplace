---
id: VERIFY-PDP-REVIEW-RATING-20261126
title: تحویل ستاره‌های نظرات و میانگین بالای PDP — #119
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-NEXT-031, API-PRODUCT-REVIEW-RATING-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# نتیجه #119

ستاره‌ها با SVG واحد، پر طلایی و خالی خاکستری روشن و پرشدن جزئی میانگین بازطراحی شدند؛ کادرهای مجزای ورودی حذف شد و انتخاب hover/کیبورد باقی است. ردیف ظریف امتیاز، تعداد دیدگاه و لینک به نظرات زیر نام و کنار کد محصول، پیش از تنوع قرار دارد. جای‌گذاری از نمونه‌های فروشگاهی در [قرارداد](../../architecture/contracts/product-review-rating.md#مبنای-طراحی-و-حدود) الهام گرفته است.

میانگین سمت سرور از **تمام نظرات approved همان محصول**، مستقل از صفحه‌بندی محاسبه می‌شود؛ count صفر با average=null و متن بدون امتیاز است. API موجود بدون کش تازه خوانده می‌شود و focus صفحه نتیجه جدید را در هدر و نظرات با یک درخواست یکسان نشان می‌دهد. ثبت pending در میانگین اثری ندارد. [سورس](https://github.com/MohamadParsa/marketplace/commit/5b630e2366c333895b375e4a288d09384e2a3378) · [تسک](tasks/cv-next-31-pdp-review-stars-and-average.md) · [شواهد](issue-119-checks.json).

| پذیرش | شاهد |
| --- | --- |
| AC01 نمایش و ورودی | پنج SVG تک‌ردیف، طلایی/خاکستری، fraction، radio۱..۵ با click و keyboard، بدون باکس مجزا، لینک هدر پیش از تنوع |
| AC02 صحت میانگین | صفر/null،۵و۳→۴،۴٫۳ و۲٫۳، عدم احتساب pending/rejected/محصول دیگر/امتیاز نامعتبر؛۷رأی۴٫۷ در برابر صفحه۵تایی تمام۵ |
| AC03 تازگی/حفاظت | تأیید مجاز و تکرار امن به۸رأی۴٫۸ می‌رسد؛ focus هدر و نظرات را تازه می‌کند؛ خطا/retry، عدم Snapshot خصوصی PG و عدم تغییر داده با خواندن |
| AC04 شواهد | تست memory/PostgreSQL، موبایل۳۹۰px و دسکتاپ،۲۱کنترل مرورگر، build/vet، restart و حفظ۲۹بخش داده |

## بررسی‌ها

- `go test -json ./...`: **۳۹۳تست/زیرتست موفق،۴skip اختیاری،صفرشکست** با PostgreSQL schema جدا و Redis تست. آزمون قابلیت: `apps/api/internal/bootstrap/review_rating_test.go`؛ مقایسه rounding، pagination، visibility و restart.
- `go vet ./...` و build storefront/backoffice موفق؛ هشدار شناخته‌شده chunk بزرگ بک‌آفیس باقی است. وابستگی تازه نداریم. race محلی اجرا نشده؛ نتیجه CI در رسید بعدی ثبت می‌شود.
- `tests/e2e/review-rating.cjs`: **۲۱کنترل موفق**، schema جدا، صفر runtime pageerror. پاسخ503 عمداً برای مسیر خطا ساخته شد؛ ادعای نبود خطای شبکه نداریم. محصول، نام و تصویر نمونه screenshot ساختگی‌اند.
- نسخه اصلی در `D:\Cherriva` و پورت‌های5197/8098 به‌روز است. HTTP خانه و endpoint نظرات200؛ پس از restart مقایسه **۲۹بخش داده** شامل نظرات، کاربر/رمز/نقش، کاتالوگ/رسانه، سفارش/کیف پول/دفتر/بانک/پرداخت/ارسال با baseline نهایی برابر است. baseline نهایی پیش از restart برای حفظ آخرین ویرایش‌های صاحب محیط گرفته شد؛ داده اصلی پاک یا جایگزین نشد.

## تصاویر نسخه نهایی

[امتیاز زیر نام محصول](../../../output/playwright/issue-119/pdp-header-rating.png) · [نظرات دسکتاپ](../../../output/playwright/issue-119/reviews-desktop.png) · [انتخاب امتیاز](../../../output/playwright/issue-119/rating-input.png) · [هدر موبایل](../../../output/playwright/issue-119/pdp-header-mobile.png) · [نظرات موبایل](../../../output/playwright/issue-119/reviews-mobile.png).

## حدود

سیاست ثبت/مدیریت نظرات #89 حفظ شده است؛ API جدید پس‌گرفتن تأیید/حذف/ویرایش امتیاز اضافه نشده. تست تغییر eligibility ذخیره‌شده، تازگی محاسبه را شبیه‌سازی می‌کند. aggregate ماندگار، کش مضاعف و migration نداریم؛ structured data SEO در #74 باقی است. main/Liara ادغام یا منتشر نشده‌اند؛ سند in-review است.
