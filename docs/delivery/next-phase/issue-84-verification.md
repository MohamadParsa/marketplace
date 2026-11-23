---
id: DEL-NEXT-084
title: گزارش اجرای Issue 84
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [DEL-NEXT-001]
approved_by: []
implementation_status: done
---

# نتیجه #84

hover متن یا فلش ریشه/گره‌های دارای فرزند در دسکتاپ زیرمنو را باز می‌کند؛ trigger اصلی نیز با hover باز می‌شود. intent۱۲۰میلی‌ثانیه و لغو با خروج/pointerdown از پرش و بازشدن دوباره بعد از کلیک جلوگیری می‌کند. کلیک متن همچنان لینک دسته و کلیک فلش toggle است. hover روی دستگاه touch/عرض موبایل فعال نیست.

بازشدن با hover focus را نمی‌دزدد. منوی حساب با بازشدن منوی دسته، حتی بدون pointerdown، بسته می‌شود. حرکت به زیرمنو آن را نمی‌بندد؛ روش‌های بستن خارج/Escape و کیبورد حفظ‌اند. این تغییر هیچ دسته/محصولی اضافه یا پنهان نمی‌کند.

**مجموعه مشترک۲۱کنترلی #84/#85 موفق** در [category-menu-hover-scroll.cjs](../../../tests/e2e/category-menu-hover-scroll.cjs): متن/فلش/سطح سوم، حرکت pointer، click navigation، focus، هماهنگی منوی حساب، touch و اسکرول/بستن. داده ناوبری fixture۴۰ریشه/۳۵فرزند/۴۰نوه است؛ navigation به دسته fixture می‌تواند404 مورد انتظار بدهد، ولی runtime pageerror نداشت.

[hover دسکتاپ](../../../output/playwright/issue-84/menu-hover-desktop.png).

## اعتبار و مرز تحویل

build/lint سایت و بک‌آفیس موفق؛ Go test کامل با PostgreSQL/Redis محلی در schemaهای جدا: **۲۱۷تست/زیرتست موفق، یک skip اختیاری فایل بزرگ**؛ go vet موفق. race محلی اجرا نشده؛ job API در CI شامل race و vet موفق بود. warning حجم chunk بک‌آفیس از قبل وجود دارد؛ dependency تازه اضافه نشد. مجموعه یکتای مرورگر این تحویل **۸۴کنترل موفق** است و موارد مشترک دوباره شمرده نمی‌شوند. داده مالی/بانکی/سفارش و بسته تاریخی تغییر نکردند. عکس‌ها نیز بصری بازبینی شدند.

نسخه بررسی از D:\Cherriva روی localhost:5197 و localhost:8098 اجراست؛ انتشار لیارا با پارساست. هر پنج بررسی CI موفق شدند؛ تغییرات در [PR #92](https://github.com/MohamadParsa/marketplace/pull/92) با کامیت `f17fb49995647431b6dbe519aff347151d710068` روی main ادغام و ایشو با کامنت شواهد بسته شد. انتشار لیارا انجام نشده است.
