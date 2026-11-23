---
id: DEL-NEXT-086
title: گزارش اجرای Issue 86
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [DEL-NEXT-001]
approved_by: []
implementation_status: done
---

# نتیجه #86

طبق پاسخ مالک، عدد کوچک‌تر مقدم، مقدار خالی پس از اولویت‌دارها و ترتیب فقط بین siblingهای یک والد است. فیلد اختیاری display_priority در Category، فرم ایجاد/ویرایش و هر دو گرید دسته اضافه شد. محدوده صحیح۰تا۱۰۰۰۰۰۰ انتخاب فنی است؛ صفر معتبر و تساوی با نام سپس ID پایدار است.

همان catalog.category.create/update، permission و version/idempotency/audit مصرف می‌شوند. display_priority حذف‌شده یا null در درخواست قدیمی مقدار موجود را حفظ می‌کند؛ clear_display_priority:true حذف صریح است. flag حذف همراه مقدار عددی رد می‌شود. ساختار درختی و parent_id تغییر نمی‌کنند. DTO عمومی flat/parent-linked باقی است و سرور به‌صورت preorder با مرتب‌سازی مستقل هر سطح آن را می‌سازد. ترتیب محصولات یا آیکن‌های خانه در این برش تغییر نکرد.

فیلد در JSON پایدار workspace/خوانش کاتالوگ موجود ذخیره می‌شود؛ migration یا backfill مخرب لازم نبود. تغییر معتبر همان تراکنش نسخه وابستگی و کش navigation/ETag را تازه می‌کند؛ warmup موجود همان خواندن عمومی را مصرف می‌کند.

**تست Go** مجوز403، stale409، retry یک‌باره، صفر/بدون مقدار، محدوده نامعتبر، clear صریح، حفظ legacy omission، create، تساوی/پنهان/محدوده sibling، تغییر ETag/کش و reopen PostgreSQL را در schema موقت بررسی کرد. **۱۱کنترل فرم مرورگر موفق:** serialization صفر/clear، حفظ سایر فیلدها، خواندن workspace، validation اعشار/بازه، موبایل و cancel. فرمان‌های مرورگر intercept شدند؛ دسته واقعی gsp-cat-755 ویرایش نشد. فقط نشست حساب ساختگی تیم برای ورود استفاده شد. ماندگاری واقعی سرور را تست PostgreSQL پوشش می‌دهد، نه mock پنل.

[تست Go](../../../apps/api/internal/bootstrap/category_priority_test.go) · [تست فرم](../../../tests/e2e/category-priority-panel.cjs) · [دسکتاپ](../../../output/playwright/issue-86/priority-panel-desktop.png) · [موبایل](../../../output/playwright/issue-86/priority-panel-mobile.png).

## اعتبار و مرز تحویل

build/lint سایت و بک‌آفیس موفق؛ Go test کامل با PostgreSQL/Redis محلی در schemaهای جدا: **۲۱۷تست/زیرتست موفق، یک skip اختیاری فایل بزرگ**؛ go vet موفق. race محلی اجرا نشده؛ job API در CI شامل race و vet موفق بود. warning حجم chunk بک‌آفیس از قبل وجود دارد؛ dependency تازه اضافه نشد. مجموعه یکتای مرورگر این تحویل **۸۴کنترل موفق** است و موارد مشترک دوباره شمرده نمی‌شوند. داده مالی/بانکی/سفارش و بسته تاریخی تغییر نکردند. عکس‌ها نیز بصری بازبینی شدند.

نسخه بررسی از D:\Cherriva روی localhost:5197 و localhost:8098 اجراست؛ انتشار لیارا با پارساست. هر پنج بررسی CI موفق شدند؛ تغییرات در [PR #92](https://github.com/MohamadParsa/marketplace/pull/92) با کامیت `f17fb49995647431b6dbe519aff347151d710068` روی main ادغام و ایشو با کامنت شواهد بسته شد. انتشار لیارا انجام نشده است.
