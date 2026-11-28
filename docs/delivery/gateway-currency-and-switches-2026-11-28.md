---
id: FEAT-BO-GATEWAY-CURRENCY-001
title: اصلاح سوییچ‌های درگاه و ضریب تبدیل مبلغ
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [FEAT-BO-GATEWAY-SETTINGS-REVISION-001, DOM-PAYMENTS-001]
approved_by: []
implementation_status: done
execution_scope: settings-only
---

# اصلاح سوییچ و ضریب تبدیل — #154

[Issue #154](https://github.com/MohamadParsa/marketplace/issues/154) ادامه بازبینی [تنظیمات #153](payment-settings-revision-2026-11-28.md) است. مالک، هم‌پوشانی سوییچ با دکمه و پراکندگی دو گزینه را گزارش کرد و واحد enum و ضریب تبدیل قابل ویرایش خواست. توضیح اولیه «هر ۱ تومان چند واحد درگاه» با تصریح مالک به **ضریب ضرب مبلغ** اصلاح شد.

## مسئله و نتیجه

فعال بودن درگاه و امکان شارژ کیف پول، دو سوییچ کنار هم در کارت و مودال‌اند؛ دکمه ویرایش جداست. واحد سایت تومان (`IRT`) و واحد قابل انتخاب درگاه تومان/ریال (`IRT`/`IRR`) است. مبلغ سایت به تومان × ضریب = مبلغ ارسالی؛ برای تومان به ریال ضریب معمول۱۰و برای تومان به تومان۱است. عنوان فیلد با انتخاب واحد به «ضریب تبدیل تومان به ریال/تومان» تغییر می‌کند.

## دامنه و قواعد

فقط تنظیمات بک‌آفیس با مجوزهای موجود خواندن/نوشتن پرداخت؛ حریم و اعتبارنامه‌ها طبق [نقش‌ها](../product/roles.md) و [حوزه پرداخت، R28/R29](../domains/payments/README.md#gateway-settings-20261128). داده دفتر مالی طبق [ADR-001](../decisions/adr-001-pricing-and-payment-boundaries.md) و [دامنه محصول](../product/scope.md) ریال صحیح می‌ماند؛ enum پیش‌فرض سایت، واحد ورود/نمایش است و تغییر تنظیم درگاه داده گذشته را تبدیل نمی‌کند.

اتصال واقعی، ارسال OTP، OAuth و توکن، callback handler و request/verify در این درخواست اجرا نمی‌شوند. راه‌اندازی حساب پذیرنده پیشنهاداً در همین تنظیمات و پیش از فعال‌سازی، با عملیات ادمین انجام خواهد شد؛ منتظر اولین پرداخت مشتری نمی‌ماند. فعلاً فرم صریحاً می‌گوید فقط مشخصات ذخیره می‌شوند و کنترل عملیاتی راه‌اندازی هنوز اضافه نشده است.

## مسیر اصلی

| گام | اقدام | نتیجه |
| --- | --- | --- |
| ۱ | بازکردن پرداخت و مقصد وجه | همه بخش‌ها بسته در شروع |
| ۲ | بازکردن زرین‌پال | دو سوییچ کنار هم و دکمه مستقل |
| ۳ | تغییر هر سوییچ | ذخیره فوری و حفظ پس از refresh؛ اجرای بانکی هنوز آماده نیست |
| ۴ | ویرایش ← اتصال ← تنظیمات پرداخت | واحد enum و ضریب قابل ویرایش |
| ۵ | انتخاب ریال | ضریب پیشنهادی۱۰؛ انتخاب تومان ضریب۱؛ قابل تغییر پیش از ذخیره |
| ۶ | ذخیره | اعتبارسنجی سروری و خواندن بدون build؛ اطلاعات خالی credential قبلی را حفظ می‌کند |

## ورودی، خطا و تبدیل

[enum مشترک و تبدیل دقیق](../../apps/api/internal/domain/commerce/currency.go) فقط IRT/IRR را می‌پذیرد. ضریب رشته اعشاری مثبت با حداکثر۶رقم اعشار و سقف فنی۱۰۰۰۰۰۰است؛ float در محاسبات مالی استفاده نمی‌شود. صفر/منفی/نمای علمی/کسر متنی/واحد ناشناخته رد می‌شوند و تنظیمات قبلی حفظ می‌شوند. ارقام فارسی پذیرفته و ضریب نرمال ذخیره می‌شود. تنظیم قدیمی IRR بدون ضریب با۱۰خوانده می‌شود و خواندن هیچ write ندارد؛ پیش‌فرض جدید IRT/1است.

سازنده مبلغ برای adapter آینده، `Amount` ریالی را به مبلغ تومان و سپس با ضرب ضریب به عدد صحیح درگاه می‌رساند؛ اعشار غیرقابل ارسال و overflow خطا هستند، نه گردکردن پنهان. مثلاً۱۰۰۰۰تومان ×۱۰=۱۰۰۰۰۰ریال. استفاده عملیاتی و snapshot مبلغ/واحد/ضریب هر تلاش، در مرحله اتصال لازم است؛ تغییر تنظیم نباید verify تلاش قبلی را با ضریب جدید انجام دهد.

نداشتن مجوز: سوییچ‌ها غیرفعال و نوشتن سروری ممنوع؛ قطع ارتباط: خطا و امکان تلاش مجدد با idempotency؛ نسخه قدیمی: conflict و نیاز به بازخوانی؛ داده خالی: provider پیش‌فرض بدون ایجاد رکورد؛ موفقیت: ذخیره و پیام؛ بخش خطادار در مودال باز می‌شود.

## چرا دو آدرس و راه‌اندازی جدا؟

[مستند احراز هویت](https://www.zarinpal.com/docs/apiDocs/auth) برای حساب پذیرنده از `https://next.zarinpal.com/api` و مسیرهای `oauth/initialize` و `oauth/token` استفاده می‌کند؛ کلاینت/حساب و کد اولیه برای API مدیریت حساب است. Access/Refresh Token خروجی ورودند؛ تازه‌سازی با سرور خواهد بود. این موضوع با پرداخت مشتری متفاوت است.

[درخواست/تأیید پرداخت](https://www.zarinpal.com/docs/paymentGateway/connectToGateway) از `https://payment.zarinpal.com/pg/v4/payment` و مسیرهای `request.json` و `verify.json` استفاده می‌کند؛ محیط آزمایشی پرداخت آدرس جدا دارد. Sandbox پرداخت به معنای Sandbox OAuth نیست. Client ID/Secret جای الزام مستند Merchant برای عملیات مستقیم V4 را نمی‌گیرند.

## callback و verify

طبق [راهنمای اتصال](https://www.zarinpal.com/docs/paymentGateway/connectToGateway)، Status=OK در callback به‌تنهایی نتیجه قطعی نیست. سرور verify می‌کند؛ کد۱۰۰موفقیت و۱۰۱قبلاً verifyشده است و ثبت مالی دوباره ایجاد نمی‌کند. تطبیق authority/مبلغ ثبت‌شده و جلوگیری از اثر تکراری لازم‌اند.

[auto_verify](https://www.zarinpal.com/docs/paymentGateway/moreFeatures/session-validation) فقط تعیین می‌کند آغاز تأیید نزد بانک با زرین‌پال باشد یا با فراخوانی verify چریوا. پیش‌فرض false همین مسیر callback→verify سرور است. در true بانک ممکن است زودتر خودکار تأیید شود؛ سایت همچنان نتیجه معتبر را سروری بررسی می‌کند. هیچ‌کدام احراز هویت API نیستند و مشتری مرحله ورود پذیرنده ندارد.

## پذیرش

- FEAT-BO-GATEWAY-CURRENCY-001-AC01: دو سوییچ کنار هم در کارت/مودال در۱۴۴۰و۳۹۰پیکسل، بدون تداخل با دکمه و سرریز و حفظ وضعیت پس از refresh.
- FEAT-BO-GATEWAY-CURRENCY-001-AC02: enum IRT/IRR با پیش‌فرض سایت IRT؛ ضریب۱۰برای ریال و۱برای تومان؛ ذخیره ضریب سفارشی دقیق و حفظ legacy ریالی بدون تغییر ledger.
- FEAT-BO-GATEWAY-CURRENCY-001-AC03: واحد/ضریب نامعتبر بدون تغییر داده رد شود؛ تبدیل صحیح، اعشار غیرقابل ارسال و overflow کنترل و credential/مجوز/نسخه حفظ شوند.
- FEAT-BO-GATEWAY-CURRENCY-001-AC04: توضیح روشن setup پیش از فعال‌سازی، دو base URL و callback/verify؛ همه بخش‌ها بسته شروع شوند و ذخیره تنظیم هیچ عملیات بانکی اجرا نکند.

[BDD طراحی](../bdd/features/experiences/gateway-currency-and-switches.feature) runner مستقل اجراشده نیست. شواهد زیر فقط تنظیمات و fixture ساختگی را پوشش می‌دهند؛ پرداخت واقعی آزمایش نشده است.
## شواهد اجرا

- `go test ./...` با PostgreSQL ایزوله و Redis DB15 و `go vet ./...` موفق؛ آزمون واحد، API، حفظ legacy ریالی و ذخیره/بازگشایی PostgreSQL ضریب۱۰٫۲۵، credential و مرز مالی را پوشش دادند.
- TypeScript/build بک‌آفیس موفق؛ آزمون مرورگر در۱۴۴۰و۳۹۰پیکسل دو سوییچ کنار هم، نبود تداخل دکمه، نبود سرریز، حفظ پس از refresh، واحد پیش‌فرض IRT/1 و ریال/10، ذخیره ضریب سفارشی و رد صفر را تأیید کرد. تنها خطای console، HTTP400عمدی آزمون ضریب صفر بود.
- تصاویر: [کارت دسکتاپ](gateway-currency-and-switches-assets/gateway-switches-desktop.png)، [کارت موبایل](gateway-currency-and-switches-assets/gateway-switches-mobile.png)، [مودال دسکتاپ](gateway-currency-and-switches-assets/gateway-switches-modal-desktop.png)، [مودال موبایل](gateway-currency-and-switches-assets/gateway-switches-modal-mobile.png)، [ضریب دسکتاپ](gateway-currency-and-switches-assets/gateway-currency-help-desktop.png)، [ضریب موبایل](gateway-currency-and-switches-assets/gateway-currency-help-mobile.png). داده‌ها ساختگی‌اند و مقدار۱۰در تصویر پیش‌نویس معمول ضریب ریال است.
- اجرای اصلی محلی با API جدید به‌روز شد؛ fingerprint دفتر مالی `4eef5db469592d7c060f730f49e8ce7e5fdf417c01667707ac12f6dd9768862b`،۱۴ورودی مالی،۱۵خرید،۲۲سفارش،۳۵۳۴محصول و۴۸فروشگاه پیش/پس یکسان‌اند. تنظیمات واقعی کاربر توسط agent ذخیره نشدند. health/ready و فایل‌های build نهایی در۸۰۹۸/۵۱۹۷کد۲۰۰دارند.
- [سورس38315de](https://github.com/MohamadParsa/marketplace/commit/38315de2c1545a90fd0e56894965ee0c6e60b6bb) پوش و [هر پنج بررسی CI](https://github.com/MohamadParsa/marketplace/actions/runs/46142128190) موفق شد؛ #154 با [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/154#issuecomment-7238895886) بسته و والد #95 به‌روز شد. مرورگر و fixture بسته شدند؛ دو اجرای اصلی محلی روشن‌اند. والد #95 و PR #106 باز می‌مانند. merge/deploy عمومی و اجرای درگاه در این تحویل نیست.
کنترل ۹ سند Markdown، ۵۴۴پیوند نسبی، شناسه یکتا، چهار معیار پذیرش/چهار سناریوی BDD و git diff --check موفق بود.
