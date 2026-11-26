---
id: VERIFY-AXIS-CONVERSION-20261126
title: تحویل مستقل تبدیل کنترل‌شده محور تنوع — #109
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-BO-REFINEMENT-003]
approved_by: []
implementation_status: done
execution_authorized: true
---

# مسیر بررسی

بک‌آفیس ← دسته‌بندی‌ها ← تنظیم خصوصیات ← «تبدیل کنترل‌شده محورها». محورهای مقصد و مقدار صریح **هر تنوع** انتخاب می‌شوند؛ ابتدا اثر/برخورد، سپس برنامه نسخه‌دار و سرانجام تأیید مستقل اجرا نمایش داده می‌شود. عملیات نیازمند مجوز مستقل `category.axes.convert` است؛ پیش‌فرض فقط ادمین و قابل انتساب به نقش دیگر.

[قرارداد API، مرز نگاشت و اندازه تراکنش](../../architecture/contracts/category-axis-conversion.md) · [نتیجه ماشین‌خوان](issue-109-checks.json) · [BDD مستند](../../bdd/features/experiences/category-axis-conversion.feature).

## رفتار و ردیابی پذیرش

| معیار | رفتار و شاهد |
| --- | --- |
| AC01/02 | preview هیچ داده/audit نمی‌نویسد؛ افزودن، حذف آخرین محور و جایگزینی آزموده شدند. مقدار ناقص، ترکیب تکراری و ارتباط ناسازگار عرضه مانع برنامه‌اند. [تست واحد](../../../apps/api/internal/application/backoffice/axis_conversion_test.go) |
| AC03 | fingerprint به نسخه دسته/تعریف/محصول/تنوع، عرضه و رزرو، سبد و previewهای مرتبط وابسته است. تغییر قیمت/موجودی، رزرو هم‌زمان، انقضا و سلب مجوز اجرا را رد می‌کنند؛۱۰retry هم‌زمان یک رسید ثابت دارند. خطای عمدی انتهای ذخیره PostgreSQL تمام تراکنش را برمی‌گرداند. [تست HTTP و PG](../../../apps/api/internal/bootstrap/axis_conversion_test.go) |
| AC04 | تنوع/SKU/رسانه و شناسه عرضه و قیمت/stock محفوظ، مقادیر قبلی آرشیو و سابقه سفارش/دفتر ثابت‌اند. سبد انتخاب قبلی و فعلی را نشان می‌دهد و تأیید نسخه فعلی لازم است؛ کم/زیادکردن تعداد جای رضایت نیست. preview بی‌سفارش منقضی و کش PDP/PLP تازه می‌شود. [مرورگر](../../../tests/e2e/category-axis-conversion.cjs) |

طبق پاسخ صریح مالک، برخورد چندبه‌یک مسدود و رزرو فعال مانع apply است. اجرا روی داده فعلی کاتالوگ خودکار انجام نشده؛ هر تبدیل واقعی برنامه و تأیید همان تبدیل را می‌خواهد.

## آزمون و حفظ داده

۳۵۰تست/زیرتست Go و۴skip اختیاری روی سورس مستقل #109 موفق بودند؛ فایل‌های آماده #113 هنگام این کنترل کنار گذاشته شدند. PostgreSQL ایزوله و Redis فعال بودند. `go vet ./...` و build/TypeScript بک‌آفیس و storefront موفق‌اند؛ هشدار قدیمی bundle بزرگ بک‌آفیس باقی است. وابستگی تازه اضافه نشده؛ race محلی اجرا نشد و نتیجه CI جدا ثبت می‌شود.

۱۱کنترل مرورگر روی حافظه ایزوله با `CHERRIVA_AXES_BROWSER=1` موفق: نگاشت ناقص، برخورد، تأیید صریح، stale واقعی پس از تغییر stock، حفظ draft، موبایل، اجرای واقعی برنامه و تأیید دوباره سبد مهمان و منع عملیات فاقد مجوز. runtime pageerror صفر بود؛ پاسخ۴۰۹ عمدی آزموده شد. BDD فقط سند است و runner مستقل ادعا نمی‌شود.

نسخه محلی در `D:\Cherriva` راه‌اندازی شد؛ Windows هر دو۸۰۹۸/۵۱۹۷ را۲۰۰ می‌بیند. ۲۴بخش هویت/هش رمز/نقش/عضویت، سفارش/دفتر/بانک و کاتالوگ/قیمت/موجودی/audit پیش/پس restart یکسان بودند. snapshot کاتالوگ با SHA-256 manifest برابر است. بسته تاریخی تغییر نکرد؛ چهار مغایرت قبلی manifest در design.css/design.js/handoff-fa.md/purchase-controls.js با revision آغاز کار یکسان باقی ماندند. warmup صد محصول و۴listing با صفر failure ساخت؛ Redis PONG دارد و بعضی فراخوانی‌های startup از fallback استفاده کردند.

## تصاویر

[برخورد](../../../output/playwright/issue-109/mapping-collision.png) · [برنامه](../../../output/playwright/issue-109/prepared-plan.png) · [موبایل](../../../output/playwright/issue-109/plan-mobile.png) · [سبد نیازمند رضایت](../../../output/playwright/issue-109/cart-requires-consent.png) · [سبد تأییدشده](../../../output/playwright/issue-109/cart-reviewed.png). داده‌ها ساختگی‌اند و تصاویر بازبینی شده‌اند.

## نسخه و محدودیت

سورس مستقل: [a029b88](https://github.com/MohamadParsa/marketplace/commit/a029b88f6d78d9745311cf4afbdeb8dd3f4538d9). برای این قابلیت تغییر DDL لازم نشد؛ برنامه و آرشیو تنوع به JSONB موجود افزوده شدند. تراکنش دسته محدود به۱۰۰۰محصول/۱۰۰۰۰تنوع و بدنه۸MiB است؛ زمان قفل تولید benchmark نشده است. ادغام چندبه‌یک و تغییر تعریف مشترک enum در این نسخه انجام نمی‌شود. #113/#114/#115 تحویل‌های مستقل‌اند. main و Liara ادغام/مستقر نشده‌اند؛ رسید پوش و CI پس از ثبت اضافه می‌شود.

## رسید نهایی

[سورس](https://github.com/MohamadParsa/marketplace/commit/a029b88f6d78d9745311cf4afbdeb8dd3f4538d9) و [مدارک اولیه](https://github.com/MohamadParsa/marketplace/commit/d317123a39504fd2c190b4ed69c6d0a61a673534) پوش شدند؛ [کامنت پایان](https://github.com/MohamadParsa/marketplace/issues/109#issuecomment-7189712677) ثبت و #109 completed بسته شد. هر۵job [CI](https://github.com/MohamadParsa/marketplace/actions/runs/45803253989) روی مدارک با source tree یکسان موفق‌اند؛ race/vet شامل آن است. [رسید ماشین‌خوان](issue-109-delivery.json). #113/#114/#115 جدا، و main/Liara دست‌نخورده‌اند.
