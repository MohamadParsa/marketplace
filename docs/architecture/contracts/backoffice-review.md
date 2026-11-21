---
id: API-BO-REVIEW-001
title: قرارداد اجرایی نسخه بررسی بک‌آفیس
status: in-review
owner: product-owner
updated: 2026-11-21
related: [FEAT-BO-001, FEAT-BO-002, FEAT-BO-003]
approved_by: []
---

# قرارداد نسخه بررسی بک‌آفیس

تکمیل #6: [API-PRODUCT-WORKSPACE-001](product-workspace.md) مسیرهای تنوع/عرضه و workspace محدود seller را مشخص می‌کند؛ الگوی کلی فرمان، audit و مخزن حافظه‌ای حفظ شده است. رجیستری نهایی نقش/عضویت در #1 جداست.

تکمیل برش آزمون Issue#5: [خصوصیات مشترک و سقف کارت](catalog-attributes.md) فرمان‌های کتابخانه، انتساب و برش نمایش از تنظیمات عمومی را تعریف می‌کند؛ تنظیمات کامل و تبدیل تنوع استفاده‌شده از این برش نتیجه نمی‌شوند.

## هدف و طرف‌ها

رابط بک‌آفیس React و API مشترک Go برای بررسی با داده ساختگی. این قرارداد مجوز عملیات بانکی، پیامک واقعی یا نگهداری داده عملیاتی نیست. schema پاسخ در `apps/api/internal/application/backoffice/model.go` و انواع رابط در `apps/backoffice-web/src/shared/types.ts` نگهداری می‌شوند؛ تولید خودکار client هنوز انجام نشده و از محدودیت‌های این نسخه است.

## عملیات و دسترسی

| روش و مسیر | کاربرد | دسترسی |
| --- | --- | --- |
| `POST /backoffice/auth/request-otp` | درخواست OTP آزمون | شماره کاربر موجود |
| `POST /backoffice/auth/verify-otp` | ایجاد نشست | کد معتبر و حساب فعال |
| `POST /backoffice/auth/logout` | ابطال نشست | Bearer token |
| `GET /backoffice/workspace` | نمای یکپارچه صفحات آزمون | کاربر بک‌آفیس؛ کاربران فقط برای super-admin، اطلاعات مالی فقط برای super-admin/finance |
| `POST /backoffice/commands` | اجرای فرمان با ثبت audit | طبق مجوز هر فرمان |

در هر درخواست، فعال‌بودن و نقش جاری حساب از repository خوانده می‌شود؛ تغییر نقش یا غیرفعال‌سازی فقط مخفی‌کردن منو نیست. فروشنده حق استفاده از این endpointها را ندارد.

## فرمان‌ها

- `user.create / user.toggle / user.role`: فقط super-admin؛ آخرین super-admin فعال قابل غیرفعال‌سازی یا تنزل نیست.
- `store.create / store.toggle / category.create / category.archive / product.publish / product.reject / shipping.update`: super-admin و operations.
- `case.start / case.close / case.refund / case.assign / order.deliver`: super-admin و support.
- `transfer.resolve`: super-admin و finance.
- `rule.toggle / notification.retry`: super-admin.

بدنه فرمان شامل `action`، `target` و `idempotency_key` و ورودی لازم همان فرمان است. `version` برای محصول، سفارش، پرونده و انتقال مالی لازم است. درخواست مجدد با کلید و بدنه یکسان اثر دوباره ندارد؛ استفاده از همان کلید با بدنه متفاوت conflict است. کل تغییر و audit اتمیک اعمال می‌شوند. بدنه حداکثر ۱۶KB است و فیلد ناشناخته رد می‌شود.

## خطاها

| HTTP | معنا | رفتار رابط |
| --- | --- | --- |
| 400 | ورودی نامعتبر | نمایش خطای فرم |
| 401 | نشست نامعتبر، منقضی یا حساب غیرفعال | خروج و پاک‌سازی query cache |
| 403 | فقدان مجوز | عدم اعمال تغییر |
| 404 | مرجع ناموجود | نمایش خطا |
| 409 | نسخه کهنه یا وضعیت نامعتبر | دریافت مجدد داده و بررسی عملیات |
| 429 | محدودیت OTP | تلاش پس از مهلت |
| 500 | خطای داخلی | پیام عمومی؛ جزئیات داخلی در response نیست |

## قواعد داده و سازگاری

مبلغ API ریال صحیح است؛ نمایش رابط تومان است. زمان ISO 8601 و نمایش با تقویم فارسی و منطقه `Asia/Tehran` انجام می‌شود. پاسخ خصوصی `no-store` است و token فقط در حافظه مرورگر نگهداری می‌شود. داده آزمون محدود است و این endpoint هنوز pagination ندارد؛ برای محیط عملیاتی به قرارداد صفحه‌بندی و query اختصاصی نیاز است.

## نمونه و پذیرش

حساب‌های کاملاً ساختگی `09000000001` تا `09000000005` به‌ترتیب مدیرکل، عملیات، پشتیبانی، مالی و مشاهده‌گر هستند. کد آزمون فقط در محیط local/test نمایش داده می‌شود.

شواهد اجرا: `apps/api/internal/bootstrap/backoffice_test.go`، `apps/api/internal/application/backoffice/service_test.go` و `tests/e2e/backoffice-smoke.py`.

## پرسش‌ها و محدودیت‌ها

این نسخه مرور صفحات و بخشی از عملیات را ممکن می‌کند؛ کل گیت مرحله ۱ نیست. ورود رمزعبوری/بازیابی رمز، persistence/migration، ویرایش کامل محتوا، نسخه قرارداد/تعدیلات کامل مالی، صف پایدار و کل سناریوهای BDD محصول هنوز تکمیل نشده‌اند. انتقال مالی واقعی و ارائه‌دهنده پیامک همچنان جدا هستند. [گزارش تحویل](../../delivery/stage-1-progress.md) مرجع وضعیت است.

## افزوده Issue#4

[قرارداد آیکن](catalog-icons.md) upload تصویر و فرمان‌های category.icon.remove/brand.icon.remove را با مجوز کاتالوگ مدیرکل/عملیات به این نسخه اضافه می‌کند. workspace شامل brands و icon_policy و icon_id/version دسته/برند است؛ باینری icon_media برنمی‌گردد. این افزوده حافظه‌ای آزمون است؛ [شواهد](../../delivery/review-v1-1/issue-4-verification.md).
