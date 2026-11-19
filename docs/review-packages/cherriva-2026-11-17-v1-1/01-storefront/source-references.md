# مراجع بررسی‌شده و مرز این بسته

مخزن محلی `D:/Cherriva` روی commit `d6e9a64b4bd5c4eee9b5821c691984ac5647e210` بررسی شده است. وجود تغییرات محلی قبلی در README و ابزارهای اجرای محلی حفظ شد. این بسته در پوشه‌ای مستقل ذخیره شده و هیچ commit، push یا انتشار محصول انجام نشده است.

| موضوع | مرجع نسخه بررسی‌شده |
| --- | --- |
| نقشه مستندات | [docs/README](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/README.md) |
| راهنمای مستندسازی | [documentation-guide](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/governance/documentation-guide.md) |
| دامنه | [scope](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/product/scope.md) |
| واژه‌نامه و نقش‌ها | [glossary](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/product/glossary.md)، [roles](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/product/roles.md) |
| پرسش‌های محصول | [open-questions](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/product/open-questions.md) |
| قیمت و پرداخت | [ADR-001](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/decisions/adr-001-pricing-and-payment-boundaries.md) |
| رفتار کشف محصول | [discovery](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/experiences/storefront/discovery.md) |
| کاتالوگ و آمار فروش | [catalog](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/domains/catalog/README.md) |
| ترکیب، فروشنده، سبد و مبنای قیمت | [checkout](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/domains/checkout/README.md) |
| رنگ و نسخه بررسی storefront | [storefront-review](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/delivery/storefront-review.md) |
| فایل لوگوی کپی‌شده | [open-c-main.svg](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/apps/storefront-web/public/brand/open-c-main.svg) |

لینک‌ها به snapshot بررسی‌شده‌اند؛ تغییرات بعدی مخزن خودکار در طرح منعکس نمی‌شوند. جدیدترین دستور صریح کاربر درباره PDP مستقل و حذف خرید/تنوع روی کارت، در این طراحی مبناست و اختلاف آن با اسناد قدیمی در راهنمای تحویل ثبت شده است. قواعد حوزه‌ای که کاربر تغییر نداده، از یک پیشنهاد بصری جدید برداشت نشده‌اند.


## مراجع تکمیلی نسخه ۰٫۳

فایل‌های تجربه checkout، هویت، حمل و پرداخت از همان snapshot مخزن چریوا بررسی شدند:

| موضوع | مرجع |
| --- | --- |
| تجربه خرید FEAT-SF-002 | [checkout experience](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/experiences/storefront/checkout.md) |
| قواعد پرداخت | [payments](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/domains/payments/README.md) |
| هویت و OTP | [identity](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/domains/identity/README.md) |
| ارسال | [shipping](https://github.com/MohamadParsa/marketplace/blob/d6e9a64b4bd5c4eee9b5821c691984ac5647e210/docs/domains/shipping/README.md) |

اجزای `apps/storefront-web/src/checkout/checkout-page.tsx` و `payment-step.tsx` نیز فقط برای ایده و مقایسه خوانده شدند. فرم‌ها و نام ارائه‌دهنده از روی آن‌ها سیاست تازه محصول تلقی نشده‌اند.

در مرجع گوشی‌شاپ، صفحه عمومی و اجزای محلی frontend بررسی شدند؛ مرور بصری کامل مراحل زنده ممکن نشد. [شرح دقیق شواهد و اقتباس](reference-adaptations-fa.md) محدودیت را ثبت کرده است. نه مخزن چریوا و نه مخزن مرجع در این کار ویرایش نشده‌اند.
