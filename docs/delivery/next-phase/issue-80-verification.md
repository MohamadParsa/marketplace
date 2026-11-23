---
id: DEL-NEXT-080
title: گزارش اجرای Issue 80
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [DEL-NEXT-001]
approved_by: []
implementation_status: done
---

# نتیجه #80

فیلتر مقادیر خصوصیت مانند رنگ/گارانتی از Switch به checkbox بومی سفارشی تبدیل شد؛ رنگ primary، علامت انتخاب، focus و touch target برقرارند. چیپ هر مقدار بالای گزینه‌های همان ویژگی قابل حذف است؛ گزینه ششم پس از جمع‌شدن هم چیپ دارد. کنترل «فقط کالاهای موجود» و رضایت خرید Switch باقی‌اند. این تغییر جایگزین محدود بخش attribute در #49 است، نه تغییر کل Switchهای سایت.

جهت جدید را مالک صریحاً مشخص کرد؛ وابستگی سخت به گروه‌بندی معنایی #79 برای این برش حذف شد. #79 همچنان مستقل و انتخاب‌نشده است. slider، نمایش۵گزینه/بیشتر، URL، OR/AND و تطبیق یک تنوع دست‌نخورده‌اند؛ هیچ مدل یا query فیلتر موازی ساخته نشد.

**۱۳کنترل جدید و۲۳رگرسیون PLP موفق:** کیبورد، primary، دو مقدار، چیپ ششم، محل چیپ، حذف مستقل، Refresh/Back، موبایل، پاک‌سازی و slider/ریال‌تومان/دامنه قیمت/درگ تک‌درخواست/empty و تک‌قیمت. موارد لبه fixture مرورگرند؛ کاتالوگ نوشته نشد. [تست جدید](../../../tests/e2e/filter-checkbox-chips.cjs) · [رگرسیون](../../../tests/e2e/plp-filters.cjs).

[دسکتاپ](../../../output/playwright/issue-80/filters-desktop.png) · [موبایل](../../../output/playwright/issue-80/filters-mobile.png) · [رگرسیون slider](../../../output/playwright/issue-80/regression/price-range-desktop.png).

## اعتبار و مرز تحویل

build/lint سایت و بک‌آفیس موفق؛ Go test کامل با PostgreSQL/Redis محلی در schemaهای جدا: **۲۱۷تست/زیرتست موفق، یک skip اختیاری فایل بزرگ**؛ go vet موفق. race محلی اجرا نشده؛ job API در CI شامل race و vet موفق بود. warning حجم chunk بک‌آفیس از قبل وجود دارد؛ dependency تازه اضافه نشد. مجموعه یکتای مرورگر این تحویل **۸۴کنترل موفق** است و موارد مشترک دوباره شمرده نمی‌شوند. داده مالی/بانکی/سفارش و بسته تاریخی تغییر نکردند. عکس‌ها نیز بصری بازبینی شدند.

نسخه بررسی از D:\Cherriva روی localhost:5197 و localhost:8098 اجراست؛ انتشار لیارا با پارساست. هر پنج بررسی CI موفق شدند؛ تغییرات در [PR #92](https://github.com/MohamadParsa/marketplace/pull/92) با کامیت `f17fb49995647431b6dbe519aff347151d710068` روی main ادغام و ایشو با کامنت شواهد بسته شد. انتشار لیارا انجام نشده است.
