---
id: VERIFY-BRAND-MANAGEMENT-20261126
title: شواهد تحویل مستقل مدیریت برند — #115
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-BO-REFINEMENT-009, API-BRAND-MANAGEMENT-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# نتیجه #115

مدل و فرم مستقل `/brands` با نام انگلیسی و توضیحات اختیاری، وضعیت Active موجود، فیلتر نام فارسی/انگلیسی/وضعیت/تصویر و DataTable مشترک تکمیل شد. ایجاد/ویرایش از همان commands و تصویر از endpointهای قبلی است. فرم اثر غیرفعال‌سازی و تعداد محصولات متصل را قبل از ذخیره نشان می‌دهد. نسخه، تکرار امن، نام تکراری و درخواست بی‌مجوز کنترل می‌شوند؛ خطا ورودی را حفظ می‌کند. عکس برند غیرفعال قابل آماده‌سازی است ولی عمومی نیست.

[سورس مستقل](https://github.com/MohamadParsa/marketplace/commit/4ec243011c3c37ddb2fbbc6da9976037e084915e) · [قرارداد](../../architecture/contracts/brand-management.md) · [تسک](tasks/brand-profile-crud-and-filters.md) · [نتیجه ماشین‌خوان](issue-115-checks.json). تغییر محصول، واحدها یا PDP روش‌محور در این کار اجرا نشده‌اند.

| معیار پذیرش | شاهد واقعی |
| --- | --- |
| AC01 ایجاد/ویرایش و ماندگاری | HTTP و PostgreSQL/restart در `apps/api/internal/bootstrap/brand_management_test.go`؛ ایجاد و ویرایش مرورگر از /brands |
| AC02 تصویر، فیلتر و صفحه‌بندی | upload معتبر/نامعتبر، حذف به fallback، تصویر برند غیرفعال غیرعمومی؛ ۴۶برند ساختگی، صفحه بعد، ترکیب فیلتر و نتیجه خالی و موبایل در `tests/e2e/brand-management.cjs` |
| AC03 مجوز/نسخه/حفاظت تاریخچه | 401/403،409 نسخه/نام تکراری،404 رکورد گمشده،400 طول نام، تکرار فرمان؛ مقایسه محصولات/عرضه/سفارش غیرخالی/کاربر/نقش/دفتر/commerce قبل و بعد؛ cache PDP پس از تغییر نام تازه و پس از غیرفعال‌سازی404 |

## بررسی‌ها

- `go test -json ./...` با متغیرهای خصوصی `MARKETPLACE_TEST_POSTGRES_URL` و `MARKETPLACE_TEST_REDIS_URL`: **۳۷۰ تست/زیرتست موفق، ۴ skip اختیاری و صفر شکست**. تست‌های DB در schemaهای جدا اجرا شدند؛ به SQL Server مبدأ وصل نشدیم.
- `go vet ./...` و `npm --prefix apps/backoffice-web run build`: موفق. هشدار شناخته‌شده chunk بیش از۵۰۰KiB باقی است؛ dependency جدید نداریم. race محلی به‌دلیل نبود CGO/toolchain اجرا نشد؛ نتیجه CI جدا در رسید نهایی می‌آید.
- مرورگر: **۱۵کنترل موفق**، صفر runtime pageerror؛ screenshotهای زیر بازبینی شدند. خطای409 و فایل خراب عمداً شبیه‌سازی شدند، پس ادعای نبود هر نوع خطای شبکه نداریم.
- نسخه اصلی محلی از `D:\Cherriva` با `Start-Cherriva.ps1 -NoBuild` و Redis/warmup به‌روز شد؛ BO8098 و سایت5197 هر دو200. مقایسه **۲۸بخش** پس از restart برابر baseline است؛ کاربر/رمز/نقش/بانک/پرداخت/ارسال/کیف پول/دفتر/سفارش/کاتالوگ/رسانه تغییر نکردند. تست مرورگر صرفاً در حافظه جدا روی8097/5199 بود.
- در ادامه پس از وقفه، سرویس PostgreSQL محلی متوقف بود؛ راه‌اندازی و اجرای دوباره روی پورت واقعی همان cluster انجام شد. نتیجه نهایی بالا از اجرای موفق پس از آن است.
- snapshot کاتالوگ و بسته طراحی تاریخی تغییر نکردند. چهار مغایرت manifest طراحی از قبل موجود بود و دست‌نخورده ماند.

## تصاویر

[ایجاد برند](../../../output/playwright/issue-115/brand-create.png) · [اثر غیرفعال‌سازی](../../../output/playwright/issue-115/brand-inactive-impact.png) · [فرم موبایل](../../../output/playwright/issue-115/brand-mobile-edit.png) · [گرید موبایل](../../../output/playwright/issue-115/brands-mobile.png).

## محدودیت و مرز انتشار

تصویر بعد از ذخیره اطلاعات از اکشن همان ردیف مدیریت می‌شود؛ ذخیره متن و فایل دو اقدام صریح و نسخه‌دارند. برند غیرفعال در پیش‌نمایش عمومی آیکن fallback دارد؛ فایل محفوظ و این رفتار توضیح داده شده است. فیلتر و صفحه‌بندی روی workspace مجاز فعلی در کلاینت انجام می‌شود، نه query حجیم سرور. نام‌های انگلیسی قدیمی حدس زده یا بازنویسی نشدند. سند in-review است؛ main/Liara ادغام یا مستقر نشده‌اند.
