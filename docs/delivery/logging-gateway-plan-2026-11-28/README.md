---
id: DEL-LOGGING-GATEWAY-001
title: برنامه لاگ محلی و کارهای معوق درگاه
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [API-OPERATION-LOGGING-001, FEAT-NEXT-014, FEAT-NEXT-015, FEAT-NEXT-016]
approved_by: []
implementation_status: logging-complete-gateway-deferred
execution_authorized: true
planning_status: done
---

# اول لاگ‌گیری؛ سپس درگاه

پس از تسک‌بندی، مالک اجرای همه موارد غیر درگاه را انتخاب کرد و NEXT-Q09/Q10 را پاسخ داد. #155/#156/#157 و #75/#76/#77 تکمیل، پوش و با کامنت شواهد بسته‌اند؛ هر۵CI سورس نهایی موفق است. [گزارش اجرا](execution.md)، [راهنمای محلی](local-logging-guide.md) و [ماتریس پوشش](coverage.md) شاهد وضعیت جاری‌اند. همه تسک‌ها مسئول `eliassarband` دارند؛ #158..163 درگاه با برچسب `status:deferred` همچنان معوق‌اند.

## خواسته صریح و تصمیم‌ها

| موضوع | قاعده |
| --- | --- |
| callback | هنگام ساخت هر درخواست پرداخت، با دامنه اصلی storefront همان محیط و مسیر ثابت کد تولید شود؛ دامنه بک‌آفیس/API یا Host ورودی مشتری مبنا نباشد. |
| تقدم | احراز هویت و اجرای زرین‌پال فعلاً عقب بیفتد؛ زیرساخت و پوشش لاگ پیش‌نیاز شروع دوباره‌اند. |
| پوشش | اکشن‌های مهم سایت، نمونه افزودن سبد و مراحل خرید، نقاط تماس سرویس‌های بیرونی و تمام اکشن‌های درگاه ثبت شوند. |
| مقصد | فعلاً محلی؛ امکان تغییر sink در آینده حفظ شود. ClickHouse و Elastic انتخاب فناوری یا سرویس آماده این مرحله نیستند. |
| تفکیک | نوع رویداد از شدت/نتیجه جدا و اطلاعات لازم برای دنبال‌کردن یک عملیات موجود باشد. قرارداد فیلدها و JSON Lines طراحی پیشنهادی‌اند. |

مراجع: [قرارداد مشترک پیشنهادی لاگ](../../architecture/contracts/operation-logging.md)، [کیفیت و حریم](../../architecture/quality.md)، [نقش‌ها](../../product/roles.md)، [قواعد پرداخت](../../domains/payments/README.md) و [دفتر پرسش‌ها](../../product/open-questions.md).

## وضعیت واقعی سورس

[logger ورودی API](../../../apps/api/cmd/api/main.go) و [bootstrap](../../../apps/api/internal/bootstrap/server.go) اکنون از recorder مشترک پالایش‌شده استفاده می‌کنند؛ مقصد فایل خصوصی، correlation و گزارش ادمین اجرا شده‌اند. مرز هر پوشش و محدودیت در [ماتریس](coverage.md) مشخص است.

[سازنده callback](../../../apps/api/internal/application/backoffice/gateway_callback_config.go) مسیر ثابت را با origin محیطی می‌سازد، ولی request/handler بانکی هنوز نداریم. تسک #158 استفاده در زمان ایجاد تلاش با origin سایت و snapshot را تکمیل می‌کند؛ تنظیم دوباره همان قابلیت تحویل‌شده #152..154 نیست.

## مالکیت یکتا؛ سه تسک موجود

