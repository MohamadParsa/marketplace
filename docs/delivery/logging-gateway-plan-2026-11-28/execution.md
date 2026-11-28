---
id: DEL-LOGGING-EXECUTION-001
title: اجرای لاگ محلی و گزارش عملیات
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [DEL-LOGGING-GATEWAY-001, API-OPERATION-LOGGING-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# اجرای شش تسک لاگ؛ درگاه معوق

مالک اجرای #155/#156/#157 و #75/#76/#77 را انتخاب کرد؛ #158..163 درگاه همچنان معوق‌اند. [برنامه اولیه](README.md) شاهد تسک‌بندی پیشین است؛ این گزارش وضعیت اجرای انتخاب تازه را نگه می‌دارد.

## پاسخ‌های صریح مالک

- فنی/سرویس بیرونی۷روز، رفتار۳۰روز و مجموع فایل‌ها۵۰۰مگابایت؛ همه قابل تنظیم و مشاهده فقط ادمین.
- سابقه حسابرسی دیتابیس خودکار پاک نشود. شکست فایل/telemetry خرید را متوقف نکند و افت قابل مشاهده باشد؛ audit الزامی همراه تغییر اصلی اتمیک باشد.
- رویدادهای سروری مشاهده محصول، جست‌وجو، سبد، شروع خرید و پرداخت قطعی؛ مهمان فقط شناسه مستعار سبد و بدون fingerprint یا ردیاب بیرونی. گزارش ساده قیف فقط ادمین.

## مراحل و شواهد

| تسک | وضعیت اجرا | شاهد |
| --- | --- | --- |
| #155 | تکمیل و بسته؛ سورس1c1d19d و هر۵CI موفق | allowlist پیش از sink، context هم‌روند، adapter slog، خطای sink بدون نشت؛ race محلی نیازمند CGO است و در CI لینوکس اجرا می‌شود. |
| #156 | تکمیل و بسته؛ سورس316f2f8 و هر۵CI موفق | JSON Lines، صف محدود، قفل بین پردازه‌ای، rotation/retention/سقف مشترک و CLI فقط‌خواندنی |
| #157 | تکمیل و بسته؛ سورسcd183ee و هر۵CI موفق | runtime/lifecycle، HTTP/actor، commerce/preview/command و adapter پیامک/نشان |
| #75 | تکمیل و بسته؛ سورسbaaa312 و هر۵CI موفق | audit اتمیک/غیرقابل بازنویسی، diff مجاز و گزارش فقط ادمین |
| #76 | تکمیل و بسته؛ سورس7e7cb57 و هر۵CI موفق | رویدادهای سروری، dedupe و گزارش قیف فقط ادمین |
| #77 | تکمیل و بسته؛ سورس4f76041 و هر۵CI موفق | پذیرش یکپارچه، مسیر خصوصی با ancestor واقعی و راهنمای محلی |

تکمیل هر مرحله با commit/push، بررسی CI و کامنت شواهد همان Issue ثبت می‌شود. هیچ تماس زرین‌پال یا احراز پذیرنده در این دامنه ساخته نمی‌شود؛ callback زمان پرداخت/دامنه سایت به تسک معوق خودش تعلق دارد.

## شاهد گزارش و حریم #75

تست شکست commit و بازنویسی audit هیچ تغییر/رویداد موفق باقی نمی‌گذارد؛ PostgreSQL پس از بازگشایی metadata و موجودی درست را حفظ می‌کند. فقط ادمین می‌تواند audit/رویداد/وضعیت مقصد را بخواند؛ readback شامل reason آزاد، fingerprint یا credential نیست. ارقام/تاریخ فارسی و فیلترها در شروع بسته‌اند. [audit دسکتاپ](assets/audit-report-desktop.png) · [لاگ دسکتاپ](assets/operation-logs-desktop.png) · [موبایل](assets/operation-logs-mobile.png). تصاویر fixture هستند؛ اسکرول افقی جدول در ظرف خودش است.

## شاهد رفتار مشتری #76

تست کامل API با PostgreSQL ایزوله/Redis15 و vet، build هر دو رابط و مرورگر با داده ساختگی موفق‌اند. replay سبد یک رخداد؛ ادغام دوباره افزودن نیست؛ شارژ کیف پول paid نیست؛ تأیید خرید یک paid با actor واقعی و subject مشتری دارد. origin دستگاه/عملیات quote پس از تأیید ادمین حفظ می‌شود. مهمان فقط alias سبد موجود؛ متن جست‌وجو، token و UA خام حذف می‌شوند. دستگاه/دسته/محیط و دسترسی ادمین در API/مرورگر بررسی شد. [دسکتاپ](assets/activity-funnel-desktop.png) · [موبایل](assets/activity-funnel-mobile.png). این شمارش مرحله‌هاست؛ ادعای نرخ تبدیل cohort یا تضمین بدون افت telemetry ندارد.

## پذیرش نهایی #77

تست کامل Go با PostgreSQL ایزوله و Redis15 و go vet موفق؛ lint/build بک‌آفیس موفق. رد checkout از HTTP تا domain با operation/request مشترک و actor معتبر دنبال شد؛ نه موجودی/دفتر مالی تغییر کرد و نه body/query/کوپن خصوصی ثبت شد. قرارگیری لاگ زیر ancestor از نوع symlink/junction در web-root نیز رد می‌شود. نمایش نام عامل فقط در DTO گزارش مجاز انجام می‌شود و وارد فایل لاگ نمی‌شود. CLI بازه ابتدا/انتها، شدت و نام رخداد را دارد. [راهنمای نگهداری/خطا/جست‌وجو](local-logging-guide.md) محدودیت best-effort و تفاوت audit/ledger را توضیح می‌دهد.

تست‌های فایل: rotation/retention/cap/restart، دو نویسنده، صف پر و مسیر غیرقابل نوشتن. تست‌های audit: rollback، منع بازنویسی، PostgreSQL restart و مجوز ادمین. integration پیامک/نشان با mock، شامل HTTP200 با رد معنایی و timeout؛ سرویس واقعی فراخوانی نشده است. BDDهای بسته، سناریوی مستند هستند؛ runner BDD اجرا نشده و ادعای آن نداریم.


اجرای اصلی محلی8098/5197 با API و رابط تازه آماده و health/صفحه گزارش/خانه۲۰۰بودند. پیش/پس راه‌اندازی:۳۵۳۴محصول،۴۸فروشگاه،۲۲سفارش و۱۵خرید ثابت؛۱۴ورودی ledger با fingerprint یکسان `4eef5db469592d7c060f730f49e8ce7e5fdf417c01667707ac12f6dd9768862b`. هیچ درخواست/استرداد جدید ایجاد نشد. CLI فقط‌خواندنی رویداد فنی با operation معتبر را از مقصد اصلی خواند. مرورگر هر سه تب با fixture دوباره موفق بود؛ actor در audit با نام مجاز نمایش داده می‌شود.


## بستن و پوش نهایی

هر شش Issue غیر درگاه با کامنت شواهد بسته شدند: [#155](https://github.com/MohamadParsa/marketplace/issues/155#issuecomment-7239503828)، [#156](https://github.com/MohamadParsa/marketplace/issues/156#issuecomment-7239616618)، [#157](https://github.com/MohamadParsa/marketplace/issues/157#issuecomment-7239784469)، [#75](https://github.com/MohamadParsa/marketplace/issues/75#issuecomment-7239916708)، [#76](https://github.com/MohamadParsa/marketplace/issues/76#issuecomment-7240111001) و [#77](https://github.com/MohamadParsa/marketplace/issues/77#issuecomment-7240245858). [CI سورس نهایی](https://github.com/MohamadParsa/marketplace/actions/runs/46151516858) شامل هر۵بررسی موفق است؛ guard symlink در Linux نیز همراه race پاس شد. بررسی لینک‌های نسبی و یکتایی شناسه‌های اسناد تغییرکرده بدون خطا بود. شش Issue #158..163 همچنان open/deferred و PR #106 و والد #95 بازند؛ هیچ merge/deploy انجام نشد.
