# سایت مشتری Cherriva

رابط بررسی با داده ساختگی؛ وضعیت، دامنه، رنگ برند و موارد باز در [گزارش تحویل](../../docs/delivery/storefront-review.md) ثبت شده‌اند.

## اجرای محلی

```sh
npm ci
npm run dev -- --host 127.0.0.1 --port 5174
```

در یک ترمینال دیگر از ریشه مخزن، با Python Playwright و Chrome نصب‌شده:

```sh
python3 tests/e2e/storefront-smoke.py
```

آدرس مقصد را می‌توان به‌عنوان آرگومان اول آزمون داد. آزمون سبد فقط در حافظه همان مرورگر اجرا می‌شود و عملیاتی در سرور ثبت نمی‌کند.

## ساخت و استقرار

```sh
npm run build
```

خروجی در `build/` است. از ریشه مخزن و با احراز هویت محلی Liara:

```sh
npx @liara/cli@9 deploy --app cherriva-store-review --path apps/storefront-web --platform react --liara-json apps/storefront-web/liara.json --no-app-logs
```

نام app برای انتشار مجدد ثابت است؛ app جدید ایجاد نشود. secret و متغیر API برای این نسخه لازم نیست. `.liaraignore` وابستگی‌ها، خروجی build و فایل‌های محیطی را از بسته سورس حذف می‌کند. از مقدار public در متغیر محیطی frontend به‌عنوان secret استفاده نشود.

ظاهر کاتالوگ و مودال در [راهنمای طراحی بصری چریوا](../../docs/experiences/storefront/visual-design.md) ثبت شده است. تصاویر نمونه SVG و WebP از `public/images/` و فونت محلی استعداد از `public/fonts/` بارگذاری می‌شوند. [مجوز فونت](public/fonts/estedad-ofl.txt) همراه فایل آن است. در هدر، نشان B از بسته برند جای حرف اول لوگوتایپ چریوا را می‌گیرد.

## لوگو و بررسی جریان تا پرداخت

نشان B، Open C، طبق انتخاب کاربر از بسته برند استفاده می‌شود. فایل‌های لوگوتایپ و favicon در `public/brand/` هستند.

برای ورود نمونه شماره `09000000003` و کد `123456` را استفاده کنید. از سبد «ادامه خرید و انتخاب آدرس» را بزنید؛ پس از تکمیل اطلاعات و تأیید روش‌های ارسال به صفحه پرداخت می‌رسید. این صفحه درگاه متصل ندارد و هیچ تراکنشی انجام نمی‌دهد. همه داده‌ها فقط تا reload در حافظه باقی می‌مانند.

آزمون تکمیلی از ریشه مخزن:

```sh
python3 tests/e2e/storefront-checkout-smoke.py
```
