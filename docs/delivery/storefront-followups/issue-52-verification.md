---
id: DEL-SF-UX-052
title: تحویل مشاهده بیشتر معرفی و مشخصات محصول
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-SF-UX-008, API-PRODUCT-DETAIL-001]
approved_by: []
implementation_status: review
---

# نتیجه #52

معرفی محصول (شامل مزایا/معایب) و مشخصات فنی هرکدام مستقل با جزء ExpandableContent جمع و باز می‌شوند. برای طرح A حد اجرایی۳۲۰پیکسل در دسکتاپ و۲۴۰پیکسل تا عرض۶۰۰ انتخاب شد؛ این اندازه تصمیم پیاده‌سازی قابل اصلاح است، نه تصویب کل سند. fade ظریف۲۶پیکسل و دکمه primary روشن با حداقل ارتفاع۴۴پیکسل دارند. محتوای کوتاه/خالی کنترل ندارد.

متن HTML پاک‌سازی‌شده کامل، ساختار پاراگراف‌ها و همه ردیف‌های مشخصات در DOM می‌مانند؛ هیچ برش رشته/کاهش API یا تغییر سیاست sanitizer انجام نشده است. لینک‌های خارج قسمت نمایان در حالت جمع‌شده tabindex=-1 می‌گیرند؛ با بازشدن مقدار قبلی دقیق برمی‌گردد. aria-expanded/controls و Enter/Space فعال‌اند. بستن بخش بلند آن بخش را در دید نگه می‌دارد و focus دکمه حفظ می‌شود. ResizeObserver، resize پنجره، load محتوای داخلی و loadingdone فونت اندازه را دوباره بررسی می‌کنند. با عوض‌شدن محصول حالت جمع‌شده بازنشانی می‌شود؛ بازشدن یک بخش وابسته به دیگری نیست. چاپ محدودیت ارتفاع ندارد.

## شواهد

[۱۹کنترل مرورگر](../../../tests/e2e/pdp-read-more.cjs) موفق: دو بخش مستقل،۳۲۰/۲۴۰، حفظ۳۷پاراگراف/۳۶ردیف و strong، خارج‌کردن لینک پنهان از Tab و بازگرداندن آن، Enter/Space/focus، عرض۳۹۰/۱۴۴۰، resize، نبود overflow صفحه، متای همان محصول و robots، متن کوتاه/خالی و reset پس از reload، بدون runtime pageerror. build/typecheck موفق است.

توضیح بسیار بلند و تعداد زیاد ردیف‌ها fixture خواندنی مرورگرند؛ کاتالوگ/موجودی/قیمت و داده مالی نوشته نشد. متن واقعی API و sanitizer سرور دست‌نخورده‌اند؛ این آزمایش، اجرای دوباره تست امنیتی sanitizer یا ادعای SSR کامل بدنه نیست. متای SEO و robots پیشین حفظ می‌شوند.

[جمع‌شده دسکتاپ](../../../output/playwright/issue-52/collapsed-desktop.png) · [جمع‌شده موبایل](../../../output/playwright/issue-52/collapsed-mobile.png) · [بازشده موبایل](../../../output/playwright/issue-52/expanded-mobile.png) · [شرح و AC](tasks/cv-sf-ux-08-pdp-content-read-more.md).

کد در D:\Cherriva است؛ انتشار لیارا با پارساست. نتیجه ادغام/CI جدا ثبت می‌شود.
