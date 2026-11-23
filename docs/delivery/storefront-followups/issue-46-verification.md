---
id: DEL-SF-UX-046
title: تحویل منوی درختی دسته‌ها
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-SF-UX-002, API-PLP-SEARCH-001, FEAT-CATALOG-PUBLIC-URL-001]
approved_by: []
implementation_status: review
---

# نتیجه #46

دکمه دسته‌بندی کالاها منوی بازشونده واقعی باز می‌کند؛ ادامه همان نوار خالی است. در دسکتاپ ریشه‌ها در ستون اصلی و زیرشاخه‌ها در پنل کنارشان‌اند؛ هر عمق بعدی با disclosure باز می‌شود. لینک خود والد جدا از دکمه زیرشاخه است. در موبایل، ورود به شاخه و بازگشت به ریشه، sheet قابل پیمایش و قفل scroll/focus دارد. منو به hover وابسته نیست و dropdown فرم یا context menu سیستم‌عامل ساخته نشد. طرح A، پالت و لوگو حفظ‌اند؛ آیکن‌های حساب/سبد در #47 هستند.

## داده و کنترل رفتار

`GET /storefront/navigation` از projection مشترک دسته‌ها، فقط DTO ناوبری `id/name/url/parent_id/icon_url` می‌دهد؛ خواندن محصولات/اقلام مالی لازم نیست. parent_id سلسله‌مراتب را می‌سازد، URL از helper canonical #42 می‌آید و انتشار تمام اجداد کنترل می‌شود. داده دقیق موجودی/کاربر/نشانی و audit در پاسخ نیست. query ناشناخته400است؛ ETag پس از داده جاری بررسی می‌شود.

منوی اصلی همه دسته‌های منتشر را نشان می‌دهد؛ موجودی شرط تازه‌ای برای آن نیست. در داده محلی۴۱ریشه و ساختار واقعی سه‌سطحی بررسی شد. **دسته‌های زیر بنر همان۲۸ریشه دارای کالای قابل خرید و نمای جمع‌وجور۸تایی باقی‌اند**؛ این دو کاربرد با هم مخلوط نشدند. PLP والد همچنان فقط محصول مستقیم خود را نشان می‌دهد.

پاسخ navigation با namespace جاری کش، dependency دسته و TTL پنج‌دقیقه cache می‌شود؛ startup warmup اکنون navigation را هم می‌خواند. تغییر قیمت محصول نامرتبط کل منو را دوباره نمی‌سازد و تغییر دسته/انتشار با نسخه shared در درخواست بعد اعمال می‌شود. مرورگر روی mount هدر و retry پاسخ جاری می‌خواهد؛ cache حافظه‌ای مستقل و طولانی در client اضافه نشد.

هدر مشترک خانه، PLP، PDP، سبد، حساب و wizard خرید یک جزء را مصرف می‌کند. Escape و کلیک بیرون منو را می‌بندند؛ Escape focus را به trigger برمی‌گرداند. Tab در sheet موبایل داخل منو می‌ماند و ورود شاخه focus را به کنترل برگشت می‌برد؛ پس از بستن، scroll قبلی بازمی‌گردد. لینک دسته بدون فرزند مستقیم به PLP می‌رود. loading، خطای شبکه/retry و دسته خالی پیام مستقل دارند.

## شواهد

- Go۲۱۳تست/زیرتست موفق؛ یک roundtrip اختیاری۳GB اجرا نشد. تست جدید `TestPublicCategoryNavigationAncestorsAndCanonical` درخت سه‌سطحی، slug، انتشار جد، ETag و query نامعتبر را پوشش می‌دهد؛ تست warmup بعد از افزودن navigation مجدداً موفق بود. vet و build/typecheck storefront موفق‌اند.
- [۲۰کنترل Playwright](../../../tests/e2e/category-menu.cjs)، بدون runtime pageerror: همه ریشه‌ها، سه‌سطح داده واقعی، لینک والد/برگ، خالی‌بودن ادامه نوار، Escape/focus، ArrowDown، کلیک بیرون، موبایل۳۹۰، قفل/برگشت scroll، منو در سبد و خطای شبکه/retry. یک503عمدی برای آزمون retry شبیه‌سازی شد؛ ادعای صفر خطای network نمی‌شود.
- بررسی بصری، clipping ناشی از overflow قدیمی nav را مشخص و اصلاح کرد؛ کنترل مستقل `popover-not-clipped-by-navbar` اضافه شد.
- [دسکتاپ](../../../output/playwright/issue-46/menu-desktop.png) · [ریشه‌ها در موبایل](../../../output/playwright/issue-46/menu-mobile-roots.png) · [شاخه در موبایل](../../../output/playwright/issue-46/menu-mobile-branch.png).

[شرح/ACها](tasks/cv-sf-ux-02-hierarchical-category-menu.md) · [فهرست ماشین‌خوان](issue-46-checks.json). هیچ دسته، محصول، سابقه سفارش/مالی، قیمت یا موجودی برای این UI تغییر نکرد. اجرای محلی با `Start-Cherriva.ps1` از D:\Cherriva است؛ انتشار لیارا با پارساست. PR/CI و کامنت بستن پس از گیت تحویل ثبت می‌شوند.
