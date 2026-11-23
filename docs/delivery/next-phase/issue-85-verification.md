---
id: DEL-NEXT-085
title: گزارش اجرای Issue 85
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [DEL-NEXT-001]
approved_by: []
implementation_status: review
---

# نتیجه #85

علت clipping، اتکای grid به max-height بدون row height قابل کوچک‌شدن بود. ارتفاع panel از viewport و موقعیت واقعی‌اش محدود می‌شود؛ row صریح minmax(0,1fr)، min-height:0 و paneهای overflow مستقل به ریشه/فرزند اجازه اسکرول می‌دهند. wheel در pane، body را نمی‌چرخاند. موبایل یک scroll container و body-lock خود را حفظ می‌کند؛ focus trap فقط عناصر واقعاً نمایان را می‌شمارد.

**۲۱کنترل مشترک #84/#85 و۵کنترل واقعی موفق**: ارتفاع۶۵۰/۳۶۰ دسکتاپ، آخرین ریشه/نوه، دو pane، موبایل۳۹۰×۶۰۰، body-lock/restore، Escape، focus و نبود overflow. [تست fixture](../../../tests/e2e/category-menu-hover-scroll.cjs) و [خواندن کاتالوگ واقعی](../../../tests/e2e/category-menu-real-scroll.cjs) با۴۱ریشه و۲۴۶گره؛ هیچ نوشتن کاتالوگ یا سفارش انجام نشد. شمار۲۱ در جمع تحویل تنها یک بار حساب می‌شود.

[اسکرول دسکتاپ](../../../output/playwright/issue-85/menu-scrolled-desktop.png) · [موبایل](../../../output/playwright/issue-85/menu-mobile.png) · [واقعی دسکتاپ](../../../output/playwright/issue-85/menu-real-desktop.png) · [واقعی موبایل](../../../output/playwright/issue-85/menu-real-mobile.png).

## اعتبار و مرز تحویل

build/lint سایت و بک‌آفیس موفق؛ Go test کامل با PostgreSQL/Redis محلی در schemaهای جدا: **۲۱۷تست/زیرتست موفق، یک skip اختیاری فایل بزرگ**؛ go vet موفق. race محلی اجرا نشده و در CI بررسی می‌شود. warning حجم chunk بک‌آفیس از قبل وجود دارد؛ dependency تازه اضافه نشد. مجموعه یکتای مرورگر این تحویل **۸۴کنترل موفق** است و موارد مشترک دوباره شمرده نمی‌شوند. داده مالی/بانکی/سفارش و بسته تاریخی تغییر نکردند. عکس‌ها نیز بصری بازبینی شدند.

نسخه بررسی از D:\Cherriva روی localhost:5197 و localhost:8098 اجراست؛ انتشار لیارا با پارساست. نتیجه CI/ادغام و بستن با کامنت بعداً در همین گزارش ثبت می‌شود.
