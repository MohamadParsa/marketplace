---
id: DEL-SF-UX-049
title: تحویل Switch مشترک سایت مشتری
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-SF-UX-005, FEAT-SF-UX-004]
approved_by: []
implementation_status: review
---

# نتیجه #49

checkboxهای سایت به Switch مشترک با primary فعلی، track روشن/خاموش، focus و پشتیبانی disabled/error تبدیل شدند. جزء `ToggleSwitch` از input بومی checkbox با `role=switch` و label موجود استفاده می‌کند؛ input با opacity صفر در محل واقعی کنترل است و دکمه/submit جداگانه‌ای ایجاد نمی‌کند. حالت checked، پیش‌فرض، callback و شرط رضایت قبلی حفظ‌اند. کیبورد Space و اعلان حالت بومی برقرار است؛ transition در reduced-motion خاموش می‌شود.

## فهرست پوشش

| محل | وضعیت/رفتار محفوظ |
| --- | --- |
| `listing-page.tsx` — فقط کالاهای موجود | پیش‌فرض خاموش، همان in_stock در URL و پاک‌سازی/Refresh |
| `attribute-filter.tsx` — هر مقدار ویژگی | چندانتخابی، OR داخل ویژگی/AND میان ویژگی‌ها، پنجره۵گزینه و چیپ پنهان منتخب #48 |
| `commerce-pages.tsx` — تأیید مبلغ/آدرس/ارسال | ابتدا خاموش؛ ثبت سفارش تا تأیید صریح غیرفعال؛ resetهای قبلی انتخاب روش/quote محفوظ |
| `checkout/checkout-page.tsx` — پیش‌نمایش قدیمی | در entry فعلی route ندارد؛ منبع نیز از جزء مشترک استفاده می‌کند و disabled پیش از تکمیل ارسال/accepted قبلی حفظ است؛ build آن بررسی شد |

جست‌وجوی تمام TSXها فقط **یک تعریف type=checkbox در خود ToggleSwitch** باقی گذاشت. Checkbox مربعِ قابل مشاهده یا role=checkbox در صفحات فعال بررسی‌شده وجود ندارد. sort و آدرس/روش ارسال/روش پرداخت همچنان انتخاب تکی radio/کارت قبلی هستند؛ بک‌آفیس و سلر تغییر نکرده‌اند. هیچ رضایت/گزینه مالی تازه یا حالت پیش‌فرض فعال ساخته نشد. Labelهای rich متن تأیید خرید نیز همان متن قبلی‌اند.

## شواهد و محدوده اعتبار

- build/typecheck storefront موفق؛ کد Go، API مالی/هویتی یا مدل داده در #49 تغییر نکرده است. CI عمومی مخزن پیش از ادغام، Go/race/vet را نیز بررسی می‌کند.
- **۲۰کنترل Switch موفق و بدون runtime pageerror** در [کنترل‌های Switch](../../../tests/e2e/storefront-switches.cjs): پیش‌فرض خاموش، Space، رنگ primary حالت روشن، چند انتخاب هم‌زمان، URL/Refresh/پاک‌سازی، نبود role مربع checkbox و input نمایان، radio باقی‌مانده و موبایل۳۹۰.
- تأیید خرید با **fixture خصوصی فقط در شبکه مرورگر** بررسی شد: config، حساب ساختگی، سبد، آدرس، preview و quote همگی intercept شدند؛ POST خرید guard شد و صفر بار درخواست شد. Switch رضایت ابتدا خاموش/ثبت غیرفعال، با Space روشن/فعال و با خاموش‌کردن دوباره غیرفعال شد. هیچ کاربر، نشانی، رزرو، سفارش یا پرداختی از این سناریو در دیتابیس ثبت نشد.
- **۲۳کنترل رگرسیون #48 موفق** با role جدید switch، همان drag/تومان‌ریال/دامنه/بیشترکمتر/انتخاب ششم/Back و حالت‌های خالی را پوشش داد. selectorهای مشتری در اسکریپت‌های checkout به‌روز شدند؛ checkboxهای مالی بک‌آفیس در همان فایل‌ها دست‌نخورده‌اند.
- حالت disabled در پیش‌نمایش قدیمی غیرقابل route است؛ برای نمایش آن مسیر آزمایشی تازه‌ای به محصول اضافه نشد. native disabled و کلاس حالت در جزء مشترک وجود دارند؛ بررسی runtime این گزارش برای صفحات فعال و fixture خرید بالاست.

## تصاویر

[فیلتر دسکتاپ](../../../output/playwright/issue-49/filters-desktop.png) · [فیلتر موبایل](../../../output/playwright/issue-49/filters-mobile.png) · [رضایت خرید دسکتاپ](../../../output/playwright/issue-49/checkout-consent-desktop.png) · [رضایت خرید موبایل](../../../output/playwright/issue-49/checkout-consent-mobile.png).

تصاویر رگرسیون جدید در پوشه [رگرسیون](../../../output/playwright/issue-49/regression/filters-mobile.png) ذخیره شدند؛ تصاویر تحویل اصلی #48 بازنویسی نشدند. [شرح/ACها](tasks/cv-sf-ux-05-storefront-switch-controls.md) و نتایج ماشین‌خوان کنار همین گزارش‌اند.

اجرا از D:\Cherriva با Start-Cherriva.ps1؛ لینک [فیلترهای موبایل](http://localhost:5197/category/mobile). #45تا#48 پیش‌تر با کامنت بسته شدند؛ #49 پس از تحویل سورس/CI با کامنت بسته می‌شود. انتشار لیارا با پارساست.
