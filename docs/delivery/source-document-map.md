# تطبیق کامیت سورس و مستندات چریوا

مسیر کاری و سورس محلی: **D:\Cherriva**. سورس Go/React در `apps` و `packages` است؛ تمام dependencyها با lockfile ثبت‌اند و `node_modules` و خروجی build سورس محسوب نمی‌شوند.

سورس کامل با [PR #38](https://github.com/MohamadParsa/marketplace/pull/38) در `main` ادغام شد: [کامیت 3163d17](https://github.com/MohamadParsa/marketplace/commit/3163d1702f2396d8d1818d12c7bf92bcab9a35bf). درخت کد اجرایی این کامیت با سورس تأییدشده `2cab27a` یکسان است. همگام‌سازی سورس، استقرار Liara ایجاد نکرد. [راهنمای اجرای محلی](../local-development.md).

## ارتباط نسخه‌ها

| تحویل | کامیت سورس | کامیت مستندات | شاهد |
| --- | --- | --- | --- |
| #27 | [8c9ba90](https://github.com/MohamadParsa/marketplace/commit/8c9ba90e0338aaea66e6e537f716147c0c76a2ca) | [31be4ce](https://github.com/MohamadParsa/marketplace/commit/31be4ce4d3ff08d8fcc672d62b65195429c0972f) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/31be4ce4d3ff08d8fcc672d62b65195429c0972f/docs/delivery/initial-launch/issue-27-verification.md) |
| #28 | [8cc3773](https://github.com/MohamadParsa/marketplace/commit/8cc37732270b3a7b1e470ec8879db8c81fd82e15) | [8cc3773](https://github.com/MohamadParsa/marketplace/commit/8cc37732270b3a7b1e470ec8879db8c81fd82e15) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/8cc37732270b3a7b1e470ec8879db8c81fd82e15/docs/delivery/initial-launch/issue-28-verification.md) |
| #29 | [d6e6662](https://github.com/MohamadParsa/marketplace/commit/d6e666235efd70123c0f9279751620f902ab42f3) | [d6e6662](https://github.com/MohamadParsa/marketplace/commit/d6e666235efd70123c0f9279751620f902ab42f3) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/d6e666235efd70123c0f9279751620f902ab42f3/docs/delivery/initial-launch/issue-29-verification.md) |
| #30 | [7d0e325](https://github.com/MohamadParsa/marketplace/commit/7d0e32502301d53c8266ca0b844865b4d10184fe) | [5da8c48](https://github.com/MohamadParsa/marketplace/commit/5da8c48cdad5a0bef8ab76181886b653859df33b) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/5da8c48cdad5a0bef8ab76181886b653859df33b/docs/delivery/initial-launch/issue-30-verification.md) |
| #25 | [e2deacc](https://github.com/MohamadParsa/marketplace/commit/e2deacce9d1d6fe249c2c00bf5618f21bc51e58f) | [79a48df](https://github.com/MohamadParsa/marketplace/commit/79a48df1ab03e817703de2318de3ea88e0dff881) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/79a48df1ab03e817703de2318de3ea88e0dff881/docs/delivery/review-v1-1/issue-25-verification.md) |
| اصلاح خرید/استان و شهر/مسیر D | [2cab27a](https://github.com/MohamadParsa/marketplace/commit/2cab27adffbc1632f5a60dedad86da5f4a49934f) | [fe72552](https://github.com/MohamadParsa/marketplace/commit/fe72552341bc2a878bacc37f2b2f96ddcb9ba17b) | [گزارش](https://github.com/MohamadParsa/marketplace/blob/fe72552341bc2a878bacc37f2b2f96ddcb9ba17b/docs/delivery/initial-launch/purchase-refinements-verification.md) |

برای هر ردیف، tree فایل‌های اجرایی در کامیت مدارک با کامیت سورس یکسان کنترل شده است؛ این تطبیق به معنی اخذ مجوز یا پذیرش محیط عمومی نیست. برخی تحویل‌ها کد و گزارش را در یک کامیت دارند.

سورس اصلاح جاری: [2cab27a](https://github.com/MohamadParsa/marketplace/commit/2cab27adffbc1632f5a60dedad86da5f4a49934f)؛ اجرای قابلیت در `a52e599`، راه‌اندازی از ریشه پروژه در `997162c` و اصلاح ذخیره تنظیمات در کامیت سورس جدول ثبت شدند. کامیت‌های صرفاً مستنداتی بعدی، شناسه سورس متناظر را در پیام commit و این فهرست دارند.

## دلیل تفاوت مشاهده‌شده در گیت

تا پیش از همگام‌سازی این تحویل، `main` روی `4719643` و عمدتاً شامل تغییرات مستندات بود. پیاده‌سازی‌ها در شاخه‌های `codex/...` و PRهای زنجیره‌ای باقی مانده بودند. سورس پیش‌تر به همین شاخه‌ها پوش شده بود؛ همگام‌نبودن شاخه پیش‌فرض و پوشه قدیمی D باعث اختلاف نسخه قابل مشاهده می‌شد. این تحویل، مسیر D و انتشار یکپارچه سورس را هماهنگ می‌کند.

تغییرات محلی قدیمی D پیش از انتقال در stash با عنوان `Preserve D Cherriva local files before source sync 2026-11-22` نگه داشته شدند؛ اطلاعات خصوصی runtime و دیتابیس بررسی به گیت منتقل نمی‌شوند. بسته تاریخی طراحی تغییر نکرده است.

[راه‌اندازی و حساب محلی](local-review-access.md) · [قرارداد اصلاحات](../architecture/contracts/purchase-experience.md) · [فهرست ماشین‌خوان](source-document-map.json)
