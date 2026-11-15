# استقرار نسخه بررسی بک‌آفیس

این بسته فقط برای داده ساختگی است؛ ورود آزمون، notifier حافظه و آداپتر مالی آزمون دارد. پس از restart داده‌ها از fixture بازسازی می‌شوند و اتصال واقعی پرداخت/پیامک/حمل وجود ندارد. این نسخه مجوز استفاده عملیاتی با داده واقعی ندارد.

یک برنامه Docker در Liara، API و خروجی مستقل build بک‌آفیس را از یک origin ارائه می‌کند. این بسته‌بندی برای بررسی صفحات است و مرز کد frontend و backend را ادغام نمی‌کند.

از ریشه مخزن:

```sh
npx @liara/cli@9 login
npx @liara/cli@9 network:create --name marketplace-review
npx @liara/cli@9 create --app marketplace-bo-review --platform docker --plan small-g2 --feature-plan basic --network marketplace-review
npx @liara/cli@9 deploy --app marketplace-bo-review --path . --platform docker --dockerfile deploy/liara/Dockerfile --liara-json deploy/liara/liara.json --port 8080 --no-app-logs
```

برنامه و شبکه بالا در ۲۰۲۶-۱۱-۱۴ ایجاد شدند؛ برای استقرار بعدی فقط فرمان deploy اجرا شود. رمز/توکن در repository یا دستور ثبت‌شده قرار نمی‌گیرد.

## نتیجه استقرار

[نسخه عمومی بک‌آفیس](https://marketplace-bo-review.liara.run) با release `aaknbxkd912k` منتشر شده است. build کانتینر، health check و smoke test مرورگر روی ۱۵ نما موفق بودند. گزارش کامل در [پیشرفت مرحله ۱](../../docs/delivery/stage-1-progress.md) آمده است.

ورود مدیر آزمون با `09000000001` انجام می‌شود؛ کد آزمون در خود فرم نمایش داده می‌شود. برای بررسی نقش‌های دیگر از دکمه حساب‌های آزمون استفاده کنید.

طبق اعلام Liara، زیردامنه پیش‌فرض تا یک هفته بعد از اولین deploy معتبر است. برای دسترسی پایدار باید دامنه اختصاصی به برنامه اضافه شود؛ هنوز چنین دامنه‌ای تنظیم نشده است.
