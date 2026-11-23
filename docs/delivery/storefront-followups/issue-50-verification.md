---
id: DEL-SF-UX-050
title: تحویل چیپ مرتب‌سازی با رنگ اصلی
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-SF-UX-006]
approved_by: []
implementation_status: review
---

# نتیجه #50

گروه مرتب‌سازی از ظاهر chips در ChoiceGroup استفاده می‌کند؛ input بومی radio حفظ و فقط نمای آن پنهان است. چیپ منتخب سراسر primary فعلی #B51F46 و متن سفید دارد. انتخاب تک‌گزینه‌ای، جهت‌نماهای کیبورد و focus حفظ شده‌اند. گزینه پرفروش‌ترین نیز همین ظاهر را دارد؛ سایر گروه‌های انتخاب، query و پیش‌فرض مرتب‌سازی تغییر نکردند.

## شواهد پذیرش

build/typecheck موفق؛ [اسکریپت مرورگر](../../../tests/e2e/sort-chips.cjs) با۸کنترل موفق: رنگ و متن، input پنهان، انتخاب یکتا، جهت‌نما، Refresh، Back، موبایل۳۹۰ و نبود pageerror. این بررسی صرفاً نمایش است؛ هیچ داده سفارش/مالی تغییر نکرد.

[دسکتاپ](../../../output/playwright/issue-50/sort-desktop.png) · [موبایل](../../../output/playwright/issue-50/sort-mobile.png) · [شرح و AC](tasks/cv-sf-ux-06-sort-primary-chips.md).

کد در D:\Cherriva است؛ انتشار لیارا با پارساست. وضعیت تحویل پس از CI و ادغام ثبت می‌شود.