| Issue | دامنه حفظ‌شده | تکمیل این بازبینی |
| --- | --- | --- |
| [#77](https://github.com/MohamadParsa/marketplace/issues/77) | لاگ فنی عملیات حساس | مالک پوشش فنی؛ زیرتسک‌های #155..157 اجراهای قرارداد، مقصد و instrumentation هستند؛ اجرای تکراری مستقل در والد ساخته نمی‌شود. |
| [#75](https://github.com/MohamadParsa/marketplace/issues/75) | ActivityLog کسب‌وکار | تکمیل/اتصال audit موجود به قرارداد و correlation با diff مجاز، عامل/مجوز و سابقه حساس؛ ledger یا telemetry فایل جای آن را نمی‌گیرد. |
| [#76](https://github.com/MohamadParsa/marketplace/issues/76) | رفتار مشتری و Analytics | افزودن/ویرایش/حذف سبد و مسیر خرید با معنای سروری و dedupe؛ بخش نیاز فعلی از گزارش/ردیابی فراگیرِ هنوز نیازمند تصمیم جداست. |

## تسک‌های تازه

| Issue | کد و خروجی | وابستگی | وضعیت |
| --- | --- | --- | --- |
| [#155](https://github.com/MohamadParsa/marketplace/issues/155) | [CV-LOG-01 — قرارداد مشترک لاگ، شناسه پیگیری و حذف داده حساس](tasks/logging-contract-correlation-redaction.md) | قواعد موجود | تکمیل و بسته |
| [#156](https://github.com/MohamadParsa/marketplace/issues/156) | [CV-LOG-02 — ذخیره محلی لاگ، چرخش فایل و امکان تغییر مقصد](tasks/local-log-storage-and-sink-boundary.md) | #155 | تکمیل و بسته |
| [#157](https://github.com/MohamadParsa/marketplace/issues/157) | [CV-LOG-03 — پوشش مسیر خرید، عملیات حساس و تماس سرویس‌های بیرونی](tasks/instrument-critical-actions-and-external-services.md) | #155، #156 | تکمیل و بسته |
| [#158](https://github.com/MohamadParsa/marketplace/issues/158) | [CV-PAY-GW-04 — ساخت callback هنگام پرداخت با دامنه اصلی سایت](tasks/storefront-domain-callback-at-payment-time.md) | #157 | معوق |
| [#159](https://github.com/MohamadParsa/marketplace/issues/159) | [CV-PAY-GW-05 — احراز حساب پذیرنده و چرخه فعال‌سازی زرین‌پال](tasks/merchant-authentication-and-activation.md) | #157 | معوق |
| [#160](https://github.com/MohamadParsa/marketplace/issues/160) | [CV-PAY-GW-06 — اتصال request و verify زرین‌پال با تلاش نسخه‌دار](tasks/zarinpal-request-verify-adapter.md) | #157، #158 | معوق |
| [#161](https://github.com/MohamadParsa/marketplace/issues/161) | [CV-PAY-GW-07 — هماهنگ‌سازی پرداخت درگاه با کیف پول و سفارش مشتری](tasks/verified-funding-and-order-payment.md) | #160 | معوق |
| [#162](https://github.com/MohamadParsa/marketplace/issues/162) | [CV-PAY-GW-08 — بازیابی نتیجه نامشخص و retry امن پرداخت](tasks/gateway-recovery-inquiry-and-safe-retries.md) | #160، #161 | معوق |
| [#163](https://github.com/MohamadParsa/marketplace/issues/163) | [CV-PAY-GW-09 — پذیرش یکپارچه زرین‌پال و راهنمای فعال‌سازی](tasks/gateway-end-to-end-and-activation-review.md) | #159، #161، #162 | معوق |

## ترتیب پیشنهادی اجرا پس از انتخاب مالک

1. #155 قرارداد/correlation/پالایش → #156 ذخیره محلی/rotation/مقصد قابل تغییر.
2. #157 تماس‌های بیرونی و نقاط حساس موجود؛ هم‌زمان بخش‌های مستقل #75/#76 روی قرارداد مشترک، سپس شاهد پیگیری خرید و تمام مسیرهای بیرونی اجراشده.
3. پس از آماده‌شدن پوشش لاگ و انتخاب مجدد اجرای درگاه: #158 callback و #159 راه‌اندازی پذیرنده؛ #160 پرداخت V4 به OAuth مدیریتیِ غیرضروری وابسته نمی‌شود.
4. #160 request/verify → #161 credit/debit/نتیجه مشتری با تعیین GATEWAY-Q01 → #162 بازیابی/retry → #163 پذیرش/راهنمای فعال‌سازی.

این ترتیب برای لاگ با درخواست بعدی مالک اجرا شد؛ مجوز شروع درگاه، ادغام یا انتشار نیست. پیاده‌سازی تنظیمات #152/#153/#154 بسته باقی می‌ماند؛ شش Issue تازه صرفاً باقیمانده اجرا را پوشش می‌دهند و موارد درگاه از روی این برنامه done/closed نمی‌شوند.

## معیار آمادگی و تصمیم‌های باز

[NEXT-Q09/Q10](../../product/open-questions.md#next-phase-questions) برای برش محلی با پاسخ صریح مالک حل شدند:۷/۳۰روز،۵۰۰MiB، فقط ادمین، telemetry بدون توقف خرید و audit اتمیک؛ مهمان فقط alias سبد موجود و قیف ساده سروری. ردیابی بیرونی و Analytics جامع انتخاب نشده‌اند. برای درگاه، GATEWAY-Q01 خودکار/دستی‌بودن مصرف پس از شارژ، Q02 شرایط پذیرنده واقعی و Q03 مهلت/الگوی بازیابی هنوز مرجع تصمیم‌اند. چرخه فعال‌سازی با سوییچ و OTP ادمین طراحی پیشنهادی مرحله بعد است، نه قابلیت اجراشده.

قسمت وابسته به ابهام مالی/حریم قبل از کدنویسی همان بخش روشن شود؛ روی بخش مستقل روشن می‌توان پس از انتخاب اجرا پیش رفت. داده ساختگی و [readiness](../readiness.md) برای هر برش لازم است.

## سابقه پذیرش و راستی‌آزمایی برنامه

هر تسک چهار معیار پذیرش و شاهد تست آینده دارد؛ [۹سناریوی BDD طراحی](../../bdd/features/experiences/logging-gateway-plan.feature) به۳۶معیار تازه متصل‌اند و runner اجرا نشده است. تست برنامه‌ریزی: لینک‌های نسبی/شناسه‌ها، نگاشت Issue واقعی، وضعیت باز/معوق، اساین، وابستگی و نبود تغییر executable. کنترل۲۶سند Markdown و۸۰۵پیوند نسبی، یکتایی۱۱شناسه تازه، تطبیق۳۶معیار با۹سناریوی طراحی و نبود تغییر executable موفق بود. وضعیت هر۹Issue باز، اساین مالک، وابستگی‌ها و معوق‌بودن شش مورد درگاه از GitHub بازخوانی و تأیید شد. نتیجه Git در کامنت پیگیری والد #95 ثبت می‌شود؛ تست محصول/BDD runner در این تغییر مستندی اجرا نشده است.


پذیرش اجرای محصول جداگانه در [execution.md](execution.md) ثبت شده است؛ اعداد کنترل لینک بالا مربوط به کامیت برنامه‌ریزی اولیه‌اند.
