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

سورس اصلاح تجربه خرید: [2cab27a](https://github.com/MohamadParsa/marketplace/commit/2cab27adffbc1632f5a60dedad86da5f4a49934f)؛ اجرای قابلیت در `a52e599`، راه‌اندازی از ریشه پروژه در `997162c` و اصلاح ذخیره تنظیمات در کامیت سورس جدول ثبت شدند. کامیت‌های صرفاً مستنداتی بعدی، شناسه سورس متناظر را در پیام commit و این فهرست دارند.

## دلیل تفاوت مشاهده‌شده در گیت

تا پیش از همگام‌سازی این تحویل، `main` روی `4719643` و عمدتاً شامل تغییرات مستندات بود. پیاده‌سازی‌ها در شاخه‌های `codex/...` و PRهای زنجیره‌ای باقی مانده بودند. سورس پیش‌تر به همین شاخه‌ها پوش شده بود؛ همگام‌نبودن شاخه پیش‌فرض و پوشه قدیمی D باعث اختلاف نسخه قابل مشاهده می‌شد. این تحویل، مسیر D و انتشار یکپارچه سورس را هماهنگ می‌کند.

تغییرات محلی قدیمی D پیش از انتقال در stash با عنوان `Preserve D Cherriva local files before source sync 2026-11-22` نگه داشته شدند؛ اطلاعات خصوصی runtime و دیتابیس بررسی به گیت منتقل نمی‌شوند. بسته تاریخی طراحی تغییر نکرده است.

[راه‌اندازی و حساب محلی](local-review-access.md) · [قرارداد اصلاحات](../architecture/contracts/purchase-experience.md) · [فهرست ماشین‌خوان](source-document-map.json)

## بسته کاتالوگ تست — ۲۰۲۶-۱۱-۲۲

سورس و JSON این تحویل: [8921ca0](https://github.com/MohamadParsa/marketplace/commit/8921ca0497aa016d9373ae240c12025046bd3a6d)؛ [راهنما، پرامپت‌ها و manifest](catalog-seed/README.md). کامیت مدارک این بسته `Source-Commit: 8921ca0497aa016d9373ae240c12025046bd3a6d` دارد. ورود داده فقط روی schema موقت آزمون انجام شده؛ دیتابیس فعلی و سرور هنوز تغییری نگرفته‌اند.

کامیت مستندات بسته کاتالوگ: [4941774](https://github.com/MohamadParsa/marketplace/commit/49417740fccb53456f305add5b7dadaf504b00a1)؛ درخت فایل‌های اجرایی، اسکریپت‌ها و JSON با سورس `8921ca0` یکسان بررسی شد.

بسته کاتالوگ با [PR #39](https://github.com/MohamadParsa/marketplace/pull/39) و [ادغام e8d59e0](https://github.com/MohamadParsa/marketplace/commit/e8d59e0b5260c29582334aa62aadd8d4f6e1dee4) به `main` رسید؛ کد/JSON با سورس تأییدشده `8921ca0` یکسان است. CI پنج job موفق بود؛ ورود دیتابیس فعلی و انتشار Liara انجام نشده‌اند.

اصلاح راه‌انداز لوکال: [5e588a7](https://github.com/MohamadParsa/marketplace/commit/5e588a7e047d0aa7d077f959ed6cc3d1c6634e6e)؛ وجود `DATABASE_URL` را پیش از migration رد می‌کند. آزمون guard موفق است؛ کد Go/React و snapshot با `8921ca0` یکسان‌اند.

## ورود کامل کاتالوگ محلی — ۲۰۲۶-۱۱-۲۳

سورس [6342eb9](https://github.com/MohamadParsa/marketplace/commit/6342eb9e38213664c2d8dfde6f865b4cf1cc0571) و مدارک [d35855e](https://github.com/MohamadParsa/marketplace/commit/d35855e7ab4b4474e2d0b49f55def39ea5fc5d16) با tree اجرایی/JSON یکسان کنترل شدند. [گزارش اجرا، محدودیت‌ها و تصاویر](catalog-seed/local-import-2026-11-23.md) شامل ورود کامل، retry بدون تکرار، حفظ اطلاعات، راه‌اندازی مجدد و اصلاح واحد قیمت/گزینه‌های PDP/سازگاری migration است. داده فقط در محیط محلی وارد شده و گیت عمومی #12 همچنان مستقل است.

این تحویل با [PR #40](https://github.com/MohamadParsa/marketplace/pull/40) و ادغام [b1f9a12](https://github.com/MohamadParsa/marketplace/commit/b1f9a12f19441bfd9ef7907c0785be44674a3f57) روی `main` است. نسخه `D:\Cherriva` همگام و اجرای محلی دارای داده کامل است؛ استقرار Liara انجام نشده است.

## کش و اصلاحات کاتالوگ/پنل — #41

سورس [c2617fc](https://github.com/MohamadParsa/marketplace/commit/c2617fcd72c82fad01a51291360a4189f576f719) و مدارک [63c73ac](https://github.com/MohamadParsa/marketplace/commit/63c73ac2486ed9587da2330439f3cb4bb1463e1f) درخت اجرایی یکسان دارند. [گزارش #41، آزمون‌ها، سرعت، تصاویر و پاک‌سازی](storefront-followups/issue-41-verification.md) مرجع این تحویل محلی است؛ تغییر URL در #42 و انتشار Liara جدا هستند.

## کد chr و slug دسته — #42

سورس [1f4fdd0](https://github.com/MohamadParsa/marketplace/commit/1f4fdd01cafae849b0ff564d7c1c55d1c4bf4d2b) و مدارک [4a1a6d8](https://github.com/MohamadParsa/marketplace/commit/4a1a6d8720c8f19a8361f1c572ae669efcafed95) درخت اجرایی یکسان دارند. [گزارش، نگاشت، migration و تصاویر](storefront-followups/issue-42-verification.md) شامل dry-run/apply محلی، حفظ داده، repeat/restart و آزمون301 است. انتشار لیارا جداست.

#42 با [PR #44](https://github.com/MohamadParsa/marketplace/pull/44) و merge `deb949d2782bff9fed982fe9d886870c2321cbcf` روی main است؛ هر۵ job CI از جمله race/vet موفق و Issue با کامنت پایان بسته شد. نسخه محلی با همین سورس راه‌اندازی و warmup شد؛ انتشار لیارا انجام نشده است.

## پرفروش‌ترین — #45

سورس و شواهد اولیه در [b92469a](https://github.com/MohamadParsa/marketplace/commit/b92469a6cec09df95ec7c89dc4d9e37fb1f854db) ثبت شدند؛ [گزارش](storefront-followups/issue-45-verification.md) شامل سنجه خالص، بازه، کش مستقل و آزمون‌هاست. انتشار لیارا جداست.

## منوی درختی — #46

سورس و شواهد اولیه در [3ebdf64](https://github.com/MohamadParsa/marketplace/commit/3ebdf64e7e6cc363c10740163fda63308941ba1b) ثبت شدند؛ [گزارش منو](storefront-followups/issue-46-verification.md). #45 پیش‌تر در PR #55 ادغام و بسته شد.

## آیکن‌های هدر — #47

سورس و شواهد اولیه در [1c890da](https://github.com/MohamadParsa/marketplace/commit/1c890daee14ab2e5320f859e831398ed142757b4) ثبت شدند؛ [گزارش و قبل/بعد](storefront-followups/issue-47-verification.md). #46 پیش‌تر در PR #56 ادغام و بسته شد.

## فیلترهای PLP — #48

سورس و شواهد اولیه در [f812e36](https://github.com/MohamadParsa/marketplace/commit/f812e36e26e05e6dcbce483f130c66a2dec24908) ثبت شدند؛ [گزارش دامنه قیمت و گزینه‌های جمع‌شونده](storefront-followups/issue-48-verification.md).

## Switchهای storefront — #49

سورس و شواهد اولیه در [b3b918e](https://github.com/MohamadParsa/marketplace/commit/b3b918ec68bf63cca5c22f92361b9d1f6aa91e07) ثبت شدند؛ [گزارش پوشش و تصاویر](storefront-followups/issue-49-verification.md). #48 پیش‌تر در PR #58 ادغام و بسته شد.

تحویل انتخاب‌شده #45تا#49 کامل است: PRهای #55تا#59 با هر۵ job موفق CI و mergeهای `[skip ci]` روی main هستند و هر Issue کامنت پایان دارد و بسته است. مسیر `D:\Cherriva` همگام و نسخه محلی آماده بررسی است؛ انتشار لیارا انجام نشده است.

## انتخاب بعدی — #50تا#52

- #50: سورس و مدرک اولیه در [eb2c166](https://github.com/MohamadParsa/marketplace/commit/eb2c1669e4bb23b5e19105cce3e0469aae1eb6b7)؛ [گزارش و تصاویر](storefront-followups/issue-50-verification.md).
- #51: سورس و مدرک اولیه در [03a86ac](https://github.com/MohamadParsa/marketplace/commit/03a86ac1bd622d72655bb5cda355d133dc995799)؛ [گزارش و تصاویر](storefront-followups/issue-51-verification.md).
- #52: سورس و مدرک اولیه در [7f57d07](https://github.com/MohamadParsa/marketplace/commit/7f57d07a85ec60af267fed1036a321016807f6f9)؛ [گزارش و تصاویر](storefront-followups/issue-52-verification.md).

هر سه پیاده و محلی آزموده شده‌اند؛۵۸کنترل مرورگر در مجموع موفق است. هر۵ job CI در PR #60 موفق، هر سه Issue با کامنت بسته و merge 17c9d4890168d4f57e1b94ca8dae0d414bf80d9b روی main است؛ آخرین کد اجرایی با اصلاح اسکرول #51 در f236335e7d06b4f74fa1460fdf792b2236ff97c0 ثبت شده است.

## دو اصلاح پایانی تجربه storefront — #53/#54

- #53: سورس و مدرک اولیه [fec61ea](https://github.com/MohamadParsa/marketplace/commit/fec61eac457bdbac0901d43e2464f9b808ededa9)؛ [گزارش/تصاویر](storefront-followups/issue-53-verification.md).
- #54: سورس و مدرک اولیه [964824d](https://github.com/MohamadParsa/marketplace/commit/964824df0a40d7c5bdb3abcb7256613ccff0f982)؛ [گزارش/تصاویر](storefront-followups/issue-54-verification.md).

۳۹کنترل جدید و۲۳رگرسیون موفق و build/lint موفق؛ هر۵ job CI در PR #61 موفق بود و merge acac083ed0f291da7290a50c16332fb48950f289 روی main است. هر دو Issue با کامنت بسته‌اند؛ آخرین کد اجرایی 964824df0a40d7c5bdb3abcb7256613ccff0f982 است.

## برنامه۳۰درخواست جدید — فقط مستندات

[برنامه و Issueهای #62تا#91](next-phase/README.md) با وضعیت todo/open ثبت شد؛ هیچ کامیت سورس اجرایی جدیدی برای این برنامه وجود ندارد. این توصیف مربوط به زمان ثبت اولیه است؛ اجرای پنج انتخاب بعدی در بخش زیر و source-document-map.json ثبت شده است.

## انتخاب‌های #80/#83/#84/#85/#86

- #80: [گزارش و تصاویر](next-phase/issue-80-verification.md)؛ سورس مشترک [1d05e39](https://github.com/MohamadParsa/marketplace/commit/1d05e39f858cf87d1f13be5aefd7b973a78f05ad)، مدرک [a20304d](https://github.com/MohamadParsa/marketplace/commit/a20304d0f334cdfebde883752c6da4bc904c0c16).
- #83: [گزارش و تصاویر](next-phase/issue-83-verification.md)؛ سورس مشترک [1d05e39](https://github.com/MohamadParsa/marketplace/commit/1d05e39f858cf87d1f13be5aefd7b973a78f05ad)، مدرک [a20304d](https://github.com/MohamadParsa/marketplace/commit/a20304d0f334cdfebde883752c6da4bc904c0c16).
- #84: [گزارش و تصاویر](next-phase/issue-84-verification.md)؛ سورس مشترک [1d05e39](https://github.com/MohamadParsa/marketplace/commit/1d05e39f858cf87d1f13be5aefd7b973a78f05ad)، مدرک [a20304d](https://github.com/MohamadParsa/marketplace/commit/a20304d0f334cdfebde883752c6da4bc904c0c16).
- #85: [گزارش و تصاویر](next-phase/issue-85-verification.md)؛ سورس مشترک [1d05e39](https://github.com/MohamadParsa/marketplace/commit/1d05e39f858cf87d1f13be5aefd7b973a78f05ad)، مدرک [a20304d](https://github.com/MohamadParsa/marketplace/commit/a20304d0f334cdfebde883752c6da4bc904c0c16).
- #86: [گزارش و تصاویر](next-phase/issue-86-verification.md)؛ سورس مشترک [1d05e39](https://github.com/MohamadParsa/marketplace/commit/1d05e39f858cf87d1f13be5aefd7b973a78f05ad)، مدرک [a20304d](https://github.com/MohamadParsa/marketplace/commit/a20304d0f334cdfebde883752c6da4bc904c0c16).

اثر پنج اصلاح محلی بررسی شد؛ هر۵ job CI موفق، PR #92 ادغام و پنج Issue با کامنت شواهد بسته شدند. کامیت ادغام `f17fb49995647431b6dbe519aff347151d710068` است. #80 فقط ظاهر attribute را به checkbox برمی‌گرداند و وابستگی سخت #79 برای این برش ندارد؛ کارهای دیگر بدون مجوز اجرا نشده‌اند.

## سابقه آماده‌سازی محلی پس از دستور توقف پوش

روی پایه کامیت92f5e6d، تحلیل #66 و سپس سورس/مدارک [علاقه‌مندی #63](next-phase/issue-63-verification.md) در D:\Cherriva آماده‌اند؛ کامیت اجرایی یا مدرک تازه منتشر نشده و Issueها در گیت باز مانده‌اند. بخش local_working_tree در JSON این تفاوت را ثبت می‌کند؛ current_source_commit قبلی، آخرین سورس **کامیت‌شده** است و به معنی پاک‌بودن working tree کنونی نیست.

در ادامه همین working tree، [#64 و اصلاحات قلب/breadcrumb](next-phase/issue-64-verification.md) نیز done محلی است؛ base/HEAD همان92f5e6d و همه تغییرات #66/#63/#64 هنوز بدون commit/push هستند.

## شروع انتشار دسته مشتری — ۲۰۲۶-۱۱-۲۴

با مجوز مالک برای پوش تغییرات تا اینجا، سورس #63/#64 و اصلاحات UI در [e5babff](https://github.com/MohamadParsa/marketplace/commit/e5babff0f35242ffe94e7510e9833733dba1ded1) کامیت شد؛ تحلیل #66 کد اجرایی تازه ندارد. source-document-map.json بخش customer_lists_publication و ردیف‌های این سه Issue را ثبت می‌کند. انتشار سورس و مدارک در حال انجام است؛ محدودیت فقط‌لوکال قبلی تاریخچه این دسته است و برای کار بعدی پابرجاست.

نسخه ترکیبی منتشرشونده پس از حفظ main تا12bac82، [9e72faf](https://github.com/MohamadParsa/marketplace/commit/9e72faf5a1c9b5ce73e3628e539c8cd082f29b02) است؛ نسخه API آن همان کد بررسی‌شده است و frontend ترکیبی build/lint و۶۵کنترل مرورگر موفق دارد.

## انتشار نهایی دسته مشتری

تغییرات در [PR #93](https://github.com/MohamadParsa/marketplace/pull/93) پس از موفقیت هر۵ job CI، شامل Go race/vet، روی main با کامیت `a253d5edbe0adb467e1e7fb2ab5903608fb64bdc` ادغام شدند. این Issue با کامنت شواهد بسته شد. استقرار لیارا انجام نشد؛ نسخه D:\Cherriva نیز با main همگام است. تغییرهای بعد از این دسته تا دستور صریح تازه مالک لوکال می‌مانند.

نسخه سورس ترکیبی `9e72faf5a1c9b5ce73e3628e539c8cd082f29b02` و نسخه مدرک اولیه `fec70e82d1cd332d2673a470737baac1fae91df8` از نظر درخت apps/packages یکسان‌اند؛ PR #93 همه تغییرات تا این مرحله را روی main دارد. #66 فقط مدرک تحلیلی است و source_commit مستقل ندارد.

## چهار قابلیت محلی #67/#68/#69/#78 — ۲۰۲۶-۱۱-۲۴

[گزارش و فایل‌های سورس/تست](next-phase/issues-67-68-69-78-verification.md) روی پایه `52a8a09155fbcedae981eaf18a27ecea76120375` در working tree مسیر `D:\Cherriva` است؛ کامیت سورس یا سند تازه‌ای ساخته/پوش نشده و شماره کامیت برای این تحویل ادعا نمی‌شود. انتشار آینده باید کامیت واقعی را به این مدرک وصل کند.

### انتشار بسته چهار قابلیت

به دستور تازه مالک، سورس و اسناد در [کامیت 1dd7ccd](https://github.com/MohamadParsa/marketplace/commit/1dd7ccd616ba0d6380ce75adcdf9aaf38310a64d) روی شاخه `codex/checkout-coupons-credit-warranty` و [PR #94](https://github.com/MohamadParsa/marketplace/pull/94) پوش شدند؛ اصلاح اسناد انتشار همین بخش با `Source-Commit: 1dd7ccd616ba0d6380ce75adcdf9aaf38310a64d` به آن متصل است. main/لیارا هنوز ادغام/منتشر نشده‌اند. تغییر بعدی فعال‌کردن PDP از این پوش جدا و لوکال است.

## ادغام محلی پارسا و تفکیک از سورس منتشرشده

سه شاخهٔ نظر/پرسش/طرح با دستور مالک دریافت و تا `cffbf4f4125d73338d2af12972b08c41769901e3` محلی ادغام شدند؛ اصلاح دکمه و اسناد برنامه بعدی هنوز لوکال‌اند. [شواهد](backlog-review-2026-11-24/parsa-integration.md) · [تسک انتشار #105](https://github.com/MohamadParsa/marketplace/issues/105). این commit هنوز به origin پوش نشده و جای source منتشرشده `1dd7ccd` در PR #94 را نمی‌گیرد.

تحویل محلی #102/#103/#104 در سورس `ce417c5e42a1d3639332a5829e0572e14d1afb4a`؛ [گزارش و تصاویر و نتیجه ماشین‌خوان](backlog-review-2026-11-24/issues-102-103-104-verification.md). این commit پوش نشده و فعال‌سازی قبلی PDP همچنان تغییر جدا و کامیت‌نشده در لوکال است؛ مرجع نسخه پوش‌شده/main تغییر نکرد.


## کاربران واحد و نقش‌های مستقل — #111/#112

سورس، migration افزایشی۰۰۰۰۱۲، فرم‌ها و تست‌ها در کامیت محلی `37ec82c31408d69736a4f7826fa5ae07ec905077` در `D:\Cherriva` ثبت شدند. [گزارش و تصاویر](backoffice-refinement-2026-11-25/issues-111-112-verification.md) · [نتیجه ماشین‌خوان](backoffice-refinement-2026-11-25/issues-111-112-checks.json) · [رسید کامنت و بستن Issueها](backoffice-refinement-2026-11-25/issues-111-112-delivery.json).

هر دو Issue با کامنت شواهد completed بسته شدند؛۳۲۵تست Go و۳۰کنترل مرورگر موفق،۱۷بخش داده بعد از restart محفوظ. مدارک این تحویل با trailer `Source-Commit: 37ec82c31408d69736a4f7826fa5ae07ec905077` به سورس وصل‌اند و tree فایل‌های اجرایی همان است. این کامیت **طبق دستور مالک پوش نشده**؛ last-published source و main/لیارا تغییری نکرده‌اند. خواندن داخلی snapshot هویت هنوز کامل است؛ paging ایندکس‌دار SQL و ممیزی جامع #71 در این تحویل ادعا نشده‌اند.


## خصوصیات چریوا در پنل و ساده‌سازی آدرس — #110

سورس و تست‌ها در `2bc38ac30fa2b0c7e6df7fa30dcb23b2539b4950` در D:\Cherriva ثبت شدند؛ [گزارش و تصاویر](backoffice-refinement-2026-11-25/issue-110-verification.md) و [نتیجه ماشین‌خوان](backoffice-refinement-2026-11-25/issue-110-checks.json). دستور جدید مالک پوش پس از تکمیل را مجاز کرده و این نسخه همراه کامیت‌های قبلی محلی #100/#101/#73/#81/#111/#112 روی همان شاخه تحویل می‌شود؛ main و لیارا جدا هستند. کامیت مدارک trailer `Source-Commit: 2bc38ac30fa2b0c7e6df7fa30dcb23b2539b4950` دارد و درخت اجرایی آن با همین سورس برابر است.

پوش این دسته تأیید شد: [سورس](https://github.com/MohamadParsa/marketplace/commit/2bc38ac30fa2b0c7e6df7fa30dcb23b2539b4950) و [مدارک](https://github.com/MohamadParsa/marketplace/commit/c75b4b2183ffb010f840ef3b349f0f7eb81b1d0b) در [PR #106](https://github.com/MohamadParsa/marketplace/pull/106) موجودند؛ #110 با [کامنت پایان](https://github.com/MohamadParsa/marketplace/issues/110#issuecomment-7183781436) بسته شد. کامیت‌های قبلی محلی نیز در همان شاخه منتشر شدند؛ source اصلی و مدارک از نظر درخت اجرایی یکسان‌اند. main و لیارا مستقل باقی‌اند.
